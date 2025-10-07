package service

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"time"

	"wechat-robot-client/dto"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/robot"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"

	"github.com/h2non/filetype"
)

type MomentsService struct {
	ctx                context.Context
	momentRepo         *repository.Moment
	momentSettingsRepo *repository.MomentSettings
	momentCommentRepo  *repository.MomentComment
}

func NewMomentsService(ctx context.Context) *MomentsService {
	return &MomentsService{
		ctx:                ctx,
		momentRepo:         repository.NewMomentRepo(ctx, vars.DB),
		momentSettingsRepo: repository.NewMomentSettingsRepo(ctx, vars.DB),
		momentCommentRepo:  repository.NewMomentCommentRepo(ctx, vars.DB),
	}
}

func (s *MomentsService) SyncMomentStart() {
	ctx := context.Background()
	vars.RobotRuntime.SyncMomentContext, vars.RobotRuntime.SyncMomentCancel = context.WithCancel(ctx)
	for {
		select {
		case <-vars.RobotRuntime.SyncMomentContext.Done():
			return
		case <-time.After(10 * time.Minute):
			log.Println("开始同步朋友圈~")
			if vars.RobotRuntime.Status == model.RobotStatusOffline {
				continue
			}
			s.SyncMoments()
		}
	}
}

func (s *MomentsService) SyncMoments() {
	loginData, err := vars.RobotRuntime.GetCachedInfo()
	if err != nil {
		log.Println("同步朋友圈获取用户信息失败: ", err)
		return
	}

	momentSettings, err := s.momentSettingsRepo.GetMomentSettings()
	if err != nil {
		log.Println("获取朋友圈设置失败: ", err)
		return
	}
	if momentSettings == nil {
		f := false
		err = s.momentSettingsRepo.Create(&model.MomentSettings{
			SyncKey:             loginData.SyncKey,
			AutoLike:            &f,
			AutoComment:         &f,
			Whitelist:           nil,
			Blacklist:           nil,
			AIBaseURL:           "",
			AIAPIKey:            "",
			WorkflowModel:       "",
			CommentModel:        "",
			CommentPrompt:       "",
			MaxCompletionTokens: nil,
		})
		if err != nil {
			log.Println("创建朋友圈设置失败: ", err)
			return
		}
		momentSettings, err = s.momentSettingsRepo.GetMomentSettings()
		if err != nil {
			log.Println("获取朋友圈设置失败2: ", err)
			return
		}
	}
	if momentSettings.SyncKey == "" {
		momentSettings.SyncKey = loginData.SyncKey
		err = s.momentSettingsRepo.Update(momentSettings)
		if err != nil {
			log.Println("更新朋友圈设置失败: ", err)
			return
		}
		momentSettings, err = s.momentSettingsRepo.GetMomentSettings()
		if err != nil {
			log.Println("获取朋友圈设置失败3: ", err)
			return
		}
	}

	// 获取新朋友圈
	syncResp, err := vars.RobotRuntime.FriendCircleMmSnsSync(momentSettings.SyncKey)
	if err != nil {
		log.Println("获取新朋友圈失败: ", err)
		return
	}
	if syncResp.KeyBuf.Buffer == momentSettings.SyncKey {
		return
	}

	defer func() {
		momentSettings.SyncKey = syncResp.KeyBuf.Buffer
		err = s.momentSettingsRepo.Update(momentSettings)
		if err != nil {
			log.Println("defer 更新朋友圈设置失败: ", err)
			return
		}
	}()

	if len(syncResp.AddSnsBuffer) == 0 {
		// 没有新的朋友圈，直接返回
		return
	}

	now := time.Now().Unix()
	aiMomentService := NewAIMomentService(s.ctx)

	for _, momentMsg := range syncResp.AddSnsBuffer {
		var timelineObject robot.TimelineObject
		if err := xml.Unmarshal([]byte(momentMsg), &timelineObject); err != nil {
			log.Println("反序列化朋友圈消息失败: ", err)
			continue
		}
		moment := &model.Moment{
			WechatID:  timelineObject.Username,
			MomentID:  timelineObject.ID,
			Content:   momentMsg,
			CreatedAt: now,
			UpdatedAt: now,
		}
		err := s.momentRepo.Create(moment)
		if err != nil {
			log.Println("创建朋友圈失败: ", err)
			continue
		}
		if momentSettings.Whitelist != nil && *momentSettings.Whitelist != "" {
			if !strings.Contains(*momentSettings.Whitelist, timelineObject.Username) {
				log.Printf("用户[%s]不在白名单中，跳过处理", timelineObject.Username)
				continue
			}
		} else if momentSettings.Blacklist != nil && *momentSettings.Blacklist != "" {
			if strings.Contains(*momentSettings.Blacklist, timelineObject.Username) {
				log.Printf("用户[%s]在黑名单中，跳过处理", timelineObject.Username)
				continue
			}
		}
		if timelineObject.ContentDesc == "" && len(timelineObject.ContentObject.MediaList.Media) < 3 {
			log.Println("朋友圈文字内容为空，且图片少于三张，跳过处理")
			continue
		}
		if timelineObject.ContentDesc == "" {
			if momentSettings.AutoLike != nil && *momentSettings.AutoLike {
				_, err := vars.RobotRuntime.FriendCircleComment(robot.FriendCircleCommentRequest{
					Id:   strconv.FormatUint(timelineObject.ID, 10),
					Type: 1,
				})
				if err != nil {
					log.Println("自动点赞朋友圈失败: ", err)
					continue
				}
				log.Println("只有图片没有文字，只点赞不评论")
				continue
			}
		}
		// 自动评论
		var momentMood *MomentMood
		if momentSettings.AutoComment != nil && *momentSettings.AutoComment {
			// 判断下这个好友今天的朋友圈是否已经评论过了，每人每天只能被评论一次
			commented, err := s.momentCommentRepo.IsTodayHasCommented(timelineObject.Username)
			if err != nil {
				log.Println("获取朋友圈评论状态失败: ", err)
				continue
			}
			if commented {
				// 已经评论过了，只自动点赞
				if momentSettings.AutoLike != nil && *momentSettings.AutoLike {
					_, err := vars.RobotRuntime.FriendCircleComment(robot.FriendCircleCommentRequest{
						Id:   strconv.FormatUint(timelineObject.ID, 10),
						Type: 1,
					})
					if err != nil {
						log.Println("自动点赞朋友圈失败: ", err)
						continue
					}
					log.Println("今天已经评论过了，只点赞")
				} else {
					log.Println("今天已经评论过了，跳过")
				}
				continue
			}
			momentMood = aiMomentService.GetMomentMood(timelineObject.ContentDesc, *momentSettings)
			if momentMood == nil {
				log.Println("获取朋友圈心情失败")
				continue
			}
			if momentMood.Comment == "no" {
				log.Printf("%s: 朋友圈不适合评论，跳过", timelineObject.ContentDesc)
				continue
			}
			commentContent, err := aiMomentService.Comment(timelineObject.ContentDesc, *momentSettings)
			if err != nil {
				log.Println("获取朋友圈评论内容失败: ", err)
				continue
			}
			if commentContent.Content == "" {
				log.Println("获取朋友圈评论内容为空，跳过")
				continue
			}
			// 先保存一下，利用数据库唯一索引约束，保证同一条朋友圈不会重复评论
			newComment := model.MomentComment{
				WechatID:  timelineObject.Username,
				MomentID:  timelineObject.ID,
				Comment:   commentContent.Content,
				CreatedAt: now,
				UpdatedAt: now,
			}
			err = s.momentCommentRepo.Create(&newComment)
			if err != nil {
				log.Println("保存朋友圈评论记录失败: ", err)
				continue
			}
			_, err = vars.RobotRuntime.FriendCircleComment(robot.FriendCircleCommentRequest{
				Id:      strconv.FormatUint(timelineObject.ID, 10),
				Type:    2,
				Content: commentContent.Content,
			})
			if err != nil {
				log.Println("自动评论朋友圈失败: ", err)
				_ = s.momentCommentRepo.Delete(&newComment)
				continue
			}
		}
		// 自动点赞
		if momentSettings.AutoLike != nil && *momentSettings.AutoLike {
			if momentMood == nil {
				momentMood = aiMomentService.GetMomentMood(timelineObject.ContentDesc, *momentSettings)
				if momentMood == nil {
					log.Println("获取朋友圈心情失败")
					continue
				}
			}
			if momentMood.Like == "no" {
				log.Printf("%s: 朋友圈不适合点赞，跳过", timelineObject.ContentDesc)
				continue
			}
			_, err := vars.RobotRuntime.FriendCircleComment(robot.FriendCircleCommentRequest{
				Id:   strconv.FormatUint(timelineObject.ID, 10),
				Type: 1,
			})
			if err != nil {
				log.Println("自动点赞朋友圈失败: ", err)
				continue
			}
		}
	}
}

func (s *MomentsService) FriendCircleComment(req dto.FriendCircleCommentRequest) (robot.SnsCommentResponse, error) {
	return vars.RobotRuntime.FriendCircleComment(robot.FriendCircleCommentRequest{
		Type:           req.Type,
		Id:             req.Id,
		ReplyCommnetId: req.ReplyCommnetId,
		Content:        req.Content,
	})
}

func (s *MomentsService) FriendCircleGetDetail(req dto.FriendCircleGetDetailRequest) (robot.SnsUserPageResponse, error) {
	return vars.RobotRuntime.FriendCircleGetDetail(robot.FriendCircleGetDetailRequest{
		Towxid:       req.Towxid,
		Fristpagemd5: req.Fristpagemd5,
		Maxid:        req.Maxid,
	})
}

func (s *MomentsService) FriendCircleGetIdDetail(req dto.FriendCircleGetIdDetailRequest) (robot.SnsObjectDetailResponse, error) {
	return vars.RobotRuntime.FriendCircleGetIdDetail(robot.FriendCircleGetIdDetailRequest{
		Towxid: req.Towxid,
		Id:     req.Id,
	})
}

func (s *MomentsService) FriendCircleGetList(fristpagemd5 string, maxID string) (robot.GetListResponse, error) {
	return vars.RobotRuntime.FriendCircleGetList(fristpagemd5, maxID)
}

func (s *MomentsService) GetFriendCircleSettings() (*model.MomentSettings, error) {
	momentSettings, err := s.momentSettingsRepo.GetMomentSettings()
	if err != nil {
		return nil, fmt.Errorf("获取朋友圈设置失败: %w", err)
	}
	if momentSettings == nil {
		return &model.MomentSettings{}, nil
	}
	return momentSettings, nil
}

func (s *MomentsService) SaveFriendCircleSettings(req *model.MomentSettings) error {
	if req.ID == 0 {
		momentSettings, err := s.momentSettingsRepo.GetMomentSettings()
		if err != nil {
			return err
		}
		if momentSettings != nil {
			return fmt.Errorf("朋友圈设置已存在，不能重复创建")
		}
		return s.momentSettingsRepo.Create(req)
	}
	return s.momentSettingsRepo.Update(req)
}

func (s *MomentsService) FriendCircleDownFriendCircleMedia(url, key string) (string, error) {
	return vars.RobotRuntime.FriendCircleDownFriendCircleMedia(url, key)
}

func (s *MomentsService) FriendCircleUpload(media io.Reader) (robot.FriendCircleUploadResponse, error) {
	mediaBytes, err := io.ReadAll(media)
	if err != nil {
		return robot.FriendCircleUploadResponse{}, fmt.Errorf("读取文件内容失败: %w", err)
	}

	switch {
	case filetype.IsImage(mediaBytes):
		totalSize := len(mediaBytes)
		var width, height int
		// 图片处理
		cfg, _, err := image.DecodeConfig(bytes.NewReader(mediaBytes))
		if err != nil {
			return robot.FriendCircleUploadResponse{}, fmt.Errorf("无法解码图片: %w", err)
		}
		width = cfg.Width
		height = cfg.Height
		resp, err := vars.RobotRuntime.FriendCircleUpload(mediaBytes)
		if err != nil {
			return resp, err
		}
		resp.Size.Width = strconv.Itoa(width)
		resp.Size.Height = strconv.Itoa(height)
		resp.Size.TotalSize = strconv.Itoa(totalSize)
		return resp, nil
	case filetype.IsVideo(mediaBytes):
		return s.FriendCircleCdnSnsUploadVideo(mediaBytes)
	default:
		// 未知类型，保持媒体类型为 0
	}
	return robot.FriendCircleUploadResponse{}, fmt.Errorf("不支持的媒体类型")
}

func (s *MomentsService) FriendCircleCdnSnsUploadVideo(mediaBytes []byte) (robot.FriendCircleUploadResponse, error) {
	var width, height int
	var videoDuration float64
	totalSize := len(mediaBytes)

	mediaFileType, err := filetype.Match(mediaBytes)
	if err != nil {
		return robot.FriendCircleUploadResponse{}, fmt.Errorf("无法识别媒体类型: %w", err)
	}
	if mediaFileType == filetype.Unknown {
		return robot.FriendCircleUploadResponse{}, fmt.Errorf("未知媒体类型")
	}

	mediaExt := mediaFileType.Extension
	inFile, err := os.CreateTemp("", "wechat_moments_input_*."+mediaExt)
	if err != nil {
		return robot.FriendCircleUploadResponse{}, fmt.Errorf("创建临时输入文件失败: %w", err)
	}
	defer os.Remove(inFile.Name())

	if _, err = inFile.Write(mediaBytes); err != nil {
		return robot.FriendCircleUploadResponse{}, fmt.Errorf("写入临时文件失败: %w", err)
	}
	inFile.Close()

	outFile, err := os.CreateTemp("", "wechat_moments_output_*.mp4")
	if err != nil {
		return robot.FriendCircleUploadResponse{}, fmt.Errorf("创建临时输出文件失败: %w", err)
	}
	defer os.Remove(outFile.Name())
	outFile.Close()

	videoPath := inFile.Name()
	if strings.ToLower(mediaExt) != "mp4" {
		// avi mov mkv flv webm 格式转换成mp4
		cmd := exec.Command("ffmpeg",
			"-i", videoPath,
			"-c:v", "libx264",
			"-c:a", "aac",
			outFile.Name(),
			"-y",
		)
		if err = cmd.Run(); err != nil {
			return robot.FriendCircleUploadResponse{}, fmt.Errorf("转换视频格式失败: %w", err)
		}
		videoPath = outFile.Name()
	}

	// 使用 ffprobe 获取 width, height, duration
	cmd := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=width,height:format=duration", "-of", "default=noprint_wrappers=1:nokey=1", videoPath)
	out, err := cmd.Output()
	if err != nil {
		return robot.FriendCircleUploadResponse{}, fmt.Errorf("获取视频信息失败: %w", err)
	}
	parts := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(parts) >= 3 {
		width, _ = strconv.Atoi(parts[0])
		height, _ = strconv.Atoi(parts[1])
		videoDuration, _ = strconv.ParseFloat(parts[2], 64)
	}
	if strings.ToLower(mediaExt) != "mp4" {
		mediaBytes, err = os.ReadFile(videoPath)
		if err != nil {
			return robot.FriendCircleUploadResponse{}, fmt.Errorf("读取视频文件失败: %w", err)
		}
		totalSize = len(mediaBytes)
	}

	resp, err := vars.RobotRuntime.FriendCircleCdnSnsUploadVideo(mediaBytes, mediaBytes)
	if err != nil {
		return robot.FriendCircleUploadResponse{}, fmt.Errorf("上传视频失败: %w", err)
	}

	mediaId := uint64(0)
	mediaType := uint32(6)
	thumbUrlCount := uint32(1)
	thumbUrls := []*robot.SnsBufferUrl{
		{Url: &resp.ThumbURL},
	}

	return robot.FriendCircleUploadResponse{
		Id:            &mediaId,
		Type:          &mediaType,
		BufferUrl:     &robot.SnsBufferUrl{Url: &resp.FileURL},
		ThumbUrlCount: &thumbUrlCount,
		ThumbUrls:     thumbUrls,
		Size: robot.Size{
			Width:     strconv.Itoa(width),
			Height:    strconv.Itoa(height),
			TotalSize: strconv.Itoa(totalSize),
		},
		VideoDuration: videoDuration,
	}, nil
}

func (s *MomentsService) FriendCirclePost(req dto.MomentPostRequest) (robot.FriendCircleMessagesResponse, error) {
	var momentMessage robot.FriendCircleMessagesRequest

	if req.Content == "" && len(req.MediaList) == 0 {
		return robot.FriendCircleMessagesResponse{}, fmt.Errorf("朋友圈内容不能为空")
	}

	if len(req.WithUserList) > 0 {
		if slices.Contains(req.WithUserList, vars.RobotRuntime.WxID) {
			return robot.FriendCircleMessagesResponse{}, fmt.Errorf("提醒谁看列表不能包含自己")
		}
		momentMessage.WithUserList = strings.Join(req.WithUserList, ",")
	}

	switch req.ShareType {
	case "public":
		momentMessage.GroupUser = ""
		momentMessage.BlackList = ""
		momentMessage.Privacy = 0
	case "private":
		momentMessage.GroupUser = ""
		momentMessage.BlackList = ""
		momentMessage.Privacy = 1
	case "share_with":
		momentMessage.Privacy = 0
		if len(req.ShareWith) == 0 {
			return robot.FriendCircleMessagesResponse{}, fmt.Errorf("分享对象不能为空")
		}
		if slices.Contains(req.ShareWith, vars.RobotRuntime.WxID) {
			return robot.FriendCircleMessagesResponse{}, fmt.Errorf("分享对象列表不能包含自己")
		}
		momentMessage.GroupUser = strings.Join(req.ShareWith, ",")
		momentMessage.BlackList = ""
	case "donot_share":
		momentMessage.Privacy = 0
		if len(req.DoNotShare) == 0 {
			return robot.FriendCircleMessagesResponse{}, fmt.Errorf("黑名单列表不能为空")
		}
		if slices.Contains(req.DoNotShare, vars.RobotRuntime.WxID) {
			return robot.FriendCircleMessagesResponse{}, fmt.Errorf("黑名单列表不能包含自己")
		}
		momentMessage.GroupUser = ""
		momentMessage.BlackList = strings.Join(req.DoNotShare, ",")
	default:
		return robot.FriendCircleMessagesResponse{}, fmt.Errorf("发布朋友圈参数错误")
	}

	// 校验提醒谁看列表与分享对象/黑名单的关系
	if len(req.ShareWith) > 0 && len(req.WithUserList) > 0 {
		for _, u := range req.WithUserList {
			if !slices.Contains(req.ShareWith, u) {
				return robot.FriendCircleMessagesResponse{}, fmt.Errorf("提醒谁看的好友必须在分享对象列表中")
			}
		}
	}

	if len(req.DoNotShare) > 0 && len(req.WithUserList) > 0 {
		for _, u := range req.WithUserList {
			if slices.Contains(req.DoNotShare, u) {
				return robot.FriendCircleMessagesResponse{}, fmt.Errorf("提醒谁看的好友不能在黑名单列表中")
			}
		}
	}

	if len(req.MediaList) > 0 {
		if len(req.MediaList) > 9 {
			return robot.FriendCircleMessagesResponse{}, fmt.Errorf("朋友圈最多只能上传9张图片")
		}
		// 视频只能上传一个，且只能包含视频，图片只能上传9张，且只能都是图片
		if len(req.MediaList) > 1 {
			for _, media := range req.MediaList {
				if media.Type == nil {
					return robot.FriendCircleMessagesResponse{}, fmt.Errorf("朋友圈图片类型不能为空")
				}
				if *media.Type == 6 {
					return robot.FriendCircleMessagesResponse{}, fmt.Errorf("上传的图片不能包含视频")
				}
			}
		}
	}

	momentTimeline := robot.TimelineObject{
		ID:          0,
		Username:    vars.RobotRuntime.WxID,
		Private:     momentMessage.Privacy,
		SightFolded: 0,
		ShowFlag:    0,
		CreateTime:  uint32(time.Now().Unix()),
		AppInfo: robot.AppInfo{
			IsForceUpdate: 0,
			IsHidden:      0,
		},
		ContentDescShowType:    0,
		PublicBrandContactType: 0,
	}

	contentObject := robot.ContentObject{}
	if req.Content != "" && len(req.MediaList) == 0 {
		contentObject.ContentStyle = 2 // 文本
	} else if len(req.MediaList) == 1 && req.MediaList[0].Type != nil && *req.MediaList[0].Type == 6 {
		contentObject.ContentStyle = 15 // 视频
	} else {
		contentObject.ContentStyle = 1 // 图文
	}
	momentTimeline.ContentObject = contentObject

	if req.Content != "" {
		momentTimeline.ContentDesc = req.Content
	}

	if len(req.MediaList) > 0 {
		for _, mediaReq := range req.MediaList {
			if mediaReq.Type == nil {
				return robot.FriendCircleMessagesResponse{}, fmt.Errorf("朋友圈图片类型不能为空")
			}
			if mediaReq.BufferUrl == nil {
				return robot.FriendCircleMessagesResponse{}, fmt.Errorf("朋友圈图片BufferUrl不能为空")
			}
			mediaItem := robot.Media{
				Type:    *mediaReq.Type,
				Private: momentMessage.Privacy,
			}
			if mediaReq.Id != nil {
				mediaItem.ID = *mediaReq.Id
			}
			mediaItem.Size = mediaReq.Size
			if mediaItem.Type == 6 {
				videoDuration, err := strconv.ParseFloat(mediaReq.VideoDurationStr, 64)
				if err != nil {
					return robot.FriendCircleMessagesResponse{}, fmt.Errorf("视频时长格式错误: %w", err)
				}
				mediaItem.VideoDuration = videoDuration
			}

			mediaItemURL := robot.URL{}
			if mediaReq.BufferUrl.Type != nil {
				mediaItemURL.Type = strconv.FormatUint(uint64(*mediaReq.BufferUrl.Type), 10)
			} else {
				mediaItemURL.Type = "1"
			}
			if mediaReq.BufferUrl.Url != nil {
				mediaItemURL.Value = *mediaReq.BufferUrl.Url
			}
			mediaItem.URL = mediaItemURL

			if len(mediaReq.ThumbUrls) > 0 {
				mediaItemThumb := robot.Thumb{}
				if mediaReq.ThumbUrls[0].Type != nil {
					mediaItemThumb.Type = strconv.FormatUint(uint64(*mediaReq.ThumbUrls[0].Type), 10)
				} else {
					mediaItemThumb.Type = "1"
				}
				if mediaReq.ThumbUrls[0].Url != nil {
					mediaItemThumb.Value = *mediaReq.ThumbUrls[0].Url
				}
				mediaItem.Thumb = mediaItemThumb
			}

			momentTimeline.ContentObject.MediaList.Media = append(momentTimeline.ContentObject.MediaList.Media, mediaItem)
		}
	}

	momentTimelineBytes, err := xml.Marshal(momentTimeline)
	if err != nil {
		return robot.FriendCircleMessagesResponse{}, err
	}

	momentMessage.Content = string(momentTimelineBytes)

	return vars.RobotRuntime.FriendCircleMessages(momentMessage)
}

func (s *MomentsService) FriendCircleOperation(req dto.MomentOpRequest) (robot.SnsObjectOpResponse, error) {
	return vars.RobotRuntime.FriendCircleOperation(robot.FriendCircleOperationRequest{
		Id:        req.Id,
		Type:      req.Type,
		CommnetId: req.CommentId,
	})
}

func (s *MomentsService) FriendCirclePrivacySettings(req dto.MomentPrivacySettingsRequest) (robot.OplogResponse, error) {
	return vars.RobotRuntime.FriendCirclePrivacySettings(robot.FriendCirclePrivacySettingsRequest{
		Function: req.Function,
		Value:    req.Value,
	})
}

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
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"time"

	"wechat-robot-client/dto"
	"wechat-robot-client/pkg/robot"
	"wechat-robot-client/vars"

	"github.com/h2non/filetype"
)

type MomentsService struct {
	ctx context.Context
}

func NewMomentsService(ctx context.Context) *MomentsService {
	return &MomentsService{
		ctx: ctx,
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

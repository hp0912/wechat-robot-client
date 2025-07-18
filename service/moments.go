package service

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"time"
	"wechat-robot-client/dto"
	"wechat-robot-client/pkg/robot"
	"wechat-robot-client/vars"
)

type MomentsService struct {
	ctx context.Context
}

func NewMomentsService(ctx context.Context) *MomentsService {
	return &MomentsService{
		ctx: ctx,
	}
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
	return vars.RobotRuntime.FriendCircleUpload(mediaBytes)
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
		contentObject.ContentStyle = 1 // 文本
	} else if len(req.MediaList) == 1 && req.MediaList[0].Type != nil && *req.MediaList[0].Type == 6 {
		contentObject.ContentStyle = 15 // 视频
	} else {
		contentObject.ContentStyle = 2 // 图文
	}
	momentTimeline.ContentObject = contentObject

	if req.Content != "" {
		momentTimeline.ContentDesc = req.Content
	}

	mediaInfoCount := uint32(len(req.MediaList))
	momentMessage.MediaInfoCount = &mediaInfoCount
	momentMessage.MediaInfo = make([]*robot.MediaInfo, 0, len(req.MediaList))

	if len(req.MediaList) > 0 {
		for mediaIndex, mediaReq := range req.MediaList {
			if mediaReq.Type == nil {
				return robot.FriendCircleMessagesResponse{}, fmt.Errorf("朋友圈图片类型不能为空")
			}
			if mediaReq.BufferUrl == nil {
				return robot.FriendCircleMessagesResponse{}, fmt.Errorf("朋友圈图片BufferUrl不能为空")
			}
			// Media
			mediaItem := robot.Media{
				Type:    *mediaReq.Type,
				Private: momentMessage.Privacy,
			}
			// MediaInfo
			mediaInfo := &robot.MediaInfo{}

			if mediaReq.Id != nil {
				mediaItem.ID = *mediaReq.Id
			}
			if mediaIndex == 0 && req.Content != "" {
				mediaItem.Title = req.Content
				mediaItem.Description = req.Content
			}

			mediaItem.Size = mediaReq.Size
			if mediaItem.Type == 6 {
				vd, err := strconv.ParseFloat(mediaReq.VideoDuration, 64)
				if err != nil {
					return robot.FriendCircleMessagesResponse{}, fmt.Errorf("解析视频时长失败: %w", err)
				}
				mediaItem.VideoDuration = vd
			}

			mediaItemURL := robot.URL{}
			if mediaReq.BufferUrl.Type != nil {
				mediaInfo.Source = mediaReq.BufferUrl.Type
				mediaItemURL.Type = strconv.FormatUint(uint64(*mediaReq.BufferUrl.Type), 10)
			}
			if mediaReq.BufferUrl.Url != nil {
				mediaItemURL.Value = *mediaReq.BufferUrl.Url
			}
			mediaItem.URL = mediaItemURL

			if len(mediaReq.ThumbUrls) > 0 {
				mediaItemThumb := robot.Thumb{}
				if mediaReq.ThumbUrls[0].Type != nil {
					mediaItemThumb.Type = strconv.FormatUint(uint64(*mediaReq.ThumbUrls[0].Type), 10)
				}
				if mediaReq.ThumbUrls[0].Url != nil {
					mediaItemThumb.Value = *mediaReq.ThumbUrls[0].Url
				}
				mediaItem.Thumb = mediaItemThumb
			}

			// MediaInfo
			mediaType := robot.SnsMediaType(mediaItem.Type - 1)
			mediaInfo.MediaType = &mediaType
			playLength := uint32(mediaItem.VideoDuration)
			mediaInfo.VideoPlayLength = &playLength
			currentTime := int(time.Now().Unix())
			sessionID := "memonts-" + strconv.Itoa(currentTime)
			mediaInfo.SessionId = &sessionID
			startTime := uint32(time.Now().Unix())
			mediaInfo.StartTime = &startTime

			momentMessage.MediaInfo[mediaIndex] = mediaInfo
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

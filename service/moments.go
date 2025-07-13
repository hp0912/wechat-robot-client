package service

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
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
	// 暂时先写死
	momentMessage.Privacy = 0
	momentMessage.GroupUser = ""

	if req.Content == "" && len(req.MediaList) == 0 {
		return robot.FriendCircleMessagesResponse{}, fmt.Errorf("朋友圈内容不能为空")
	}
	if req.ShareType == "range" {
		if req.Range == "" {
			return robot.FriendCircleMessagesResponse{}, fmt.Errorf("朋友圈范围不能为空")
		}
		if req.Range == "share_with" {
			if len(req.ShareWith) == 0 {
				return robot.FriendCircleMessagesResponse{}, fmt.Errorf("分享对象不能为空")
			}
			momentMessage.WithUserList = strings.Join(req.ShareWith, ",")
		}
		if req.Range == "donot_share" {
			if len(req.DoNotShare) == 0 {
				return robot.FriendCircleMessagesResponse{}, fmt.Errorf("不分享对象不能为空")
			}
			momentMessage.BlackList = strings.Join(req.DoNotShare, ",")
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
		ID:       0,
		Username: vars.RobotRuntime.WxID,
		// 暂时先写死
		Private:     0,
		SightFolded: 0,
		ShowFlag:    0,
		CreateTime:  uint32(time.Now().Unix()),
		AppInfo: robot.AppInfo{
			IsForceUpdate: 0,
			IsHidden:      0,
		},
		ContentDescShowType:    0,
		PublicBrandContactType: 0,
		ContentObject: robot.ContentObject{
			// 暂时先写死
			ContentStyle: 2,
		},
	}
	if req.Content != "" {
		momentTimeline.ContentDesc = req.Content
	}
	if len(req.MediaList) > 0 {
		for mediaIndex, mediaReq := range req.MediaList {
			if mediaReq.Type == nil {
				return robot.FriendCircleMessagesResponse{}, fmt.Errorf("朋友圈图片类型不能为空")
			}
			mediaItem := robot.Media{
				Type: strconv.FormatUint(uint64(*mediaReq.Type), 10),
			}
			if mediaReq.Id != nil {
				mediaItem.ID = strconv.FormatUint(*mediaReq.Id, 10)
			}
			if mediaIndex == 0 && req.Content != "" {
				mediaItem.Title = req.Content
				mediaItem.Description = req.Content
			}
			if mediaReq.TotalLen != nil {
				mediaItem.Size = robot.Size{
					TotalSize: strconv.FormatUint(uint64(*mediaReq.TotalLen), 10),
				}
			}
			if mediaReq.BufferUrl != nil {
				mediaItemURL := robot.URL{}
				if mediaReq.BufferUrl.Type != nil {
					mediaItemURL.Type = strconv.FormatUint(uint64(*mediaReq.BufferUrl.Type), 10)
				}
				if mediaReq.BufferUrl.Url != nil {
					mediaItemURL.Value = *mediaReq.BufferUrl.Url
				}
				mediaItem.URL = mediaItemURL
			}
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

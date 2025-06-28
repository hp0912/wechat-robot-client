package controller

import (
	"errors"
	"wechat-robot-client/dto"
	"wechat-robot-client/pkg/appx"
	"wechat-robot-client/service"

	"github.com/gin-gonic/gin"
)

type Moments struct{}

func NewMomentsController() *Moments {
	return &Moments{}
}

func (m *Moments) FriendCircleGetList(c *gin.Context) {
	var req dto.FriendCircleGetListRequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	data, err := service.NewMomentsService(c).FriendCircleGetList(req.FristPageMd5, req.MaxID)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(data)
}

func (m *Moments) FriendCircleDownFriendCircleMedia(c *gin.Context) {
	var req dto.DownFriendCircleMediaRequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	data, err := service.NewMomentsService(c).FriendCircleDownFriendCircleMedia(req.Url, req.Key)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(data)
}

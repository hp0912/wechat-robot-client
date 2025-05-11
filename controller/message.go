package controller

import (
	"errors"
	"path/filepath"
	"wechat-robot-client/dto"
	"wechat-robot-client/pkg/appx"
	"wechat-robot-client/service"

	"github.com/gin-gonic/gin"
)

type Message struct{}

func NewMessageController() *Message {
	return &Message{}
}

func (m *Message) MessageRevoke(c *gin.Context) {
	var req dto.MessageCommonRequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	err := service.NewMessageService(c).MessageRevoke(req)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}

func (m *Message) SendTextMessage(c *gin.Context) {
	var req dto.SendTextMessageRequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	err := service.NewMessageService(c).SendTextMessage(req)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}

func (m *Message) SendImageMessage(c *gin.Context) {
	resp := appx.NewResponse(c)
	// 获取表单文件
	file, fileHeader, err := c.Request.FormFile("image")
	if err != nil {
		resp.ToErrorResponse(errors.New("获取上传文件失败"))
		return
	}
	defer file.Close()

	// 检查文件大小
	if fileHeader.Size > 50*1024*1024 { // 限制为50MB
		resp.ToErrorResponse(errors.New("文件大小不能超过50MB"))
		return
	}

	// 检查文件类型
	ext := filepath.Ext(fileHeader.Filename)
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}
	if !allowedExts[ext] {
		resp.ToErrorResponse(errors.New("不支持的图片格式"))
		return
	}

	// 解析表单参数
	var req dto.SendMessageCommonRequest
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}

	err = service.NewMessageService(c).MsgUploadImg(req.ToWxid, file)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}

func (m *Message) SendVideoMessage(c *gin.Context) {
	resp := appx.NewResponse(c)
	// 获取表单文件
	file, fileHeader, err := c.Request.FormFile("video")
	if err != nil {
		resp.ToErrorResponse(errors.New("获取上传文件失败"))
		return
	}
	defer file.Close()

	// 检查文件大小
	if fileHeader.Size > 50*1024*1024 { // 限制为50MB
		resp.ToErrorResponse(errors.New("文件大小不能超过50MB"))
		return
	}

	// 检查文件类型
	ext := filepath.Ext(fileHeader.Filename)
	allowedExts := map[string]bool{
		".mp4":  true,
		".avi":  true,
		".mov":  true,
		".mkv":  true,
		".flv":  true,
		".webm": true,
	}
	if !allowedExts[ext] {
		resp.ToErrorResponse(errors.New("不支持的视频格式"))
		return
	}

	// 解析表单参数
	var req dto.SendMessageCommonRequest
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}

	err = service.NewMessageService(c).MsgSendVideo(req.ToWxid, file, ext)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}

func (m *Message) SendVoiceMessage(c *gin.Context) {
	resp := appx.NewResponse(c)
	// 获取表单文件
	file, fileHeader, err := c.Request.FormFile("voice")
	if err != nil {
		resp.ToErrorResponse(errors.New("获取上传文件失败"))
		return
	}
	defer file.Close()

	// 检查文件大小
	if fileHeader.Size > 50*1024*1024 { // 限制为50MB
		resp.ToErrorResponse(errors.New("文件大小不能超过50MB"))
		return
	}

	// 检查文件类型
	ext := filepath.Ext(fileHeader.Filename)
	allowedExts := map[string]bool{
		".amr":   true,
		".mp3":   true,
		".silk":  true,
		".speex": true,
		".wav":   true,
		".wave":  true,
	}
	if !allowedExts[ext] {
		resp.ToErrorResponse(errors.New("不支持的音频格式"))
		return
	}

	// 解析表单参数
	var req dto.SendMessageCommonRequest
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}

	err = service.NewMessageService(c).MsgSendVoice(req.ToWxid, file, ext)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}

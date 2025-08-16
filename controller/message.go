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
	err := service.NewMessageService(c).SendTextMessage(req.ToWxid, req.Content, req.At...)
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

	_, err = service.NewMessageService(c).MsgUploadImg(req.ToWxid, file)
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
	if fileHeader.Size > 100*1024*1024 { // 限制为100MB
		resp.ToErrorResponse(errors.New("文件大小不能超过100MB"))
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
		".amr": true,
		".mp3": true,
		".wav": true,
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

func (m *Message) SendMusicMessage(c *gin.Context) {
	var req dto.SendMusicMessageRequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	err := service.NewMessageService(c).SendMusicMessage(req.ToWxid, req.Song)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}

func (m *Message) SendFileMessage(c *gin.Context) {
	resp := appx.NewResponse(c)
	// 取得分片内容
	file, fileHeader, err := c.Request.FormFile("chunk")
	if err != nil {
		resp.ToErrorResponse(errors.New("获取上传文件失败"))
		return
	}
	defer file.Close()

	if fileHeader.Size > 50000 {
		resp.ToErrorResponse(errors.New("单个分片大小不能超过50KB"))
		return
	}

	var req dto.SendFileMessageRequest
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}

	// 基本参数校验（额外）
	if req.ChunkIndex < 0 || req.TotalChunks <= 0 || req.FileSize <= 0 || req.ChunkIndex >= req.TotalChunks {
		resp.ToErrorResponse(errors.New("分片参数错误"))
		return
	}
	if len(req.FileHash) == 0 || len(req.Filename) == 0 {
		resp.ToErrorResponse(errors.New("缺少文件信息"))
		return
	}

	if err = service.NewMessageService(c).SendFileMessage(c, req, file); err != nil {
		resp.ToErrorResponse(err)
		return
	}

	resp.ToResponse(nil)
}

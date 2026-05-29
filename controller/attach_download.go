package controller

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"wechat-robot-client/dto"
	"wechat-robot-client/pkg/appx"
	"wechat-robot-client/service"
	"wechat-robot-client/utils"

	"github.com/gin-gonic/gin"
)

type AttachDownload struct {
}

const maxUploadMediaSize int64 = 50 * 1024 * 1024

func NewAttachDownloadController() *AttachDownload {
	return &AttachDownload{}
}

func (a *AttachDownload) DownloadImage(c *gin.Context) {
	var req dto.AttachDownloadRequest
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "参数错误",
		})
		return
	}
	imageData, contentType, extension, err := service.NewAttachDownloadService(c).DownloadImage(req.MessageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	// 设置响应头
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%d%s\"", req.MessageID, extension))
	c.Header("Content-Length", fmt.Sprintf("%d", len(imageData)))

	// 将图片数据写入响应
	c.Writer.WriteHeader(http.StatusOK)
	_, err = c.Writer.Write(imageData)
	if err != nil {
		// 这里已经开始写入响应，无法再更改状态码，只能记录错误
		fmt.Printf("返回图片数据失败: %v\n", err)
	}
}

func (a *AttachDownload) DownloadVoice(c *gin.Context) {
	var req dto.AttachDownloadRequest
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "参数错误",
		})
		return
	}
	voiceData, contentType, extension, err := service.NewAttachDownloadService(c).DownloadVoice(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	// 设置响应头
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%d%s\"", req.MessageID, extension))
	c.Header("Content-Length", fmt.Sprintf("%d", len(voiceData)))

	// 将语音数据写入响应
	c.Writer.WriteHeader(http.StatusOK)
	_, err = c.Writer.Write(voiceData)
	if err != nil {
		// 这里已经开始写入响应，无法再更改状态码，只能记录错误
		fmt.Printf("返回语音数据失败: %v\n", err)
	}
}

func (a *AttachDownload) DownloadFile(c *gin.Context) {
	var req dto.AttachDownloadRequest
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "参数错误",
		})
		return
	}

	reader, filename, err := service.NewAttachDownloadService(c).DownloadFile(req.MessageID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"message": err.Error()})
		return
	}
	defer reader.Close()

	// 写响应头，采用 chunked-encoding；无需提前知道 Content-Length
	c.Header("Content-Disposition",
		fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Type", "application/octet-stream")
	c.Status(http.StatusOK)

	if _, err = io.Copy(c.Writer, reader); err != nil {
		log.Printf("stream copy error: %v", err)
	}
}

func (a *AttachDownload) DownloadVideo(c *gin.Context) {
	var req dto.AttachDownloadRequest
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "参数错误",
		})
		return
	}

	reader, filename, err := service.NewAttachDownloadService(c).DownloadVideo(req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"message": err.Error()})
		return
	}
	defer reader.Close()

	// 写响应头，采用 chunked-encoding；无需提前知道 Content-Length
	c.Header("Content-Disposition",
		fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Type", "application/octet-stream")
	c.Status(http.StatusOK)

	if _, err = io.Copy(c.Writer, reader); err != nil {
		log.Printf("stream copy error: %v", err)
	}
}

func (a *AttachDownload) UploadMedia(c *gin.Context) {
	resp := appx.NewResponse(c)

	messageID, err := strconv.ParseInt(c.PostForm("message_id"), 10, 64)
	if err != nil || messageID <= 0 {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}

	mediaType := strings.TrimSpace(c.PostForm("media_type"))
	if mediaType == "" {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}

	file, fileHeader, err := c.Request.FormFile("media")
	if err != nil {
		resp.ToErrorResponse(errors.New("获取上传文件失败"))
		return
	}
	defer file.Close()

	if fileHeader.Size > maxUploadMediaSize {
		resp.ToErrorResponse(errors.New("文件大小不能超过50MB"))
		return
	}

	data, err := io.ReadAll(io.LimitReader(file, maxUploadMediaSize+1))
	if err != nil {
		resp.ToErrorResponse(fmt.Errorf("读取上传文件失败: %w", err))
		return
	}
	if int64(len(data)) > maxUploadMediaSize {
		resp.ToErrorResponse(errors.New("文件大小不能超过50MB"))
		return
	}

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}

	extension := strings.TrimSpace(c.PostForm("extension"))
	if extension == "" {
		extension = filepath.Ext(fileHeader.Filename)
	}
	if extension == "" {
		extension = utils.DetectMediaFormat(data)
	}
	if !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}

	url, err := service.NewAttachDownloadService(c).UploadDownloadedMedia(messageID, mediaType, data, contentType, extension)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}

	resp.ToResponse(gin.H{
		"message_id": messageID,
		"media_type": mediaType,
		"url":        url,
	})
}

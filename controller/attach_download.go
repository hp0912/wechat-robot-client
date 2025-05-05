package controller

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"wechat-robot-client/dto"
	"wechat-robot-client/pkg/appx"
	"wechat-robot-client/service"

	"github.com/gin-gonic/gin"
)

type AttachDownload struct {
}

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
	imageData, contentType, extension, err := service.NewAttachDownloadService(c).DownloadImage(req)
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

	reader, filename, err := service.NewAttachDownloadService(c).DownloadFile(req)
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

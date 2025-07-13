package controller

import (
	"errors"
	"path/filepath"
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

func (m *Moments) FriendCircleUpload(c *gin.Context) {
	resp := appx.NewResponse(c)
	// 获取表单文件
	file, fileHeader, err := c.Request.FormFile("media")
	if err != nil {
		resp.ToErrorResponse(errors.New("获取上传文件失败"))
		return
	}
	defer file.Close()

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
		allowedExts = map[string]bool{
			".jpg":  true,
			".jpeg": true,
			".png":  true,
			".gif":  true,
			".webp": true,
		}
		if !allowedExts[ext] {
			resp.ToErrorResponse(errors.New("不支持的图片/视频格式"))
			return
		} else {
			// 检查文件大小
			if fileHeader.Size > 50*1024*1024 { // 限制为50MB
				resp.ToErrorResponse(errors.New("图片大小不能超过50MB"))
				return
			}
		}
	} else {
		// 检查文件大小
		if fileHeader.Size > 100*1024*1024 { // 限制为100MB
			resp.ToErrorResponse(errors.New("视频大小不能超过100MB"))
			return
		}
	}

	data, err := service.NewMomentsService(c).FriendCircleUpload(file)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(data)
}

func (m *Moments) FriendCirclePost(c *gin.Context) {
	var req dto.MomentPostRequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	data, err := service.NewMomentsService(c).FriendCirclePost(req)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(data)
}

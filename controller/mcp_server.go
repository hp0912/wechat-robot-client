package controller

import (
	"errors"

	"github.com/gin-gonic/gin"

	"wechat-robot-client/model"
	"wechat-robot-client/pkg/appx"
	"wechat-robot-client/service"
)

type MCPServer struct{}

func NewMCPController() *MCPServer {
	return &MCPServer{}
}

func (s *MCPServer) GetMCPServers(c *gin.Context) {
	resp := appx.NewResponse(c)
	data, err := service.NewMCPService(c).GetMCPServers()
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(data)
}

func (s *MCPServer) GetMCPServer(c *gin.Context) {
	var req struct {
		ID uint64 `form:"id" json:"id" binding:"required"`
	}
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	data, err := service.NewMCPService(c).GetMCPServer(req.ID)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(data)
}

func (s *MCPServer) CreateMCPServer(c *gin.Context) {
	var req model.MCPServer
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	err := service.NewMCPService(c).CreateMCPServer(&req)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}

func (s *MCPServer) UpdateMCPServer(c *gin.Context) {
	var req model.MCPServer
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	err := service.NewMCPService(c).UpdateMCPServer(&req)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}

func (s *MCPServer) EnableMCPServer(c *gin.Context) {
	var req struct {
		ID uint64 `json:"id" binding:"required"`
	}
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	err := service.NewMCPService(c).EnableMCPServer(req.ID)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}

func (s *MCPServer) DisableMCPServer(c *gin.Context) {
	var req struct {
		ID uint64 `json:"id" binding:"required"`
	}
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	err := service.NewMCPService(c).DisableMCPServer(req.ID)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}

func (s *MCPServer) DeleteMCPServer(c *gin.Context) {
	var req model.MCPServer
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	err := service.NewMCPService(c).DeleteMCPServer(&req)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}

func (s *MCPServer) GetMCPServerTools(c *gin.Context) {
	var req struct {
		ID uint64 `form:"id" json:"id" binding:"required"`
	}
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	data, err := service.NewMCPService(c).GetMCPServerTools(req.ID)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(data)
}

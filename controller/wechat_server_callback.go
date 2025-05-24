package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type WechatServerCallback struct {
}

func NewWechatServerCallbackController() *WechatServerCallback {
	return &WechatServerCallback{}
}

func (ct *WechatServerCallback) SyncMessageCallback(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

func (ct *WechatServerCallback) LogoutCallback(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

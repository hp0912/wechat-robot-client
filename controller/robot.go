package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Robot struct {
}

func NewRobotController() *Robot {
	return &Robot{}
}

func (d *Robot) Probe(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

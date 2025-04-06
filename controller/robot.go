package controller

import "github.com/gin-gonic/gin"

type Robot struct {
}

func NewRobotController() *Robot {
	return &Robot{}
}

func (d *Robot) Test(c *gin.Context) {
	//
}

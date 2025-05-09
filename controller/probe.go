package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Probe struct {
}

func NewProbeController() *Probe {
	return &Probe{}
}

func (p *Probe) Probe(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

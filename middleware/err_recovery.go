package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ErrorRecover(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("服务器内部错误: %v", err)
			message := "服务器内部错误"
			if err, ok := err.(error); ok {
				message = err.Error()
			}
			response := gin.H{"code": 500, "message": message, "data": nil}
			c.JSON(http.StatusOK, response)
			c.Abort()
		}
	}()
	c.Next()
}

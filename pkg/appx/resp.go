package appx

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Response struct {
	Ctx *gin.Context
}

func NewResponse(ctx *gin.Context) *Response {
	return &Response{
		Ctx: ctx,
	}
}

func (r *Response) ToResponseWithHttpCode(code int, data any) {
	r.Ctx.JSON(code, data)
}

func (r *Response) ToResponse(data any) {
	r.Ctx.JSON(http.StatusOK, gin.H{"code": 200, "message": "", "data": data})
}
func (r *Response) ToResponseData(data any) {
	r.Ctx.JSON(http.StatusOK, gin.H{"name": data})
}

func (r *Response) ToResponseList(list any, totalRows int64) {
	r.Ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "请求成功",
		"data": gin.H{
			"items": list,
			"total": totalRows,
		},
	})
}

func (r *Response) ToErrorResponse(err error) {
	response := gin.H{"code": 500, "message": err.Error(), "data": nil}
	r.Ctx.JSON(http.StatusOK, response)
}

func (r *Response) To401Response(err error) {
	response := gin.H{"code": 401, "message": err.Error(), "data": nil}
	r.Ctx.JSON(http.StatusOK, response)
}

// 处理 Gin validator 错误
func (r *Response) ToValidatorError(err error) {
	if ve, ok := err.(validator.ValidationErrors); ok {
		details := []string{}
		for _, fe := range ve {
			details = append(details, fmt.Sprintf("Field: %s Error: failed on the '%s' tag", fe.Field(), fe.Tag()))
		}
		err = errors.New(strings.Join(details, "; "))
	}
	r.ToErrorResponse(err)
}

func (r *Response) ToInvalidResponse(err ValidErrors) {
	r.Ctx.JSON(http.StatusOK, gin.H{"code": 400, "message": err.Error(), "data": struct{}{}})
}
func (r *Response) ToInvalidResponseMsg(msg string) {
	r.Ctx.JSON(http.StatusOK, gin.H{"code": 400, "message": msg, "data": struct{}{}})
}

func (r *Response) ToInvalidResponseWithEmptyArr(err ValidErrors) {
	r.Ctx.JSON(http.StatusOK, gin.H{"code": 400, "message": err.Error(), "data": []string{}})
}

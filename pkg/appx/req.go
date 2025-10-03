package appx

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type ValidError struct {
	Key     string
	Message string
}

type ValidErrors []*ValidError

func (v *ValidError) Error() string {
	return v.Message
}

func (v ValidErrors) Error() string {
	return strings.Join(v.Errors(), ",")
}

func (v ValidErrors) Errors() []string {
	var errs []string
	for _, err := range v {
		errs = append(errs, err.Error())
	}

	return errs
}

func BindAndValid(c *gin.Context, v any) (isValid bool, errs ValidErrors) {
	err := c.ShouldBind(v)
	if err != nil {
		return false, errs
	}

	return true, nil
}

type Pager struct {
	// 页码
	PageIndex int `json:"page_index"`
	// 每页数量
	PageSize int `json:"page_size"`
	// 数据库偏移
	OffSet int
}

func InitPager(c *gin.Context) Pager {
	pager := Pager{}
	var err error
	index := c.DefaultQuery("page_index", "1")
	pager.PageIndex, err = strconv.Atoi(index)
	if err != nil {
		pager.PageIndex = 1
	}
	size := c.DefaultQuery("page_size", "20")
	pager.PageSize, err = strconv.Atoi(size)
	if err != nil {
		pager.PageSize = 20
	}
	pager.OffSet = (pager.PageIndex - 1) * pager.PageSize
	return pager
}

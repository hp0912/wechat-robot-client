package pkg

import (
	"fmt"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	hunyuan "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/hunyuan/v20230901"
)

type LogoRect struct {
	X      int `json:"X"`
	Y      int `json:"Y"`
	Width  int `json:"Width"`
	Height int `json:"Height"`
}

type LogoParam struct {
	LogoUrl   *string   `json:"LogoUrl"`
	LogoImage *string   `json:"LogoImage"`
	LogoRect  *LogoRect `json:"LogoRect"`
}

type HunyuanConfig struct {
	SecretId  string     `json:"SecretId"`
	SecretKey string     `json:"SecretKey"`
	ChatId    *string    `json:"ChatId"`
	LogoAdd   int        `json:"LogoAdd"`
	LogoParam *LogoParam `json:"LogoParam"`
	Prompt    string     `json:"prompt"`
}

// SubmitHunyuanDrawing 腾讯混元绘图
func SubmitHunyuanDrawing(config *HunyuanConfig) (string, error) {
	credential := common.NewCredential(config.SecretId, config.SecretKey)

	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "hunyuan.tencentcloudapi.com"

	client, _ := hunyuan.NewClient(credential, "", cpf)

	request := hunyuan.NewSubmitHunyuanImageChatJobRequest()
	// 返回的resp是一个SubmitHunyuanImageChatJobResponse的实例，与请求对象对应
	response, err := client.SubmitHunyuanImageChatJob(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return "", fmt.Errorf("TencentCloud SDK error: %s", err)
	}

	return *response.Response.JobId, nil
}

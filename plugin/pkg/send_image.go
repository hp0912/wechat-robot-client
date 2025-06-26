package pkg

import (
	"io"
	"os"
	"wechat-robot-client/interface/plugin"

	"github.com/go-resty/resty/v2"
)

func SendImageByURL(MessageService plugin.MessageServiceIface, toWxID, imageUrl string) error {
	resp, err := resty.New().R().SetDoNotParseResponse(true).Get(imageUrl)
	if err != nil {
		return err
	}
	defer resp.RawBody().Close()
	// 创建临时文件
	tempFile, err := os.CreateTemp("", "doubao_image_*")
	if err != nil {
		return err
	}
	defer tempFile.Close()
	defer os.Remove(tempFile.Name()) // 清理临时文件
	// 将图片数据写入临时文件
	_, err = io.Copy(tempFile, resp.RawBody())
	if err != nil {
		return err
	}
	// 重置文件指针到开始位置
	_, err = tempFile.Seek(0, 0)
	if err != nil {
		return err
	}
	return MessageService.MsgUploadImg(toWxID, tempFile)
}

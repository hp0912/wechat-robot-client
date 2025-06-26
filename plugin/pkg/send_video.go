package pkg

import (
	"io"
	"os"
	"wechat-robot-client/interface/plugin"

	"github.com/go-resty/resty/v2"
)

func SendVideoByURL(MessageService plugin.MessageServiceIface, toWxID, videoUrl string) error {
	resp, err := resty.New().R().SetDoNotParseResponse(true).Get(videoUrl)
	if err != nil {
		return err
	}
	defer resp.RawBody().Close()
	// 创建临时文件
	tempFile, err := os.CreateTemp("", "doubao_video_*")
	if err != nil {
		return err
	}
	defer tempFile.Close()
	defer os.Remove(tempFile.Name()) // 清理临时文件
	// 将视频数据写入临时文件
	_, err = io.Copy(tempFile, resp.RawBody())
	if err != nil {
		return err
	}
	// 重置文件指针到开始位置
	_, err = tempFile.Seek(0, 0)
	if err != nil {
		return err
	}
	return MessageService.MsgSendVideo(toWxID, tempFile, ".mp4")
}

package pkg

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
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
	u, err := url.Parse(videoUrl)
	if err != nil {
		return err
	}
	videoExt := ".mp4"
	queryParams := u.Query()
	if queryParams.Get("mime_type") != "" {
		extStrs := strings.Split(queryParams.Get("mime_type"), "_")
		if len(extStrs) > 0 {
			videoExt = fmt.Sprintf(".%s", extStrs[len(extStrs)-1])
		}
	}
	return MessageService.MsgSendVideo(toWxID, tempFile, videoExt)
}

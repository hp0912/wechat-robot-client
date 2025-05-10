package robot

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"wechat-robot-client/model"
)

type Robot struct {
	RobotID            int64
	WxID               string
	Status             model.RobotStatus
	DeviceID           string
	DeviceName         string
	Client             *Client
	HeartbeatContext   context.Context
	HeartbeatCancel    func()
	SyncMessageContext context.Context
	SyncMessageCancel  func()
}

func (r *Robot) IsRunning() bool {
	return r.Client.IsRunning()
}

func (r *Robot) IsLoggedIn() bool {
	if r.WxID == "" {
		return false
	}
	_, err := r.Client.GetProfile(r.WxID)
	return err == nil
}

func (r *Robot) GetQrCode() (uuid string, err error) {
	var resp GetQRCode
	resp, err = r.Client.GetQrCode(r.DeviceID, r.DeviceName)
	if err != nil {
		return
	}
	if resp.Uuid != "" {
		uuid = resp.Uuid
		return
	}
	err = errors.New("获取二维码失败")
	return
}

func (r *Robot) GetProfile(wxid string) (UserProfile, error) {
	return r.Client.GetProfile(wxid)
}

func (r *Robot) Login() (uuid string, awkenLogin, autoLogin bool, err error) {
	// 尝试唤醒登陆
	var cachedInfo CachedInfo
	cachedInfo, err = r.Client.GetCachedInfo(r.WxID)
	if err == nil && cachedInfo.Wxid != "" {
		err = r.LoginTwiceAutoAuth()
		if err == nil {
			autoLogin = true
			return
		}
		// 唤醒登陆
		var resp QrCode
		resp, err = r.Client.AwakenLogin(r.WxID, r.DeviceName)
		if err != nil {
			// 如果唤醒失败，尝试获取二维码
			uuid, err = r.GetQrCode()
			return
		}
		if resp.Uuid == "" {
			// 如果唤醒失败，尝试获取二维码
			uuid, err = r.GetQrCode()
			return
		}
		// 唤醒登陆成功
		uuid = resp.Uuid
		awkenLogin = true
		return
	}
	// 二维码登陆
	uuid, err = r.GetQrCode()
	return
}

func (r *Robot) XmlFastDecoder(xmlStr, target string) string {
	decoder := xml.NewDecoder(strings.NewReader(xmlStr))
	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return ""
		}
		switch el := tok.(type) {
		case xml.StartElement:
			if el.Name.Local == target {
				var content string
				decoder.DecodeElement(&content, &el)
				return content
			}
		}
	}
	return ""
}

func (r *Robot) XmlDecoder(xmlStr string, result any) error {
	decoder := xml.NewDecoder(strings.NewReader(xmlStr))
	err := decoder.Decode(result)
	if err != nil {
		return err
	}
	if result == nil {
		return errors.New("解析失败")
	}
	return nil
}

func (r *Robot) DetectImageType(data []byte) (string, string) {
	if len(data) < 8 {
		return "application/octet-stream", ""
	}
	// 检查常见图片格式的magic bytes
	switch {
	case bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF}):
		return "image/jpeg", ".jpg"
	case bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}):
		return "image/png", ".png"
	case bytes.HasPrefix(data, []byte{0x47, 0x49, 0x46, 0x38}):
		return "image/gif", ".gif"
	case bytes.HasPrefix(data, []byte{0x52, 0x49, 0x46, 0x46}) && bytes.Equal(data[8:12], []byte{0x57, 0x45, 0x42, 0x50}):
		return "image/webp", ".webp"
	case bytes.HasPrefix(data, []byte{0x42, 0x4D}):
		return "image/bmp", ".bmp"
	default:
		return "application/octet-stream", ""
	}
}

func (r *Robot) ProcessBase64Image(base64Data string) ([]byte, string, string, error) {
	imageData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, "", "", err
	}

	// 使用图片数据的magic bytes检测格式
	contentType, extension := r.DetectImageType(imageData)
	return imageData, contentType, extension, nil
}

func (r *Robot) DownloadImage(message model.Message) ([]byte, string, string, error) {
	var imgXml ImageMessageXml
	err := r.XmlDecoder(message.Content, &imgXml)
	if err != nil {
		return nil, "", "", err
	}
	var base64Data string
	base64Data, err = r.Client.CdnDownloadImg(r.WxID, imgXml.Img.AesKey, imgXml.Img.CdnMidImgUrl)
	if err != nil {
		return nil, "", "", err
	}
	return r.ProcessBase64Image(base64Data)
}

func (r *Robot) DownloadVideo(ctx context.Context, message model.Message) (io.ReadCloser, string, error) {
	// 解析消息中的文件信息
	var videoXml VideoMessageXml
	if err := r.XmlDecoder(message.Content, &videoXml); err != nil {
		return nil, "", err
	}

	// 客户端期望的文件名（带扩展名）
	filename := fmt.Sprintf("%d.%s", message.ID, "mp4")
	// 分片信息
	totalLen := videoXml.VideoMsg.Length
	const chunkSize = int64(60 * 1024) // 60 KB
	// 使用 io.Pipe 将分片数据实时写入响应流
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()

		for startPos := int64(0); startPos < totalLen; startPos += chunkSize {
			// 检查上下文是否已取消
			select {
			case <-ctx.Done():
				_ = pw.CloseWithError(ctx.Err())
				return
			default:
				// 继续执行
			}

			currentChunkSize := chunkSize
			if startPos+currentChunkSize > totalLen {
				currentChunkSize = totalLen - startPos
			}

			// 调用底层接口获取单个分片（base64 编码）
			base64Data, err := r.Client.DownloadVideo(DownloadVideoRequest{
				Wxid:    r.WxID,
				ToWxid:  videoXml.VideoMsg.FromUserName,
				DataLen: totalLen,
				MsgId:   message.ClientMsgId,
				Section: Section{
					StartPos: startPos,
					DataLen:  currentChunkSize,
				},
				CompressType: 0,
			})
			if err != nil {
				// 将错误传递给读取端
				_ = pw.CloseWithError(fmt.Errorf("下载文件分片错误: %w (start=%d size=%d)", err, startPos, currentChunkSize))
				return
			}

			// base64 解码
			raw, err := base64.StdEncoding.DecodeString(base64Data)
			if err != nil {
				_ = pw.CloseWithError(fmt.Errorf("base64 解码失败: %w", err))
				return
			}

			// 写入到 PipeWriter，让外层按需读取（流式传输）
			if _, err := pw.Write(raw); err != nil {
				_ = pw.CloseWithError(err)
				return
			}
		}
	}()

	// 返回可读取的流和文件名，调用者自行决定如何处理（例如直接写入 HTTP 响应）
	return pr, filename, nil
}

func (r *Robot) DownloadVoice(ctx context.Context, message model.Message) ([]byte, string, string, error) {
	var voiceXml VoiceMessageXml
	err := r.XmlDecoder(message.Content, &voiceXml)
	if err != nil {
		return nil, "", "", err
	}

	// 检查上下文是否已取消
	if ctx.Err() != nil {
		return nil, "", "", ctx.Err()
	}

	var base64Data string
	base64Data, err = r.Client.DownloadVoice(DownloadVoiceRequest{
		Wxid:         r.WxID,
		MsgId:        message.ClientMsgId,
		Bufid:        voiceXml.Voicemsg.BufID,
		FromUserName: voiceXml.Voicemsg.FromUsername,
		Length:       voiceXml.Voicemsg.Length,
	})
	if err != nil {
		return nil, "", "", err
	}
	silkData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, "", "", fmt.Errorf("将silk base64编码转成字节数组错误: %w", err)
	}
	inFile, err := os.CreateTemp("", "silk_*.silk")
	if err != nil {
		return nil, "", "", fmt.Errorf("创建silk临时文件错误: %w", err)
	}
	defer os.Remove(inFile.Name())
	if _, err = inFile.Write(silkData); err != nil {
		inFile.Close()
		return nil, "", "", fmt.Errorf("写入silk临时文件错误: %w", err)
	}
	inFile.Close()

	// 检查上下文是否已取消
	if ctx.Err() != nil {
		return nil, "", "", ctx.Err()
	}

	cmd := exec.CommandContext(ctx, "silk-convert", inFile.Name(), "wav")
	if err = cmd.Run(); err != nil {
		return nil, "", "", fmt.Errorf("silk-convert执行转换错误: %w", err)
	}
	wavFile := strings.Replace(inFile.Name(), ".silk", ".wav", 1)
	wavData, err := os.ReadFile(wavFile)
	if err != nil {
		return nil, "", "", fmt.Errorf("读取wav文件错误: %w", err)
	}
	defer os.Remove(wavFile)
	return wavData, "audio/wav", ".wav", nil
}

func (r *Robot) DownloadFile(ctx context.Context, message model.Message) (io.ReadCloser, string, error) {
	// 解析消息中的文件信息
	var fileXml FileMessageXml
	if err := r.XmlDecoder(message.Content, &fileXml); err != nil {
		return nil, "", err
	}

	// 客户端期望的文件名（带扩展名）
	filename := fmt.Sprintf("%d.%s", message.ID, fileXml.Appmsg.Attach.FileExt)
	// 分片信息
	totalLen := fileXml.Appmsg.Attach.TotalLen
	const chunkSize = int64(60 * 1024) // 60 KB
	// 使用 io.Pipe 将分片数据实时写入响应流
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()

		for startPos := int64(0); startPos < totalLen; startPos += chunkSize {
			// 检查上下文是否已取消
			select {
			case <-ctx.Done():
				_ = pw.CloseWithError(ctx.Err())
				return
			default:
				// 继续执行
			}

			currentChunkSize := chunkSize
			if startPos+currentChunkSize > totalLen {
				currentChunkSize = totalLen - startPos
			}

			// 调用底层接口获取单个分片（base64 编码）
			base64Data, err := r.Client.DownloadFile(DownloadFileRequest{
				Wxid:     r.WxID,
				AttachId: fileXml.Appmsg.Attach.AttachID,
				AppID:    fileXml.Appmsg.AppID,
				UserName: fileXml.FromUsername,
				DataLen:  totalLen,
				Section: Section{
					StartPos: startPos,
					DataLen:  currentChunkSize,
				},
			})
			if err != nil {
				// 将错误传递给读取端
				_ = pw.CloseWithError(fmt.Errorf("下载文件分片错误: %w (start=%d size=%d)", err, startPos, currentChunkSize))
				return
			}

			// base64 解码
			raw, err := base64.StdEncoding.DecodeString(base64Data)
			if err != nil {
				_ = pw.CloseWithError(fmt.Errorf("base64 解码失败: %w", err))
				return
			}

			// 写入到 PipeWriter，让外层按需读取（流式传输）
			if _, err := pw.Write(raw); err != nil {
				_ = pw.CloseWithError(err)
				return
			}
		}
	}()

	// 返回可读取的流和文件名，调用者自行决定如何处理（例如直接写入 HTTP 响应）
	return pr, filename, nil
}

func (r *Robot) LoginTwiceAutoAuth() error {
	return r.Client.LoginTwiceAutoAuth(r.WxID)
}

func (r *Robot) SyncMessage() (SyncMessage, error) {
	return r.Client.SyncMessage(r.WxID)
}

func (r *Robot) MessageRevoke(message model.Message) error {
	return r.Client.MessageRevoke(MessageRevokeRequest{
		Wxid:        r.WxID,
		NewMsgId:    message.MsgId,
		ClientMsgId: message.ClientMsgId,
		ToUserName:  message.FromWxID,
		CreateTime:  message.CreatedAt,
	})
}

func (r *Robot) SendTextMessage(toWxID, content string, at ...string) (SendTextMessageResponse, string, error) {
	atMsg := ""
	if len(at) > 0 {
		for _, wxid := range at {
			contacts, err := r.GetContactDetail([]string{wxid})
			if err != nil || len(contacts) == 0 {
				continue
			}
			atMsg += fmt.Sprintf("@%s%s", contacts[0].NickName.String, "\u2005")
		}
	}
	newMessageContent := atMsg + content
	newMessages, err := r.Client.SendTextMessage(SendTextMessageRequest{
		Wxid:    r.WxID,
		Type:    1,
		ToWxid:  toWxID,
		Content: newMessageContent,
		At:      strings.Join(at, ","),
	})
	if err != nil {
		return SendTextMessageResponse{}, "", err
	}
	return newMessages, newMessageContent, nil
}

func (r *Robot) MsgUploadImg(toWxID string, image []byte) (MsgUploadImgResponse, error) {
	base64Str := base64.StdEncoding.EncodeToString(image)
	imageMessage, err := r.Client.MsgUploadImg(r.WxID, toWxID, base64Str)
	if err != nil {
		return MsgUploadImgResponse{}, err
	}
	return imageMessage, nil
}

func (r *Robot) CheckLoginUuid(uuid string) (CheckUuid, error) {
	return r.Client.CheckLoginUuid(uuid)
}

func (r *Robot) Logout() error {
	if r.WxID == "" {
		return errors.New("您还未登陆")
	}
	return r.Client.Logout(r.WxID)
}

func (r *Robot) Heartbeat() error {
	return r.Client.Heartbeat(r.WxID)
}

func (r *Robot) GetContactList() ([]string, error) {
	return r.Client.GetContactList(r.WxID)
}

func (r *Robot) GetContactDetail(requestWxids []string) ([]Contact, error) {
	return r.Client.GetContactDetail(r.WxID, requestWxids)
}

func (r *Robot) GetChatRoomMemberDetail(QID string) ([]ChatRoomMember, error) {
	return r.Client.GetChatRoomMemberDetail(r.WxID, QID)
}

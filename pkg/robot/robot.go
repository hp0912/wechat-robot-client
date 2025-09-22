package robot

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"text/template"
	"wechat-robot-client/model"
)

type Robot struct {
	RobotID           int64
	WxID              string
	Status            model.RobotStatus
	DeviceID          string
	DeviceName        string
	Client            *Client
	HeartbeatContext  context.Context
	HeartbeatCancel   func()
	SyncMomentContext context.Context
	SyncMomentCancel  func()
}

// 实现优雅退出接口
func (r *Robot) Name() string {
	return "微信机器人"
}

func (r *Robot) Shutdown(ctx context.Context) error {
	if r.WxID == "" {
		return nil
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		if r.SyncMomentCancel != nil {
			r.SyncMomentCancel()
		}
		// 手动心跳才需要取消，现在改成了自动心跳，其实下面等代码没什么用
		if r.HeartbeatCancel != nil {
			r.HeartbeatCancel()
		}
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
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

func (r *Robot) GetQrCode(loginType string, isPretender bool) (loginData LoginResponse, err error) {
	var resp GetQRCode
	switch loginType {
	case "mac":
		if isPretender {
			resp, err = r.Client.GetQrCode(loginType, r.DeviceID, r.DeviceName)
		} else {
			resp, err = r.Client.LoginGetQRMac(r.DeviceID, "Mac Book Pro")
		}
		if err != nil {
			return
		}
	default:
		resp, err = r.Client.GetQrCode(loginType, r.DeviceID, r.DeviceName)
		if err != nil {
			return
		}
	}
	if resp.Uuid != "" {
		loginData.Uuid = resp.Uuid
		loginData.Data62 = resp.Data62
		return
	}
	err = errors.New("获取二维码失败")
	return
}

func (r *Robot) GetProfile(wxid string) (GetProfileResponse, error) {
	return r.Client.GetProfile(wxid)
}

func (r *Robot) GetCachedInfo() (LoginData, error) {
	return r.Client.GetCachedInfo(r.WxID)
}

func (r *Robot) Login(loginType string, isPretender bool) (loginData LoginResponse, err error) {
	if isPretender {
		// 二维码登陆
		loginData, err = r.GetQrCode(loginType, true)
		return
	}
	// 尝试唤醒登陆
	var cachedInfo LoginData
	cachedInfo, err = r.Client.GetCachedInfo(r.WxID)
	if err == nil && cachedInfo.Wxid != "" {
		err = r.LoginTwiceAutoAuth()
		if err == nil {
			loginData.AutoLogin = true
			return
		}
		// 唤醒登陆
		var resp QrCode
		resp, err = r.Client.AwakenLogin(r.WxID)
		if err != nil {
			// 如果唤醒失败，尝试获取二维码
			loginData, err = r.GetQrCode(loginType, false)
			return
		}
		if resp.Uuid == "" {
			// 如果唤醒失败，尝试获取二维码
			loginData, err = r.GetQrCode(loginType, false)
			return
		}
		// 唤醒登陆成功
		loginData.Uuid = resp.Uuid
		loginData.AwkenLogin = true
		return
	}
	// 二维码登陆
	loginData, err = r.GetQrCode(loginType, false)
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

	var voiceDataStr strings.Builder
	var base64Data string
	totalLen := voiceXml.Voicemsg.Length
	const chunkSize = int64(60 * 1024) // 60 KB
	for startPos := int64(0); startPos < totalLen; startPos += chunkSize {
		currentChunkSize := chunkSize
		if startPos+currentChunkSize > totalLen {
			currentChunkSize = totalLen - startPos
		}
		base64Data, err = r.Client.DownloadVoice(DownloadVoiceRequest{
			Wxid:         r.WxID,
			MsgId:        message.MsgId,
			Bufid:        voiceXml.Voicemsg.BufID,
			FromUserName: voiceXml.Voicemsg.FromUsername,
			Offset:       startPos,
			Length:       currentChunkSize,
		})
		if err != nil {
			return nil, "", "", err
		}
		voiceDataStr.WriteString(base64Data)
	}
	silkData, err := base64.StdEncoding.DecodeString(voiceDataStr.String())
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

func (r *Robot) SendTextMessage(toWxID, content string, at ...string) (SendTextMessageResponse, error) {
	textMessage, err := r.Client.SendTextMessage(SendTextMessageRequest{
		Wxid:    r.WxID,
		Type:    1,
		ToWxid:  toWxID,
		Content: content,
		At:      strings.Join(at, ","),
	})
	if err != nil {
		return SendTextMessageResponse{}, err
	}
	return textMessage, nil
}

func (r *Robot) MsgSendGroupMassMsgText(req MsgSendGroupMassMsgTextRequest) (MsgSendGroupMassMsgTextResponse, error) {
	req.Wxid = r.WxID
	resp, err := r.Client.MsgSendGroupMassMsgText(req)
	if err != nil {
		return MsgSendGroupMassMsgTextResponse{}, err
	}
	return resp, nil
}

func (r *Robot) MsgUploadImg(toWxID string, image []byte) (MsgUploadImgResponse, error) {
	base64Str := base64.StdEncoding.EncodeToString(image)
	imageMessage, err := r.Client.MsgUploadImg(r.WxID, toWxID, base64Str)
	if err != nil {
		return MsgUploadImgResponse{}, err
	}
	return imageMessage, nil
}

func (r *Robot) MsgSendVideo(toWxID string, video []byte, videoExt string) (videoMessage MsgSendVideoResponse, err error) {
	inFile, err := os.CreateTemp("", "video_input_*"+videoExt)
	if err != nil {
		err = fmt.Errorf("创建临时输入文件失败: %w", err)
		return
	}
	defer os.Remove(inFile.Name())

	if _, err = inFile.Write(video); err != nil {
		err = fmt.Errorf("写入临时输入文件失败: %w", err)
		return
	}
	inFile.Close()

	outFile, err := os.CreateTemp("", "video_output_*.mp4")
	if err != nil {
		err = fmt.Errorf("创建临时输出文件失败: %w", err)
		return
	}
	defer os.Remove(outFile.Name())
	outFile.Close()

	thumbFile, err := os.CreateTemp("", "video_thumb_*.jpg")
	if err != nil {
		err = fmt.Errorf("创建临时缩略图文件失败: %w", err)
		return
	}
	defer os.Remove(thumbFile.Name())
	thumbFile.Close()

	videoPath := inFile.Name()
	if strings.ToLower(videoExt) != ".mp4" {
		// avi mov mkv flv webm 格式转换成mp4
		cmd := exec.Command("ffmpeg",
			"-i", videoPath,
			"-c:v", "libx264",
			"-c:a", "aac",
			outFile.Name(),
			"-y",
		)
		if err = cmd.Run(); err != nil {
			err = fmt.Errorf("视频格式转换失败: %w", err)
			return
		}
		videoPath = outFile.Name()
	}

	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", videoPath)
	output, err := cmd.Output()
	if err != nil {
		err = fmt.Errorf("获取视频时长失败: %w", err)
		return
	}

	duration, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		err = fmt.Errorf("解析视频时长失败: %w", err)
		return
	}
	playLength := int64(duration)

	cmd = exec.Command("ffmpeg", "-i", videoPath, "-ss", "00:00:01", "-vframes", "1", thumbFile.Name(), "-y")
	if err = cmd.Run(); err != nil {
		err = fmt.Errorf("提取缩略图失败: %w", err)
		return
	}

	videoData, err := os.ReadFile(videoPath)
	if err != nil {
		err = fmt.Errorf("读取视频文件失败: %w", err)
		return
	}

	thumbData, err := os.ReadFile(thumbFile.Name())
	if err != nil {
		err = fmt.Errorf("读取缩略图文件失败: %w", err)
		return
	}
	if len(thumbData) == 0 {
		err = fmt.Errorf("缩略图文件为空")
		return
	}

	videoBase64 := base64.StdEncoding.EncodeToString(videoData)
	thumbBase64 := base64.StdEncoding.EncodeToString(thumbData)

	videoMessage, err = r.Client.MsgSendVideo(MsgSendVideoRequest{
		Wxid:        r.WxID,
		ToWxid:      toWxID,
		Base64:      videoBase64,
		ImageBase64: thumbBase64,
		PlayLength:  playLength,
	})
	if err != nil {
		return
	}

	return
}

func (r *Robot) GetClosestSampleRate(frameRate int) int {
	supported := []int{8000, 12000, 16000, 24000}
	closestRate := supported[0]
	smallestDiff := math.MaxInt32

	for _, rate := range supported {
		diff := int(math.Abs(float64(frameRate - rate)))
		if diff < smallestDiff {
			smallestDiff = diff
			closestRate = rate
		}
	}

	return closestRate
}

func (r *Robot) MsgSendVoice(toWxID string, voice []byte, voiceExt string) (voiceMessage MsgSendVoiceResponse, err error) {
	voiceTypeMap := map[string]int{
		".amr": 0,
		".mp3": 4,
		".wav": 4,
	}

	inFile, err := os.CreateTemp("", "voice_input_*"+voiceExt)
	if err != nil {
		err = fmt.Errorf("创建临时输入文件失败: %w", err)
		return
	}
	defer os.Remove(inFile.Name())

	if _, err = inFile.Write(voice); err != nil {
		err = fmt.Errorf("写入临时输入文件失败: %w", err)
		return
	}
	inFile.Close()

	outFile, err := os.CreateTemp("", "voice_output_*"+voiceExt)
	if err != nil {
		err = fmt.Errorf("创建临时输出文件失败: %w", err)
		return
	}
	defer os.Remove(outFile.Name())
	outFile.Close()

	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", inFile.Name())
	output, err := cmd.Output()
	if err != nil {
		err = fmt.Errorf("获取音频时长失败: %w", err)
		return
	}

	duration, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		err = fmt.Errorf("解析音频时长失败: %w", err)
		return
	}

	voiceTime := int(duration * 1000)

	truncateOption := []string{}
	if duration > 59 {
		truncateOption = append(truncateOption, "-t", "59")
		voiceTime = 59000
	}

	var targetRate int
	if strings.ToLower(voiceExt) == ".mp3" || strings.ToLower(voiceExt) == ".wav" {
		cmd = exec.Command("ffprobe", "-v", "error", "-select_streams", "a:0", "-show_entries", "stream=sample_rate", "-of", "default=noprint_wrappers=1:nokey=1", inFile.Name())
		output, err = cmd.Output()
		if err != nil {
			err = fmt.Errorf("获取音频采样率失败: %w", err)
			return
		}

		var sampleRate int
		sampleRate, err = strconv.Atoi(strings.TrimSpace(string(output)))
		if err != nil {
			err = fmt.Errorf("解析音频采样率失败: %w", err)
			return
		}

		targetRate = r.GetClosestSampleRate(sampleRate)

		args := []string{
			"-i", inFile.Name(),
			"-ac", "1",
			"-ar", strconv.Itoa(targetRate),
		}
		args = append(args, truncateOption...)
		args = append(args, outFile.Name(), "-y")

		cmd = exec.Command("ffmpeg", args...)
		if err = cmd.Run(); err != nil {
			err = fmt.Errorf("音频处理失败: %w", err)
			return
		}
	} else if len(truncateOption) > 0 {
		args := []string{
			"-i", inFile.Name(),
		}
		args = append(args, truncateOption...)
		args = append(args, outFile.Name(), "-y")

		cmd = exec.Command("ffmpeg", args...)
		if err = cmd.Run(); err != nil {
			err = fmt.Errorf("音频截断失败: %w", err)
			return
		}
	} else {
		if _, err = os.Stat(inFile.Name()); err != nil {
			err = fmt.Errorf("临时音频文件不存在: %w", err)
			return
		}
	}

	var processedVoice []byte
	if len(truncateOption) > 0 || (strings.ToLower(voiceExt) == ".mp3" || strings.ToLower(voiceExt) == ".wav") {
		// 需要转换成 silk 格式
		var pcmFile *os.File
		pcmFile, err = os.CreateTemp("", "voice_pcm_*.pcm")
		if err != nil {
			err = fmt.Errorf("创建临时pcm文件失败: %w", err)
			return
		}
		defer os.Remove(pcmFile.Name())
		pcmFile.Close()

		cmd := exec.Command("ffmpeg",
			"-i", outFile.Name(),
			"-f", "s16le",
			"-ar", strconv.Itoa(targetRate),
			"-ac", "1",
			"-acodec", "pcm_s16le",
			pcmFile.Name(),
			"-y",
		)
		if err = cmd.Run(); err != nil {
			err = fmt.Errorf("转pcm文件失败: %w", err)
			return
		}

		silkFilename := strings.Replace(pcmFile.Name(), ".pcm", ".silk", 1)
		defer os.Remove(silkFilename)

		cmd = exec.Command("silk-encoder", pcmFile.Name(), silkFilename, "-tencent")
		if err = cmd.Run(); err != nil {
			err = fmt.Errorf("decoder转换pcm文件到silk文件错误: %w", err)
			return
		}

		var silkData []byte
		silkData, err = os.ReadFile(silkFilename)
		if err != nil {
			err = fmt.Errorf("读取silk文件错误: %w", err)
			return
		}

		processedVoice = silkData
	} else {
		processedVoice = voice
	}

	base64Str := base64.StdEncoding.EncodeToString(processedVoice)
	voiceMessage, err = r.Client.MsgSendVoice(MsgSendVoiceRequest{
		Wxid:      r.WxID,
		ToWxid:    toWxID,
		Base64:    base64Str,
		Type:      voiceTypeMap[voiceExt],
		VoiceTime: voiceTime,
	})
	if err != nil {
		err = fmt.Errorf("发送语音消息失败: %w", err)
		return
	}

	return
}

func (r *Robot) MsgSendFile(req SendFileMessageRequest, file io.Reader, fileHeader *multipart.FileHeader) (*SendAppResponse, error) {
	// 1. 上传文件
	req.Wxid = r.WxID
	resp, err := r.Client.ToolsSendFile(req, file, fileHeader)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}
	if resp.CreateTime == nil {
		return nil, nil
	}
	// 2. 发送文件消息
	var fileXml FileMessageXml
	fileXml.Appmsg.AppID = ""
	fileXml.Appmsg.SDKVer = 0
	fileXml.Appmsg.Title = req.Filename
	fileXml.Appmsg.Type = 6
	fileXml.Appmsg.ShowType = 0
	fileXml.Appmsg.SoundType = 0
	fileXml.Appmsg.ContentAttr = 0
	fileXml.Appmsg.MD5 = req.FileMD5
	if resp.AppId != nil {
		fileXml.Appmsg.AppID = *resp.AppId
	}

	appAttach := AppAttach{}
	if resp.MediaId != nil {
		appAttach.AttachID = *resp.MediaId
	}
	appAttach.FileExt = strings.TrimPrefix(filepath.Ext(req.Filename), ".")
	appAttach.TotalLen = req.TotalLen
	fileXml.Appmsg.Attach = appAttach

	xmlBytes, err := xml.Marshal(fileXml.Appmsg)
	if err != nil {
		return nil, err
	}
	xmlStr := string(xmlBytes)

	appMessage, err := r.Client.SendApp(SendAppRequest{
		Wxid:   r.WxID,
		ToWxid: req.ToWxid,
		Xml:    xmlStr,
		Type:   6,
	})
	if err != nil {
		return nil, err
	}

	messageContentBytes, _ := xml.Marshal(fileXml)
	appMessage.Content = string(messageContentBytes)
	return &appMessage, nil
}

func (r *Robot) SendMusicMessage(toWxID string, songInfo SongInfo) (appMessage SendAppResponse, err error) {
	musicXmlPath := filepath.Join("xml", "music.xml")
	xmlTemplate, err := XmlFolder.ReadFile(musicXmlPath)
	if err != nil {
		err = fmt.Errorf("读取音乐XML模板失败: %w", err)
		return
	}

	// 使用模板引擎渲染XML
	tmpl, err := template.New("musicXml").Parse(string(xmlTemplate))
	if err != nil {
		err = fmt.Errorf("解析XML模板失败: %w", err)
		return
	}

	var renderedXml bytes.Buffer
	err = tmpl.Execute(&renderedXml, songInfo)
	if err != nil {
		err = fmt.Errorf("渲染XML模板失败: %w", err)
		return
	}

	// 发送音乐分享消息
	xmlStr := renderedXml.String()
	appMessage, err = r.Client.SendApp(SendAppRequest{
		Wxid:   r.WxID,
		ToWxid: toWxID,
		Xml:    xmlStr,
		Type:   3,
	})
	if err != nil {
		err = fmt.Errorf("发送音乐消息失败: %w", err)
		return
	}
	appMessage.Content = xmlStr
	return
}

func (r *Robot) SendChatHistoryMessage(toWxID string, message ChatHistoryMessage) (appMessage SendAppResponse, err error) {
	var xmlBytes []byte
	var xmlStr string
	xmlBytes, err = xml.MarshalIndent(message.AppMsg, "", "  ")
	if err != nil {
		return
	}
	xmlStr = string(xmlBytes)
	appMessage, err = r.Client.SendApp(SendAppRequest{
		Wxid:   r.WxID,
		ToWxid: toWxID,
		Xml:    xmlStr,
		Type:   19,
	})
	if err != nil {
		err = fmt.Errorf("发送聊天记录消息失败: %w", err)
		return
	}
	appMessage.Content = xmlStr
	return
}

func (r *Robot) SendEmoji(req SendEmojiRequest) (emojiMessage SendEmojiResponse, err error) {
	req.Wxid = r.WxID
	return r.Client.SendEmoji(req)
}

func (r *Robot) ShareLink(toWxID string, shareLinkInfo ShareLinkMessage) (shareLinkMessage ShareLinkResponse, xmlStr string, err error) {
	shareLinkInfo.AppID = ""
	shareLinkInfo.SDKVer = "1"
	shareLinkInfo.Type = 5
	var xmlBytes []byte
	xmlBytes, err = xml.MarshalIndent(shareLinkInfo, "", "  ")
	if err != nil {
		return
	}
	xmlStr = string(xmlBytes)
	shareLinkMessage, err = r.Client.ShareLink(ShareLinkRequest{
		ToWxid: toWxID,
		Wxid:   r.WxID,
		Type:   5,
		Xml:    xmlStr,
	})
	return
}

func (r *Robot) SendCDNFile(req SendCDNAttachmentRequest) (cdnFileMessage SendCDNFileResponse, err error) {
	req.Wxid = r.WxID
	return r.Client.SendCDNFile(req)
}

func (r *Robot) SendCDNImg(req SendCDNAttachmentRequest) (cdnImageMessage SendCDNImgResponse, err error) {
	req.Wxid = r.WxID
	return r.Client.SendCDNImg(req)
}

func (r *Robot) SendCDNVideo(req SendCDNAttachmentRequest) (cdnVideoMessage SendCDNVideoResponse, err error) {
	req.Wxid = r.WxID
	return r.Client.SendCDNVideo(req)
}

func (r *Robot) CheckLoginUuid(uuid string) (CheckUuid, error) {
	return r.Client.CheckLoginUuid(uuid)
}

func (r *Robot) LoginYPayVerificationcode(req VerificationCodeRequest) error {
	return r.Client.LoginYPayVerificationcode(req)
}

func (r *Robot) LoginData62Login(username, password string) (UnifyAuthResponse, error) {
	var data62 string
	if r.WxID != "" {
		var err error
		data62, err = r.Client.LoginGet62Data(r.WxID)
		if err != nil {
			return UnifyAuthResponse{}, err
		}
	}
	return r.Client.LoginData62SMSApply(Data62LoginRequest{
		UserName:   username,
		Password:   password,
		DeviceName: r.DeviceName,
		Data62:     data62,
	})
}

func (r *Robot) LoginData62SMSAgain(req LoginData62SMSAgainRequest) (string, error) {
	return r.Client.LoginData62SMSAgain(req)
}

func (r *Robot) LoginData62SMSVerify(req LoginData62SMSVerifyRequest) (string, error) {
	return r.Client.LoginData62SMSVerify(req)
}

func (r *Robot) LoginA16Data1(username, password string) (UnifyAuthResponse, error) {
	var a16 string
	if r.WxID == "" {
		return UnifyAuthResponse{}, errors.New("当前机器人还未成功登录过，不支持通过A16强行登录")
	}
	var err error
	a16, err = r.Client.LoginGetA16Data(r.WxID)
	if err != nil {
		return UnifyAuthResponse{}, err
	}
	return r.Client.LoginA16Data1(A16LoginRequest{
		UserName:   username,
		Password:   password,
		DeviceName: r.DeviceName,
		A16:        a16,
	})
}

func (r *Robot) Logout() error {
	if r.WxID == "" {
		return errors.New("您还未登陆")
	}
	return r.Client.Logout(r.WxID)
}

func (r *Robot) AutoHeartBeat() error {
	return r.Client.AutoHeartBeat(r.WxID)
}

func (r *Robot) CloseAutoHeartBeat() error {
	return r.Client.CloseAutoHeartBeat(r.WxID)
}

func (r *Robot) Heartbeat() error {
	return r.Client.Heartbeat(r.WxID)
}

func (r *Robot) FriendGetFriendstate(username string) (MMBizJsApiGetUserOpenIdResponse, error) {
	return r.Client.FriendGetFriendstate(r.WxID, username)
}

func (r *Robot) FriendSearch(req FriendSearchRequest) (SearchContactResponse, error) {
	req.Wxid = r.WxID
	return r.Client.FriendSearch(req)
}

func (r *Robot) FriendSendRequest(req FriendSendRequestParam) (VerifyUserResponse, error) {
	req.Wxid = r.WxID
	return r.Client.FriendSendRequest(req)
}

func (r *Robot) FriendSetRemarks(toWxid, remarks string) (OplogResponse, error) {
	return r.Client.FriendSetRemarks(r.WxID, toWxid, remarks)
}

func (r *Robot) GetContactList() ([]string, error) {
	return r.Client.GetContactList(r.WxID)
}

func (r *Robot) GetContactDetail(chatRoomID string, requestWxids []string) (GetContactResponse, error) {
	return r.Client.GetContactDetail(r.WxID, chatRoomID, requestWxids)
}

func (r *Robot) FriendPassVerify(req FriendPassVerifyRequest) (VerifyUserResponse, error) {
	req.Wxid = r.WxID
	return r.Client.FriendPassVerify(req)
}

func (r *Robot) FriendDelete(ToWxid string) (OplogResponse, error) {
	return r.Client.FriendDelete(r.WxID, ToWxid)
}

func (r *Robot) CreateChatRoom(contactIDs []string) (CreateChatRoomResponse, error) {
	if slices.Contains(contactIDs, r.WxID) {
		return CreateChatRoomResponse{}, errors.New("不能将自己添加到群聊中")
	}
	if len(contactIDs) < 2 {
		return CreateChatRoomResponse{}, errors.New("发起群聊至少需要2个成员")
	}
	chatRoom, err := r.Client.CreateChatRoom(r.WxID, contactIDs)
	if err != nil {
		return CreateChatRoomResponse{}, err
	}
	return chatRoom, nil
}

func (r *Robot) GroupAddChatRoomMember(chatRoomName string, contactIDs []string) error {
	_, err := r.Client.GroupAddChatRoomMember(r.WxID, chatRoomName, contactIDs)
	return err
}

func (r *Robot) GroupInviteChatRoomMember(chatRoomName string, contactIDs []string) error {
	_, err := r.Client.GroupInviteChatRoomMember(r.WxID, chatRoomName, contactIDs)
	return err
}

func (r *Robot) GroupConsentToJoin(Url string) (string, error) {
	return r.Client.GroupConsentToJoin(r.WxID, Url)
}

func (r *Robot) GetChatRoomMemberDetail(QID string) ([]ChatRoomMember, error) {
	return r.Client.GetChatRoomMemberDetail(r.WxID, QID)
}

func (r *Robot) GroupSetChatRoomName(QID, Content string) error {
	return r.Client.GroupSetChatRoomName(r.WxID, QID, Content)
}

func (r *Robot) GroupSetChatRoomRemarks(QID, Content string) error {
	return r.Client.GroupSetChatRoomRemarks(r.WxID, QID, Content)
}

func (r *Robot) GroupSetChatRoomAnnouncement(QID, Content string) error {
	return r.Client.GroupSetChatRoomAnnouncement(r.WxID, QID, Content)
}

func (r *Robot) GroupDelChatRoomMember(QID string, ToWxids []string) error {
	return r.Client.GroupDelChatRoomMember(r.WxID, QID, ToWxids)
}

func (r *Robot) GroupQuit(QID string) error {
	return r.Client.GroupQuit(r.WxID, QID)
}

func (r *Robot) DecodeTimelineObject(snsObject *SnsObject) {
	if snsObject != nil && snsObject.ObjectDesc != nil && snsObject.ObjectDesc.Buffer != nil {
		var timelineObject TimelineObject
		err := r.XmlDecoder(*snsObject.ObjectDesc.Buffer, &timelineObject)
		if err != nil {
			snsObject.TimelineObject = &TimelineObject{}
		} else {
			snsObject.TimelineObject = &timelineObject
		}
	}
}

func (r *Robot) FriendCircleComment(req FriendCircleCommentRequest) (SnsCommentResponse, error) {
	req.Wxid = r.WxID
	resp, err := r.Client.FriendCircleComment(req)
	if err != nil {
		return SnsCommentResponse{}, err
	}
	r.DecodeTimelineObject(resp.SnsObject)
	return resp, nil
}

func (r *Robot) FriendCircleGetDetail(req FriendCircleGetDetailRequest) (SnsUserPageResponse, error) {
	req.Wxid = r.WxID
	resp, err := r.Client.FriendCircleGetDetail(req)
	if err != nil {
		return SnsUserPageResponse{}, err
	}
	for _, snsObject := range resp.ObjectList {
		r.DecodeTimelineObject(snsObject)
	}
	return resp, nil
}

func (r *Robot) FriendCircleGetIdDetail(req FriendCircleGetIdDetailRequest) (SnsObjectDetailResponse, error) {
	req.Wxid = r.WxID
	resp, err := r.Client.FriendCircleGetIdDetail(req)
	if err != nil {
		return SnsObjectDetailResponse{}, err
	}
	r.DecodeTimelineObject(resp.Object)
	return resp, nil
}

func (r *Robot) FriendCircleGetList(Fristpagemd5 string, Maxid string) (GetListResponse, error) {
	data, err := r.Client.FriendCircleGetList(r.WxID, Fristpagemd5, Maxid)
	if err != nil {
		return GetListResponse{}, err
	}
	for index, snsObject := range data.ObjectList {
		if snsObject.ObjectDesc.Buffer != nil {
			var timelineObject TimelineObject
			err := r.XmlDecoder(*snsObject.ObjectDesc.Buffer, &timelineObject)
			if err != nil {
				data.ObjectList[index].TimelineObject = &TimelineObject{}
			} else {
				data.ObjectList[index].TimelineObject = &timelineObject
			}
		}
	}
	return data, nil
}

func (r *Robot) FriendCircleDownFriendCircleMedia(Url, Key string) (string, error) {
	base64Url := base64.StdEncoding.EncodeToString([]byte(Url))
	return r.Client.FriendCircleDownFriendCircleMedia(r.WxID, base64Url, Key)
}

func (r *Robot) FriendCircleUpload(mediaBytes []byte) (FriendCircleUploadResponse, error) {
	base64Str := base64.StdEncoding.EncodeToString(mediaBytes)
	return r.Client.FriendCircleUpload(r.WxID, base64Str)
}

func (r *Robot) FriendCircleCdnSnsUploadVideo(thumbBytes, videoBytes []byte) (CdnSnsVideoUploadResponse, error) {
	return r.Client.FriendCircleCdnSnsUploadVideo(FriendCircleCdnSnsUploadVideoRequest{
		Wxid:      r.WxID,
		ThumbData: base64.StdEncoding.EncodeToString(thumbBytes),
		VideoData: base64.StdEncoding.EncodeToString(videoBytes),
	})
}

func (r *Robot) FriendCircleMessages(req FriendCircleMessagesRequest) (FriendCircleMessagesResponse, error) {
	req.Wxid = r.WxID
	return r.Client.FriendCircleMessages(req)
}

func (r *Robot) FriendCircleMmSnsSync(synckey string) (SyncMessage, error) {
	return r.Client.FriendCircleMmSnsSync(r.WxID, synckey)
}

func (r *Robot) FriendCircleOperation(req FriendCircleOperationRequest) (SnsObjectOpResponse, error) {
	req.Wxid = r.WxID
	return r.Client.FriendCircleOperation(req)
}

func (r *Robot) FriendCirclePrivacySettings(req FriendCirclePrivacySettingsRequest) (OplogResponse, error) {
	req.Wxid = r.WxID
	return r.Client.FriendCirclePrivacySettings(req)
}

func (r *Robot) WxappQrcodeAuthLogin(URL string) error {
	return r.Client.WxappQrcodeAuthLogin(r.WxID, URL)
}

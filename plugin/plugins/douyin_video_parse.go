package plugins

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	_ "image/png"
	"mime/multipart"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/webp"

	"wechat-robot-client/dto"
	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/pkg/robot"
	"wechat-robot-client/vars"
)

type VideoParseResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Author     string   `json:"author"`
		Avatar     string   `json:"avatar"`
		Title      string   `json:"title"`
		Desc       string   `json:"desc"`
		Digg       int32    `json:"digg"`
		Comment    int32    `json:"comment"`
		Play       int32    `json:"play"`
		CreateTime int64    `json:"create_time"`
		Cover      string   `json:"cover"`
		URL        string   `json:"url"`
		Images     []string `json:"images"`
		MusicURL   string   `json:"music_url"`
	} `json:"data"`
}

type DouyinVideoParsePlugin struct{}

func NewDouyinVideoParsePlugin() plugin.MessageHandler {
	return &DouyinVideoParsePlugin{}
}

func (p *DouyinVideoParsePlugin) GetName() string {
	return "DouyinVideoParse"
}

func (p *DouyinVideoParsePlugin) GetLabels() []string {
	return []string{"text", "douyin"}
}

func (p *DouyinVideoParsePlugin) PreAction(ctx *plugin.MessageContext) bool {
	if ctx.Message.IsChatRoom {
		next := NewChatRoomCommonPlugin().PreAction(ctx)
		if !next {
			return false
		}
		if !ctx.Settings.IsShortVideoParsingEnabled() {
			return false
		}
	}
	return true
}

func (p *DouyinVideoParsePlugin) PostAction(ctx *plugin.MessageContext) {

}

func (p *DouyinVideoParsePlugin) Match(ctx *plugin.MessageContext) bool {
	return strings.Contains(ctx.Message.Content, "https://v.douyin.com")
}

func (p *DouyinVideoParsePlugin) Run(ctx *plugin.MessageContext) {
	if !p.PreAction(ctx) {
		return
	}

	re := regexp.MustCompile(`https://[^\s]+`)
	matches := re.FindAllString(ctx.Message.Content, -1)
	if len(matches) == 0 {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "未找到抖音链接")
		return
	}

	// 获取第一个匹配的链接
	douyinURL := matches[0]

	var respData VideoParseResponse
	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{
			"key": vars.ThirdPartyApiKey,
			"url": douyinURL,
		}).
		SetResult(&respData).
		Post("https://api.pearktrue.cn/api/video/api.php")
	if err != nil {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, err.Error())
		return
	}
	if resp.StatusCode() != http.StatusOK {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, http.StatusText(resp.StatusCode()))
		return
	}

	if respData.Data.URL != "" {
		shareLink := robot.ShareLinkMessage{
			Title:    fmt.Sprintf("抖音视频解析成功 - %s", respData.Data.Author),
			Des:      respData.Data.Title,
			Url:      respData.Data.URL,
			ThumbUrl: robot.CDATAString("https://mmbiz.qpic.cn/mmbiz_png/NbW0ZIUM8lVHoUbjXw2YbYXbNJDtUH7Sbkibm9Qwo9FhAiaEFG4jY3Q2MEleRpiaWDyDv8BZUfR85AW3kG4ib6DyAw/640?wx_fmt=png"),
		}
		if respData.Data.Desc != "" {
			shareLink.Des = respData.Data.Desc
		}

		_ = ctx.MessageService.ShareLink(ctx.Message.FromWxID, shareLink)
		_ = ctx.MessageService.SendVideoMessageByRemoteURL(ctx.Message.FromWxID, respData.Data.URL)

		return
	}

	if len(respData.Data.Images) > 0 {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, fmt.Sprintf("抖音图片解析成功\n作者: %s\n标题: %s\n\n%d张图片正在发送中...", respData.Data.Author, respData.Data.Title, len(respData.Data.Images)))

		imageURLs := respData.Data.Images
		batchSize := 20
		for i := 0; i < len(imageURLs); i += batchSize {
			end := i + batchSize
			end = min(end, len(imageURLs))

			mergedImage, err := mergeImagesVertical(imageURLs[i:end])
			if err != nil {
				ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, fmt.Sprintf("拼接失败(批次 %d-%d): %v", i+1, end, err))
				continue
			}
			err = sendMergedImage(ctx, mergedImage)
			if err != nil {
				ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, fmt.Sprintf("发送图片失败: %v", err))
			}
		}
		return
	}

	ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "解析失败，可能是链接已失效或格式不正确")
}

func mergeImagesVertical(imageURLs []string) ([]byte, error) {
	if len(imageURLs) == 0 {
		return nil, fmt.Errorf("图片地址为空")
	}

	client := resty.New()
	images := make([]image.Image, 0, len(imageURLs))
	maxWidth := 0

	for _, imageURL := range imageURLs {
		resp, err := client.R().SetDoNotParseResponse(true).Get(imageURL)
		if err != nil {
			return nil, fmt.Errorf("下载图片失败: %w", err)
		}
		if resp.StatusCode() != http.StatusOK {
			resp.RawBody().Close()
			return nil, fmt.Errorf("下载图片失败，HTTP状态码: %d", resp.StatusCode())
		}

		img, _, err := image.Decode(resp.RawBody())
		resp.RawBody().Close()
		if err != nil {
			return nil, fmt.Errorf("解析图片失败: %w", err)
		}

		bounds := img.Bounds()
		width := bounds.Dx()
		if width > maxWidth {
			maxWidth = width
		}
		images = append(images, img)
	}

	if maxWidth == 0 || len(images) == 0 {
		return nil, fmt.Errorf("图片尺寸无效")
	}

	totalHeight := 0
	for _, img := range images {
		width := img.Bounds().Dx()
		height := img.Bounds().Dy()
		// 等比缩放计算高度
		newHeight := int(float64(height) * float64(maxWidth) / float64(width))
		totalHeight += newHeight
	}

	canvas := image.NewRGBA(image.Rect(0, 0, maxWidth, totalHeight))
	draw.Draw(canvas, canvas.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)

	currentY := 0
	for _, img := range images {
		width := img.Bounds().Dx()
		height := img.Bounds().Dy()
		newHeight := int(float64(height) * float64(maxWidth) / float64(width))

		dstRect := image.Rect(0, currentY, maxWidth, currentY+newHeight)
		// 使用高质量缩放
		xdraw.CatmullRom.Scale(canvas, dstRect, img, img.Bounds(), xdraw.Over, nil)
		currentY += newHeight
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, canvas, &jpeg.Options{Quality: 80}); err != nil {
		return nil, fmt.Errorf("图片编码失败: %w", err)
	}

	return buf.Bytes(), nil
}

func sendMergedImage(ctx *plugin.MessageContext, imageData []byte) error {
	contentLength := int64(len(imageData))
	if contentLength == 0 {
		return fmt.Errorf("图片数据为空")
	}

	fmt.Printf("抖音图片合并后大小: %dMB\n", contentLength/1024/1024)

	clientImgId := fmt.Sprintf("%v_%v", vars.RobotRuntime.WxID, time.Now().UnixNano())
	chunkSize := vars.UploadImageChunkSize
	totalChunks := int((contentLength + chunkSize - 1) / chunkSize)

	for chunkIndex := range totalChunks {
		start := int64(chunkIndex) * chunkSize
		end := min(start+chunkSize, contentLength)

		chunkData := imageData[start:end]
		req := dto.SendImageMessageRequest{
			ToWxid:      ctx.Message.FromWxID,
			ClientImgId: clientImgId,
			FileSize:    contentLength,
			ChunkIndex:  int64(chunkIndex),
			TotalChunks: int64(totalChunks),
		}

		chunkReader := bytes.NewReader(chunkData)
		chunkHeader := &multipart.FileHeader{
			Filename: fmt.Sprintf("chunk_%d", chunkIndex),
			Size:     int64(len(chunkData)),
		}

		if _, err := ctx.MessageService.SendImageMessageStream(context.Background(), req, chunkReader, chunkHeader); err != nil {
			return err
		}
	}

	return nil
}

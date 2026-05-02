package plugins

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	_ "image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/webp"

	"wechat-robot-client/dto"
	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/pkg/robot"
	"wechat-robot-client/utils"
	"wechat-robot-client/vars"
)

type VideoParseResponse struct {
	Code int            `json:"code"`
	Msg  string         `json:"msg"`
	Data VideoParseData `json:"data"`
}

type VideoParseData struct {
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
}

const douyinUserAgent = "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1"

var (
	douyinRouterDataRegexp  = regexp.MustCompile(`(?s)window\._ROUTER_DATA\s*=\s*({.*?})\s*</script>`)
	douyinLegacyPlayRegexp  = regexp.MustCompile(`"play_addr":\s*\{\s*"uri":\s*"[^"]*",\s*"url_list":\s*\[([^\]]*)\]`)
	douyinLegacyCoverRegexp = regexp.MustCompile(`"cover":\s*\{\s*"url_list":\s*\[\s*"([^"]+)"`)
)

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
	douyinURL := matches[0]

	respData, err := parseDouyinVideo(douyinURL)
	if err != nil {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, fmt.Sprintf("解析失败: %v", err))
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
		err = ctx.MessageService.SendVideoMessageByRemoteURL(ctx.Message.FromWxID, respData.Data.URL)
		if err != nil {
			ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, fmt.Sprintf("发送抖音视频失败: %v", err.Error()))
		}

		return
	}

	if len(respData.Data.Images) > 0 {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, fmt.Sprintf("抖音图片解析成功\n作者: %s\n标题: %s\n\n%d张图片正在发送中...", respData.Data.Author, respData.Data.Title, len(respData.Data.Images)))

		if respData.Data.MusicURL != "" {
			go func(musicURL, title, author string) {
				var err error
				if isAudioURL(musicURL) {
					err = sendMusicMessageByURL(ctx, musicURL, author)
				} else {
					err = sendFileByRemoteURL(ctx, musicURL)
				}
				if err != nil {
					ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, fmt.Sprintf("发送抖音音频失败: %v", err))
				}
			}(respData.Data.MusicURL, respData.Data.Title, respData.Data.Author)
		}

		imageURLs := respData.Data.Images
		batchSize := 20
		for i := 0; i < len(imageURLs); i += batchSize {
			end := i + batchSize
			end = min(end, len(imageURLs))

			mergedImage, err := mergeImagesVertical(ctx, imageURLs[i:end])
			if err != nil {
				if isImageTooLargeError(err) {
					p.sendImagesInSmallerBatches(ctx, imageURLs[i:end], 10)
					continue
				}
				ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, fmt.Sprintf("拼接失败(批次 %d-%d): %v", i+1, end, err))
				continue
			}
			if len(mergedImage) == 0 {
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

func parseDouyinVideo(rawURL string) (VideoParseResponse, error) {
	resolvedURL, err := resolveDouyinRedirect(rawURL)
	if err != nil {
		return VideoParseResponse{}, err
	}

	htmlContent, err := fetchDouyinPageHTML(resolvedURL)
	if err != nil {
		return VideoParseResponse{}, err
	}

	data, err := parseDouyinPageHTML(htmlContent)
	if err != nil {
		return VideoParseResponse{}, err
	}
	return VideoParseResponse{Code: http.StatusOK, Data: data}, nil
}

func resolveDouyinRedirect(rawURL string) (string, error) {
	client := &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, rawURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建抖音短链请求失败: %w", err)
	}
	req.Header.Set("User-Agent", douyinUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("解析抖音短链失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusMultipleChoices && resp.StatusCode < http.StatusBadRequest {
		location, err := resp.Location()
		if err != nil {
			return rawURL, nil
		}
		return location.String(), nil
	}
	return resp.Request.URL.String(), nil
}

func fetchDouyinPageHTML(pageURL string) (string, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, pageURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建抖音页面请求失败: %w", err)
	}
	req.Header.Set("User-Agent", douyinUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("获取抖音页面失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("获取抖音页面失败，状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取抖音页面失败: %w", err)
	}
	if len(body) == 0 {
		return "", fmt.Errorf("抖音页面内容为空")
	}
	return string(body), nil
}

func parseDouyinPageHTML(htmlContent string) (VideoParseData, error) {
	if item, ok := extractDouyinAwemeItem(htmlContent); ok {
		if note, ok := parseDouyinNoteItem(item); ok {
			return note, nil
		}
		if video, ok := parseDouyinVideoItem(item); ok {
			return video, nil
		}
	}

	if video, ok := parseDouyinLegacyVideo(htmlContent); ok {
		return video, nil
	}
	return VideoParseData{}, fmt.Errorf("未找到可解析的抖音图文或视频内容")
}

func extractDouyinAwemeItem(htmlContent string) (map[string]any, bool) {
	match := douyinRouterDataRegexp.FindStringSubmatch(htmlContent)
	if len(match) < 2 {
		return nil, false
	}

	var routerData map[string]any
	if err := json.Unmarshal([]byte(match[1]), &routerData); err != nil {
		log.Printf("解析抖音 _ROUTER_DATA 失败: %v\n", err)
		return nil, false
	}

	loaderData, ok := routerData["loaderData"].(map[string]any)
	if !ok {
		return nil, false
	}
	for _, pageDataValue := range loaderData {
		pageData, ok := pageDataValue.(map[string]any)
		if !ok {
			continue
		}
		videoInfoRes, ok := pageData["videoInfoRes"].(map[string]any)
		if !ok {
			continue
		}
		itemList, ok := videoInfoRes["item_list"].([]any)
		if !ok || len(itemList) == 0 {
			continue
		}
		item, ok := itemList[0].(map[string]any)
		if ok {
			return item, true
		}
	}
	return nil, false
}

func parseDouyinNoteItem(item map[string]any) (VideoParseData, bool) {
	imageURLGroups := pickDouyinImageURLGroups(item)
	if len(imageURLGroups) == 0 {
		return VideoParseData{}, false
	}

	imageURLs := make([]string, 0, len(imageURLGroups))
	for _, group := range imageURLGroups {
		imageURLs = append(imageURLs, group[0])
	}
	desc := cleanDouyinText(stringFromAny(item["desc"]))
	return VideoParseData{
		Author:   cleanDouyinText(nestedString(item, "author", "nickname")),
		Title:    desc,
		Desc:     desc,
		Images:   imageURLs,
		MusicURL: pickFirstDouyinURL(nestedStringList(item, "music", "play_url", "url_list")),
	}, true
}

func pickDouyinImageURLGroups(item map[string]any) [][]string {
	imageList := listFromAny(item["images"])
	if len(imageList) == 0 {
		imageList = listFromAny(item["image_infos"])
	}

	imageURLGroups := make([][]string, 0, len(imageList))
	seenGroups := make(map[string]bool)
	for _, imageValue := range imageList {
		imageInfo, ok := imageValue.(map[string]any)
		if !ok {
			continue
		}

		candidates := make([]string, 0)
		seenURLs := make(map[string]bool)
		for _, imageURL := range stringListFromAny(imageInfo["url_list"]) {
			if !strings.HasPrefix(imageURL, "http") {
				continue
			}
			decodedURL := html.UnescapeString(imageURL)
			if seenURLs[decodedURL] {
				continue
			}
			candidates = append(candidates, decodedURL)
			seenURLs[decodedURL] = true
		}

		groupKey := strings.Join(candidates, "\x00")
		if len(candidates) > 0 && !seenGroups[groupKey] {
			imageURLGroups = append(imageURLGroups, candidates)
			seenGroups[groupKey] = true
		}
	}
	return imageURLGroups
}

func parseDouyinVideoItem(item map[string]any) (VideoParseData, bool) {
	video, ok := item["video"].(map[string]any)
	if !ok {
		return VideoParseData{}, false
	}
	if duration, ok := numberFromAny(video["duration"]); ok && duration == 0 {
		return VideoParseData{}, false
	}

	videoURL := pickDouyinVideoURL(nestedStringList(video, "play_addr", "url_list"))
	if videoURL == "" {
		return VideoParseData{}, false
	}

	desc := cleanDouyinText(stringFromAny(item["desc"]))
	return VideoParseData{
		Author:   cleanDouyinText(nestedString(item, "author", "nickname")),
		Title:    desc,
		Desc:     desc,
		Cover:    pickFirstDouyinURL(nestedStringList(video, "cover", "url_list")),
		URL:      videoURL,
		MusicURL: pickFirstDouyinURL(nestedStringList(item, "music", "play_url", "url_list")),
	}, true
}

func parseDouyinLegacyVideo(htmlContent string) (VideoParseData, bool) {
	match := douyinLegacyPlayRegexp.FindStringSubmatch(htmlContent)
	if len(match) < 2 {
		return VideoParseData{}, false
	}

	urls := make([]string, 0)
	if err := json.Unmarshal([]byte("["+match[1]+"]"), &urls); err != nil {
		for _, rawURL := range strings.Split(match[1], ",") {
			trimmedURL := strings.Trim(strings.TrimSpace(rawURL), `"`)
			if trimmedURL != "" {
				urls = append(urls, trimmedURL)
			}
		}
	}

	videoURL := pickDouyinVideoURL(urls)
	if videoURL == "" {
		return VideoParseData{}, false
	}

	cover := ""
	if coverMatch := douyinLegacyCoverRegexp.FindStringSubmatch(htmlContent); len(coverMatch) > 1 {
		cover = decodeDouyinEscapedValue(coverMatch[1])
	}

	desc := matchDouyinJSONString(htmlContent, "desc")
	return VideoParseData{
		Author: matchDouyinJSONString(htmlContent, "nickname"),
		Title:  desc,
		Desc:   desc,
		Cover:  cover,
		URL:    videoURL,
	}, true
}

func pickDouyinVideoURL(urls []string) string {
	decodedURLs := make([]string, 0, len(urls))
	for _, rawURL := range urls {
		if rawURL == "" {
			continue
		}
		decodedURL := strings.ReplaceAll(decodeDouyinEscapedValue(rawURL), "playwm", "play")
		decodedURLs = append(decodedURLs, decodedURL)
	}
	for _, decodedURL := range decodedURLs {
		if strings.Contains(decodedURL, "aweme.snssdk.com") {
			return decodedURL
		}
	}
	if len(decodedURLs) > 0 {
		return decodedURLs[0]
	}
	return ""
}

func pickFirstDouyinURL(urls []string) string {
	for _, rawURL := range urls {
		if rawURL != "" {
			return decodeDouyinEscapedValue(rawURL)
		}
	}
	return ""
}

func matchDouyinJSONString(text string, key string) string {
	pattern := regexp.MustCompile(fmt.Sprintf(`"%s":\s*"([^"]*)"`, regexp.QuoteMeta(key)))
	match := pattern.FindStringSubmatch(text)
	if len(match) < 2 {
		return ""
	}
	return cleanDouyinText(decodeDouyinEscapedValue(match[1]))
}

func decodeDouyinEscapedValue(value string) string {
	decodedValue := html.UnescapeString(value)
	if strings.Contains(decodedValue, `\`) {
		var unquotedValue string
		if err := json.Unmarshal([]byte(`"`+strings.ReplaceAll(decodedValue, `"`, `\"`)+`"`), &unquotedValue); err == nil {
			decodedValue = unquotedValue
		}
	}
	return html.UnescapeString(decodedValue)
}

func cleanDouyinText(value string) string {
	return strings.TrimSpace(html.UnescapeString(value))
}

func nestedString(root map[string]any, keys ...string) string {
	current := any(root)
	for _, key := range keys {
		currentMap, ok := current.(map[string]any)
		if !ok {
			return ""
		}
		current = currentMap[key]
	}
	return stringFromAny(current)
}

func nestedStringList(root map[string]any, keys ...string) []string {
	current := any(root)
	for _, key := range keys {
		currentMap, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current = currentMap[key]
	}
	return stringListFromAny(current)
}

func stringFromAny(value any) string {
	if value == nil {
		return ""
	}
	if str, ok := value.(string); ok {
		return str
	}
	return fmt.Sprint(value)
}

func listFromAny(value any) []any {
	if list, ok := value.([]any); ok {
		return list
	}
	return nil
}

func stringListFromAny(value any) []string {
	list, ok := value.([]any)
	if !ok {
		return nil
	}
	stringsList := make([]string, 0, len(list))
	for _, item := range list {
		if str, ok := item.(string); ok {
			stringsList = append(stringsList, str)
		}
	}
	return stringsList
}

func numberFromAny(value any) (float64, bool) {
	switch number := value.(type) {
	case float64:
		return number, true
	case int:
		return float64(number), true
	case int64:
		return float64(number), true
	default:
		return 0, false
	}
}

func (p *DouyinVideoParsePlugin) sendImagesInSmallerBatches(ctx *plugin.MessageContext, imageURLs []string, batchSize int) {
	if batchSize <= 0 {
		return
	}
	for i := 0; i < len(imageURLs); i += batchSize {
		end := i + batchSize
		end = min(end, len(imageURLs))

		mergedImage, err := mergeImagesVertical(ctx, imageURLs[i:end])
		if err != nil {
			ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, fmt.Sprintf("拼接失败(降级批次 %d-%d): %v", i+1, end, err))
			continue
		}
		if len(mergedImage) == 0 {
			continue
		}
		err = sendMergedImage(ctx, mergedImage)
		if err != nil {
			ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, fmt.Sprintf("发送图片失败: %v", err))
		}
	}
}

func mergeImagesVertical(ctx *plugin.MessageContext, imageURLs []string) ([]byte, error) {
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

		bodyData := new(bytes.Buffer)
		_, err = bodyData.ReadFrom(resp.RawBody())
		resp.RawBody().Close()
		if err != nil {
			return nil, fmt.Errorf("读取响应体失败: %w", err)
		}

		if utils.IsVideo(bodyData.Bytes()) {
			log.Printf("%s 解析到视频，跳过合并，直接发送视频消息\n", imageURL)
			go func(toWxID, _imageURL string) {
				err2 := ctx.MessageService.SendVideoMessageByRemoteURL(toWxID, _imageURL)
				if err2 != nil {
					ctx.MessageService.SendTextMessage(toWxID, fmt.Sprintf("发送抖音视频失败: %v", err2.Error()))
				}
			}(ctx.Message.FromWxID, imageURL)
			continue
		}

		img, _, err := image.Decode(bytes.NewReader(bodyData.Bytes()))
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

	// 有可能全是视频
	if maxWidth == 0 || len(images) == 0 {
		return nil, nil
	}

	totalHeight := 0
	for _, img := range images {
		width := img.Bounds().Dx()
		height := img.Bounds().Dy()
		// 等比缩放计算高度
		newHeight := int(float64(height) * float64(maxWidth) / float64(width))
		totalHeight += newHeight
	}
	if maxWidth > jpegMaxDimension || totalHeight > jpegMaxDimension {
		return nil, fmt.Errorf("image is too large to encode")
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

const jpegMaxDimension = 65535

var audioExtensions = map[string]bool{
	".mp3":  true,
	".m4a":  true,
	".aac":  true,
	".ogg":  true,
	".flac": true,
	".wav":  true,
	".wma":  true,
	".amr":  true,
}

func isAudioURL(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	ext := strings.ToLower(path.Ext(parsed.Path))
	return audioExtensions[ext]
}

func sendMusicMessageByURL(ctx *plugin.MessageContext, musicURL, author string) error {
	const (
		appID    = "wx8dd6ecd81906fd84"
		coverURL = "https://uranus-houhou.oss-cn-beijing.aliyuncs.com/douyin.png"
	)
	songInfo := robot.SongInfo{}
	songInfo.FromUsername = vars.RobotRuntime.WxID
	songInfo.AppID = appID
	songInfo.Title = "抖音解析背景音乐"
	songInfo.Singer = author
	songInfo.Url = musicURL
	songInfo.MusicUrl = musicURL
	songInfo.CoverUrl = coverURL
	_, err := vars.RobotRuntime.SendMusicMessage(ctx.Message.FromWxID, songInfo)
	return err
}

func isImageTooLargeError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "image is too large to encode")
}

func sendMergedImage(ctx *plugin.MessageContext, imageData []byte) error {
	contentLength := int64(len(imageData))
	if contentLength == 0 {
		return nil
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

func sendFileByRemoteURL(ctx *plugin.MessageContext, fileURL string) error {
	resp, err := resty.New().R().SetDoNotParseResponse(true).Get(fileURL)
	if err != nil {
		return fmt.Errorf("下载文件失败: %w", err)
	}
	defer resp.RawBody().Close()

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("下载文件失败，HTTP状态码: %d", resp.StatusCode())
	}

	fileData, err := io.ReadAll(resp.RawBody())
	if err != nil {
		return fmt.Errorf("读取文件数据失败: %w", err)
	}
	if len(fileData) == 0 {
		return fmt.Errorf("文件数据为空")
	}

	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return fmt.Errorf("解析文件URL失败: %w", err)
	}
	filename := path.Base(parsedURL.Path)
	if filename == "" || filename == "/" || filename == "." {
		filename = "douyin_music.mp3"
	}

	fileMD5Bytes := md5.Sum(fileData)
	fileHash := hex.EncodeToString(fileMD5Bytes[:])
	fileSize := int64(len(fileData))
	chunkSize := vars.UploadFileChunkSize
	if chunkSize <= 0 {
		chunkSize = 200 * 1000
	}
	totalChunks := (fileSize + chunkSize - 1) / chunkSize
	clientAppDataID := fmt.Sprintf("%v_%v", vars.RobotRuntime.WxID, time.Now().UnixNano())

	for chunkIndex := range totalChunks {
		start := int64(chunkIndex) * chunkSize
		end := min(start+chunkSize, fileSize)
		chunkData := fileData[start:end]

		req := dto.SendFileMessageRequest{
			ToWxid:          ctx.Message.FromWxID,
			ClientAppDataId: clientAppDataID,
			Filename:        filename,
			FileHash:        fileHash,
			FileSize:        fileSize,
			ChunkIndex:      int64(chunkIndex),
			TotalChunks:     totalChunks,
		}

		chunkReader := bytes.NewReader(chunkData)
		chunkHeader := &multipart.FileHeader{
			Filename: filename,
			Size:     int64(len(chunkData)),
		}

		if err = ctx.MessageService.SendFileMessage(context.Background(), req, chunkReader, chunkHeader); err != nil {
			if strings.Contains(err.Error(), "context canceled") || strings.Contains(err.Error(), "context deadline exceeded") {
				return fmt.Errorf("发送文件超时")
			}
			return err
		}
	}

	return nil
}

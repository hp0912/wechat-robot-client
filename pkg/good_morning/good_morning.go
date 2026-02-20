package good_morning

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"time"
	"wechat-robot-client/dto"

	"github.com/chromedp/chromedp"
)

func Draw(dailyWords string, summary dto.ChatRoomSummary) (io.Reader, error) {
	indexTplBytes, err := Assets.ReadFile("assets/index.html")
	if err != nil {
		return nil, fmt.Errorf("read index template: %w", err)
	}
	backgroundBytes, err := Assets.ReadFile("assets/background.jpg")
	if err != nil {
		return nil, fmt.Errorf("read background: %w", err)
	}

	tpl, err := template.New("index.html").Parse(string(indexTplBytes))
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}

	data := struct {
		DailyWords string
		Summary    dto.ChatRoomSummary
	}{
		DailyWords: dailyWords,
		Summary:    summary,
	}

	var rendered bytes.Buffer
	if err := tpl.Execute(&rendered, data); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "good-morning-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	indexPath := filepath.Join(tmpDir, "index.html")
	bgPath := filepath.Join(tmpDir, "background.jpg")
	if err := os.WriteFile(indexPath, rendered.Bytes(), 0644); err != nil {
		return nil, fmt.Errorf("write rendered html: %w", err)
	}
	if err := os.WriteFile(bgPath, backgroundBytes, 0644); err != nil {
		return nil, fmt.Errorf("write background: %w", err)
	}

	fileURL := (&url.URL{Scheme: "file", Path: filepath.ToSlash(indexPath)}).String()

	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), allocOpts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var png []byte
	err = chromedp.Run(ctx,
		chromedp.EmulateViewport(900, 600),
		chromedp.Navigate(fileURL),
		chromedp.WaitReady(".container", chromedp.ByQuery),
		chromedp.Sleep(150*time.Millisecond),
		chromedp.Screenshot(".container", &png, chromedp.NodeVisible, chromedp.ByQuery),
	)
	if err != nil {
		return nil, fmt.Errorf("chromedp screenshot: %w", err)
	}

	return bytes.NewReader(png), nil
}

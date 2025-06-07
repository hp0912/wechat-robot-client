package good_morning

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"path/filepath"
	"unicode/utf8"
	"wechat-robot-client/dto"

	"image/color"
	"image/draw"
	"image/png"

	"github.com/golang/freetype"
)

func Draw(dailyWords string, summary dto.ChatRoomSummary) (io.Reader, error) {
	// 打开背景图片
	bgFilePath := filepath.Join("assets", "background.png")
	bgFile, err := Assets.ReadFile(bgFilePath)
	if err != nil {
		return nil, err
	}
	// 解码
	bgFileImage, err := png.Decode(bytes.NewReader(bgFile))
	if err != nil {
		return nil, err
	}
	// 新建一张和模板文件一样大小的画布
	newBgImage := image.NewRGBA(bgFileImage.Bounds())
	// 将模板图片画到新建的画布上
	draw.Draw(newBgImage, bgFileImage.Bounds(), bgFileImage, bgFileImage.Bounds().Min, draw.Over)
	// 加载字体文件  这里我们加载两种字体文件
	fontPath := filepath.Join("assets", "simkai.ttf")
	fontBytes, err := Assets.ReadFile(fontPath)
	if err != nil {
		return nil, err
	}
	fontKai, err := freetype.ParseFont(fontBytes) // 解析字体文件
	if err != nil {
		return nil, err
	}
	// 向图片中写入文字
	// 在写入之前有一些准备工作
	content := freetype.NewContext()
	content.SetClip(newBgImage.Bounds())
	content.SetDst(newBgImage)
	content.SetSrc(image.Black) // 设置字体颜色
	content.SetDPI(72)          // 设置字体分辨率

	content.SetFontSize(24)  // 设置字体大小
	content.SetFont(fontKai) // 设置字体样式，就是我们上面加载的字体

	drawLeft := 16
	drawTop := 30

	// 参数1：要写入的文字
	// 参数2：文字坐标
	txt := fmt.Sprintf("%d年%d月%d日  %s", summary.Year, summary.Month, summary.Date, summary.Week)
	content.DrawString(txt, freetype.Pt(drawLeft, drawTop))

	content.SetSrc(image.Opaque)
	content.SetFontSize(18) // 设置字体大小
	drawTop += 44
	content.DrawString("早上好，", freetype.Pt(drawLeft, drawTop))

	drawTop += 30
	drawContent := "群内成员一共"
	content.DrawString(drawContent, freetype.Pt(drawLeft, drawTop))

	content.SetSrc(image.NewUniform(color.RGBA{R: 237, G: 39, B: 90, A: 255})) // 设置字体颜色
	drawLeft += (utf8.RuneCountInString(drawContent) * 18)
	drawContent = fmt.Sprintf("%d", summary.MemberTotalCount)
	content.DrawString(drawContent, freetype.Pt(drawLeft, drawTop))

	content.SetSrc(image.Opaque)
	drawLeft += (len(drawContent) * 10)
	drawContent = "名，"
	content.DrawString(drawContent, freetype.Pt(drawLeft, drawTop))

	drawLeft = 16
	drawTop += 30
	if summary.MemberJoinCount == 0 && summary.MemberLeaveCount == 0 {
		drawContent = "没有人加入，也没有人离开，"
		content.DrawString(drawContent, freetype.Pt(drawLeft, drawTop))
	} else if summary.MemberJoinCount == 0 {
		drawContent = "没有人加入，有"
		content.DrawString(drawContent, freetype.Pt(drawLeft, drawTop))

		drawLeft += (utf8.RuneCountInString(drawContent) * 18)
		content.SetSrc(image.NewUniform(color.RGBA{R: 237, G: 39, B: 90, A: 255})) // 设置字体颜色
		drawContent = fmt.Sprintf("%d", summary.MemberLeaveCount)
		content.DrawString(drawContent, freetype.Pt(drawLeft, drawTop))

		drawLeft += (len(drawContent) * 10)
		content.SetSrc(image.Opaque)
		drawContent = "人离开了我们，"
		content.DrawString(drawContent, freetype.Pt(drawLeft, drawTop))
	} else if summary.MemberLeaveCount == 0 {
		drawContent = "有"
		content.DrawString(drawContent, freetype.Pt(drawLeft, drawTop))

		drawLeft += (utf8.RuneCountInString(drawContent) * 18)
		content.SetSrc(image.NewUniform(color.RGBA{R: 237, G: 39, B: 90, A: 255})) // 设置字体颜色
		drawContent = fmt.Sprintf("%d", summary.MemberJoinCount)
		content.DrawString(drawContent, freetype.Pt(drawLeft, drawTop))

		drawLeft += (len(drawContent) * 10)
		content.SetSrc(image.Opaque)
		drawContent = "人加入，没有人离开，"
		content.DrawString(drawContent, freetype.Pt(drawLeft, drawTop))
	} else {
		drawContent = "有"
		content.DrawString(drawContent, freetype.Pt(drawLeft, drawTop))

		drawLeft += (utf8.RuneCountInString(drawContent) * 18)
		content.SetSrc(image.NewUniform(color.RGBA{R: 237, G: 39, B: 90, A: 255})) // 设置字体颜色
		drawContent = fmt.Sprintf("%d", summary.MemberJoinCount)
		content.DrawString(drawContent, freetype.Pt(drawLeft, drawTop))

		drawLeft += (len(drawContent) * 10)
		content.SetSrc(image.Opaque)
		drawContent = "人加入，但也有"
		content.DrawString(drawContent, freetype.Pt(drawLeft, drawTop))

		drawLeft += (utf8.RuneCountInString(drawContent) * 18)
		content.SetSrc(image.NewUniform(color.RGBA{R: 237, G: 39, B: 90, A: 255})) // 设置字体颜色
		drawContent = fmt.Sprintf("%d", summary.MemberLeaveCount)
		content.DrawString(drawContent, freetype.Pt(drawLeft, drawTop))

		drawLeft += (len(drawContent) * 10)
		content.SetSrc(image.Opaque)
		drawContent = "人离开了我们，"
		content.DrawString(drawContent, freetype.Pt(drawLeft, drawTop))
	}

	drawLeft = 16
	drawTop += 30
	content.SetSrc(image.Opaque)
	drawContent = "共有"
	content.DrawString(drawContent, freetype.Pt(drawLeft, drawTop))

	drawLeft += (utf8.RuneCountInString(drawContent) * 18)
	content.SetSrc(image.NewUniform(color.RGBA{R: 237, G: 39, B: 90, A: 255})) // 设置字体颜色
	drawContent = fmt.Sprintf("%d", summary.MemberChatCount)
	content.DrawString(drawContent, freetype.Pt(drawLeft, drawTop))

	drawLeft += (len(drawContent) * 10)
	content.SetSrc(image.Opaque)
	drawContent = "人侃侃而谈"
	content.DrawString(drawContent, freetype.Pt(drawLeft, drawTop))

	drawLeft += (utf8.RuneCountInString(drawContent) * 18)
	content.SetSrc(image.NewUniform(color.RGBA{R: 237, G: 39, B: 90, A: 255})) // 设置字体颜色
	drawContent = fmt.Sprintf("%d", summary.MessageCount)
	content.DrawString(drawContent, freetype.Pt(drawLeft, drawTop))

	drawLeft += (len(drawContent) * 10)
	content.SetSrc(image.Opaque)
	drawContent = "句。"
	content.DrawString(drawContent, freetype.Pt(drawLeft, drawTop))

	drawLeft = 16
	drawTop = 288
	content.SetFontSize(14)
	content.SetSrc(image.Opaque)
	content.SetSrc(image.NewUniform(color.RGBA{R: 13, G: 47, B: 235, A: 255})) // 设置字体颜色
	drawContent = dailyWords
	content.DrawString(drawContent, freetype.Pt(drawLeft, drawTop))

	// 返回图片
	var buf bytes.Buffer
	if err := png.Encode(&buf, newBgImage); err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}
	return &buf, nil
}

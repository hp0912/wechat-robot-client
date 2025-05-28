package good_morning

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
	"wechat-robot-client/dto"
)

func TestDraw(t *testing.T) {
	// 获取当前时间
	now := time.Now()
	// 获取年、月、日
	year := now.Year()
	month := now.Month()
	day := now.Day()
	// 获取星期
	weekday := now.Weekday()
	// 定义中文星期数组
	weekdays := [...]string{"星期日", "星期一", "星期二", "星期三", "星期四", "星期五", "星期六"}

	summary := dto.ChatRoomSummary{}
	summary.ChatRoomID = "xxxxx"
	summary.Year = year
	summary.Month = int(month)
	summary.Date = day
	summary.Week = weekdays[weekday]
	summary.UserTotalCount = 490
	summary.UserJoinCount = 0
	summary.UserLeaveCount = 1
	summary.UserChatCount = 60
	summary.MessageCount = 1490

	image, err := Draw("知识是很美的，它们可以让你不出家门就了解这世上的许多事。", summary)
	if err != nil {
		t.Errorf("绘图失败: %v", err)
		return
	}
	// 将生成的图片保存到文件
	outputDir := "test_output"
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		t.Errorf("创建输出目录失败: %v", err)
		return
	}

	// 生成文件名（包含时间戳）
	timestamp := now.Format("20060102_150405")
	filename := filepath.Join(outputDir, "good_morning_"+timestamp+".png")

	// 创建文件
	file, err := os.Create(filename)
	if err != nil {
		t.Errorf("创建文件失败: %v", err)
		return
	}
	defer file.Close()

	// 将 io.Reader 的内容复制到文件
	_, err = io.Copy(file, image)
	if err != nil {
		t.Errorf("保存图片失败: %v", err)
		return
	}

	t.Logf("图片已成功保存到: %s", filename)
}

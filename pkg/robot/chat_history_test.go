package robot

import (
	"encoding/xml"
	"os"
	"strings"
	"testing"
)

func TestChatHistoryRecordItem_ParseRecordInfo(t *testing.T) {
	data, err := os.ReadFile("../../template/ch.xml")
	if err != nil {
		t.Fatalf("read ch.xml: %v", err)
	}

	var msg ChatHistoryMessage
	if err := xml.Unmarshal(data, &msg); err != nil {
		t.Fatalf("unmarshal outer msg: %v", err)
	}

	recordInfo, err := msg.AppMsg.RecordItem.ParseRecordInfo()
	if err != nil {
		t.Fatalf("parse recorditem: %v", err)
	}
	if recordInfo == nil {
		t.Fatalf("expected recordInfo, got nil")
	}
	if recordInfo.DataList.Count != 3 {
		t.Fatalf("expected datalist count=3, got %d", recordInfo.DataList.Count)
	}
	if got := len(recordInfo.DataList.Items); got != 3 {
		t.Fatalf("expected 3 data items, got %d", got)
	}

	// The third item in the sample has nested <recordxml><recordinfo>...</recordinfo></recordxml>.
	third := recordInfo.DataList.Items[2]
	if third.RecordXML == nil {
		t.Fatalf("expected third item to have recordxml")
	}
	if third.RecordXML.RecordInfo.DataList.Count != 4 {
		t.Fatalf("expected nested datalist count=4, got %d", third.RecordXML.RecordInfo.DataList.Count)
	}
	if got := len(third.RecordXML.RecordInfo.DataList.Items); got != 4 {
		t.Fatalf("expected 4 nested data items, got %d", got)
	}

	records := ExtractChatHistoryMessageRecords(recordInfo)
	if got := len(records); got != 4 {
		t.Fatalf("expected 4 extracted records (datatype 1 and 17), got %d", got)
	}
	if records[0].Nickname != "***" {
		t.Fatalf("expected first record nickname=***, got %q", records[0].Nickname)
	}
	if records[0].Content != "*** 说: 聊天不够活跃啊~~~" {
		t.Fatalf("expected first record content, got %q", records[0].Content)
	}

	// Ensure nested text records are included.
	var hasDuiDuiDui bool
	var duiDuiDuiContent string
	for _, r := range records {
		if r.Nickname == "对对对" && r.Content != "" {
			hasDuiDuiDui = true
			duiDuiDuiContent = r.Content
			break
		}
	}
	if !hasDuiDuiDui {
		t.Fatalf("expected to include nested text record from 对对对")
	}
	if !strings.HasPrefix(duiDuiDuiContent, "对对对 对 🔥阿布💢 说: ") {
		t.Fatalf("expected 对对对 content to be rewritten with mention prefix, got %q", duiDuiDuiContent)
	}
	if strings.Contains(duiDuiDuiContent, "@") {
		t.Fatalf("expected rewritten content not to contain '@', got %q", duiDuiDuiContent)
	}
}

func TestExtractChatHistoryMessageRecords_Mentions(t *testing.T) {
	recordInfo := &RecordInfo{
		DataList: DataList{
			Count: 1,
			Items: []DataItem{
				{
					DataType:      1,
					SourceName:    "nick",
					DataDesc:      "@张三\u2005@李四\u2005你好",
					SrcMsgLocalID: 1,
				},
			},
		},
	}

	records := ExtractChatHistoryMessageRecords(recordInfo)
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if got := records[0].Content; got != "nick 对 张三、李四 说: 你好" {
		t.Fatalf("unexpected rewritten content: %q", got)
	}

	recordInfo.DataList.Items[0].DataDesc = "@张三 @李四 你好"
	records = ExtractChatHistoryMessageRecords(recordInfo)
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if got := records[0].Content; got != "nick 对 张三、李四 说: 你好" {
		t.Fatalf("unexpected rewritten content (ascii spaces): %q", got)
	}

	// Last mention at end-of-string (no trailing separator).
	recordInfo.DataList.Items[0].DataDesc = "@张三\u2005你好@李四"
	records = ExtractChatHistoryMessageRecords(recordInfo)
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if got := records[0].Content; got != "nick 对 张三、李四 说: 你好" {
		t.Fatalf("unexpected rewritten content (mention at end): %q", got)
	}
}

func TestExtractChatHistoryMessageRecords_NoMentions(t *testing.T) {
	recordInfo := &RecordInfo{
		DataList: DataList{
			Count: 1,
			Items: []DataItem{
				{
					DataType:      1,
					SourceName:    "nick",
					DataDesc:      "你好",
					SrcMsgLocalID: 1,
				},
			},
		},
	}

	records := ExtractChatHistoryMessageRecords(recordInfo)
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if got := records[0].Content; got != "nick 说: 你好" {
		t.Fatalf("unexpected rewritten content: %q", got)
	}
}

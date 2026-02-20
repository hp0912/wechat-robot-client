package robot

import (
	"encoding/xml"
	"os"
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
	if records[0].Content != "聊天不够活跃啊~~~" {
		t.Fatalf("expected first record content, got %q", records[0].Content)
	}

	// Ensure nested text records are included.
	var hasDuiDuiDui bool
	for _, r := range records {
		if r.Nickname == "对对对" && r.Content != "" {
			hasDuiDuiDui = true
			break
		}
	}
	if !hasDuiDuiDui {
		t.Fatalf("expected to include nested text record from 对对对")
	}
}

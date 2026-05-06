package relay

import (
	"context"
	"testing"
)

type capturingLogWriter struct {
	count  int
	record UsageRecord
}

func (w *capturingLogWriter) WriteLog(ctx context.Context, record *UsageRecord) error {
	w.count++
	w.record = *record
	return nil
}

func TestRecordUsageFallsBackToSyncWriteWhenChannelIsFull(t *testing.T) {
	writer := &capturingLogWriter{}
	engine := &RelayEngine{
		usageCh:   make(chan UsageRecord),
		logWriter: writer,
	}

	engine.RecordUsage(UsageRecord{UserID: 7, Model: "gpt-4o"})

	if writer.count != 1 {
		t.Fatalf("usage channel 满时应同步写入，避免账务 hold 卡死，实际写入次数=%d", writer.count)
	}
	if writer.record.UserID != 7 || writer.record.Model != "gpt-4o" {
		t.Fatalf("同步兜底写入的记录不正确: %+v", writer.record)
	}
}

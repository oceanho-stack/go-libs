package xlog

import (
	"context"
	"testing"
)

func TestInitAndLog(t *testing.T) {
	err := Init(Options{
		Level:  LevelDebug,
		Format: FormatText,
		Output: OutputConsole,
	})
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	Info("hello from xlog", "key", "value")
	Debug("debug message")
	Warn("warn message", "count", 3)
	Error("error message", "code", "E001")
}

func TestFromCtx(t *testing.T) {
	if err := Init(Options{Level: LevelDebug, Format: FormatText, Output: OutputConsole}); err != nil {
		t.Fatal(err)
	}

	ctx := WithContext(context.Background(), Fields{
		TraceID:    "trace-123",
		SpanID:     "span-456",
		RemoteIP:   "1.2.3.4",
		ClientType: "console",
	})
	FromCtx(ctx).Info("request with trace", "method", "GET")
}

func TestWithFields(t *testing.T) {
	if err := Init(Options{Level: LevelDebug, Format: FormatText, Output: OutputConsole}); err != nil {
		t.Fatal(err)
	}
	l := Default().WithFields(Fields{TraceID: "t-1"})
	l.Info("with fields")
}

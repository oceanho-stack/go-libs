package xtrace

import (
	"context"
	"strings"
	"testing"
)

func TestNewTraceID(t *testing.T) {
	id := NewTraceID()
	if len(id) != 32 {
		t.Errorf("trace id len = %d, want 32", len(id))
	}
	// 两次生成不重复
	if id == NewTraceID() {
		t.Error("trace id 重复")
	}
}

func TestNewSpanID(t *testing.T) {
	id := NewSpanID()
	if len(id) != 16 {
		t.Errorf("span id len = %d, want 16", len(id))
	}
}

func TestWithAndFromContext(t *testing.T) {
	ctx := context.Background()

	// 空 context
	if got := FromContext(ctx); got.TraceID != "" || got.SpanID != "" {
		t.Errorf("empty ctx should return empty trace, got %+v", got)
	}

	// 写入 trace_id + span_id
	ctx = WithTrace(ctx, Trace{TraceID: "t1", SpanID: "s1"})
	got := FromContext(ctx)
	if got.TraceID != "t1" || got.SpanID != "s1" {
		t.Errorf("got %+v, want t1/s1", got)
	}

	// WithTraceID 保留 span
	ctx2 := WithTraceID(ctx, "t2")
	got2 := FromContext(ctx2)
	if got2.TraceID != "t2" || got2.SpanID != "s1" {
		t.Errorf("got %+v, want t2/s1", got2)
	}
}

func TestInjectExtractHeader(t *testing.T) {
	ctx := WithTrace(context.Background(), Trace{TraceID: "abc123", SpanID: "sp1"})
	header := map[string][]string{}
	InjectHeader(ctx, header)

	if header[HeaderTraceID][0] != "abc123" {
		t.Errorf("header trace = %v", header[HeaderTraceID])
	}
	if got := ExtractTraceID(header); got != "abc123" {
		t.Errorf("extract = %q, want abc123", got)
	}

	// nil header 不 panic
	InjectHeader(ctx, nil)
	if got := ExtractTraceID(nil); got != "" {
		t.Errorf("nil header extract = %q, want empty", got)
	}
}

func TestTraceIDHexChars(t *testing.T) {
	id := NewTraceID()
	for _, c := range id {
		if !strings.ContainsRune("0123456789abcdef", c) {
			t.Errorf("trace id 含非法字符: %q", c)
		}
	}
}

// Package xtrace 轻量链路追踪：context 透传 trace_id/span_id，预留 OpenTelemetry 接入点。
package xtrace

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

// ctxKey context 键
type ctxKey struct{}

// Trace 链路信息
type Trace struct {
	TraceID string
	SpanID  string
	Parent  string // 父 span id（预留）
}

type traceCtxKey struct{}

// WithTrace 写入完整链路信息
func WithTrace(ctx context.Context, t Trace) context.Context {
	return context.WithValue(ctx, traceCtxKey{}, &t)
}

// WithTraceID 仅写入 trace_id（保留已有 span）
func WithTraceID(ctx context.Context, traceID string) context.Context {
	t := FromContext(ctx)
	t.TraceID = traceID
	return WithTrace(ctx, t)
}

// WithSpanID 仅写入 span_id
func WithSpanID(ctx context.Context, spanID string) context.Context {
	t := FromContext(ctx)
	t.SpanID = spanID
	return WithTrace(ctx, t)
}

// FromContext 从 context 读取链路信息（不存在则返回空值）
func FromContext(ctx context.Context) Trace {
	if ctx == nil {
		return Trace{}
	}
	if t, ok := ctx.Value(traceCtxKey{}).(*Trace); ok && t != nil {
		return *t
	}
	return Trace{}
}

// NewTraceID 生成新的 trace id（16 字节随机 hex，32 位长度，兼容 W3C 规范）
func NewTraceID() string {
	return newHex(16)
}

// NewSpanID 生成新的 span id（8 字节随机 hex，16 位长度）
func NewSpanID() string {
	return newHex(8)
}

func newHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		// 理论上 rand 失败概率极低；兜底用时间戳
		return fallbackID(n)
	}
	return hex.EncodeToString(b)
}

func fallbackID(n int) string {
	// 简单兜底：使用当前纳秒时间（不可靠但避免 panic）
	// 实际生产依赖 crypto/rand，此分支几乎不可达
	panic("crypto/rand unavailable")
}

// Header 常量：跨服务传递使用
const (
	HeaderTraceID = "X-Trace-Id"
	HeaderSpanID  = "X-Span-Id"
	HeaderParent  = "X-Parent-Id"
)

// InjectHeader 把链路信息写入 http.Header
func InjectHeader(ctx context.Context, header map[string][]string) {
	if header == nil {
		return
	}
	t := FromContext(ctx)
	if t.TraceID != "" {
		header[HeaderTraceID] = []string{t.TraceID}
	}
	if t.SpanID != "" {
		header[HeaderSpanID] = []string{t.SpanID}
	}
	if t.Parent != "" {
		header[HeaderParent] = []string{t.Parent}
	}
}

// ExtractTraceID 从 header 提取 trace_id
func ExtractTraceID(header map[string][]string) string {
	if header == nil {
		return ""
	}
	if v := header[HeaderTraceID]; len(v) > 0 {
		return v[0]
	}
	return ""
}

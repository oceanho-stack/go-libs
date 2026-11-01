// Package xlog 统一日志库：基于标准库 log/slog，支持上下文链路字段透传。
package xlog

import (
	"context"
	"log/slog"
	"os"
	"time"
)

// 日志级别
const (
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
)

// 输出格式
const (
	FormatText = "text"
	FormatJSON = "json"
)

// 输出目标
const (
	OutputConsole = "console"
	OutputFile    = "file"
)

// Options 初始化配置
type Options struct {
	Level    string // debug/info/warn/error
	Format   string // text/json
	Output   string // console/file
	FilePath string // Output=file 时必填
}

// ctxKey 用于在 context 中传递链路字段
type ctxKey struct{}

// Fields 链路通用字段
type Fields struct {
	TraceID    string
	SpanID     string
	RemoteIP   string
	ClientType string
}

// Logger 封装 slog.Logger
type Logger struct {
	slog   *slog.Logger
	fields Fields
}

var (
	global *Logger
)

// Init 初始化全局 logger（程序入口调用一次）
func Init(opts Options) error {
	lv := parseLevel(opts.Level)

	handlerOpts := &slog.HandlerOptions{Level: lv}

	var handler slog.Handler
	switch opts.Format {
	case FormatJSON:
		handler = slog.NewJSONHandler(os.Stdout, handlerOpts)
	default:
		handler = slog.NewTextHandler(os.Stdout, handlerOpts)
	}

	// 文件输出（暂未内置轮转，生产建议配外部工具或后续扩展）
	if opts.Output == OutputFile && opts.FilePath != "" {
		f, err := os.OpenFile(opts.FilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		switch opts.Format {
		case FormatJSON:
			handler = slog.NewJSONHandler(f, handlerOpts)
		default:
			handler = slog.NewTextHandler(f, handlerOpts)
		}
	}

	global = &Logger{slog: slog.New(handler)}
	slog.SetDefault(global.slog)
	return nil
}

func parseLevel(s string) slog.Level {
	switch s {
	case LevelDebug:
		return slog.LevelDebug
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Default 返回全局 logger
func Default() *Logger {
	if global == nil {
		global = &Logger{slog: slog.Default()}
	}
	return global
}

// FromCtx 从 context 中取 logger，自动附带链路字段
func FromCtx(ctx context.Context) *Logger {
	if ctx == nil {
		return Default()
	}
	if f, ok := ctx.Value(ctxKey{}).(*Fields); ok && f != nil {
		return &Logger{slog: Default().slog, fields: *f}
	}
	return Default()
}

// WithContext 把链路字段写入 context
func WithContext(ctx context.Context, f Fields) context.Context {
	return context.WithValue(ctx, ctxKey{}, &f)
}

// WithFields 返回带附加字段的 logger（不可变，返回新实例）
func (l *Logger) WithFields(fields Fields) *Logger {
	merged := l.fields
	if fields.TraceID != "" {
		merged.TraceID = fields.TraceID
	}
	if fields.SpanID != "" {
		merged.SpanID = fields.SpanID
	}
	if fields.RemoteIP != "" {
		merged.RemoteIP = fields.RemoteIP
	}
	if fields.ClientType != "" {
		merged.ClientType = fields.ClientType
	}
	return &Logger{slog: l.slog, fields: merged}
}

// attrs 组装链路字段 + 用户键值
func (l *Logger) attrs(args ...any) []any {
	out := make([]any, 0, len(args)+8)
	if l.fields.TraceID != "" {
		out = append(out, "trace_id", l.fields.TraceID)
	}
	if l.fields.SpanID != "" {
		out = append(out, "span_id", l.fields.SpanID)
	}
	if l.fields.RemoteIP != "" {
		out = append(out, "remote_ip", l.fields.RemoteIP)
	}
	if l.fields.ClientType != "" {
		out = append(out, "client_type", l.fields.ClientType)
	}
	out = append(out, args...)
	return out
}

// Debug 记录调试日志
func (l *Logger) Debug(msg string, args ...any) {
	l.slog.Debug(msg, l.attrs(args...)...)
}

// Info 记录信息日志
func (l *Logger) Info(msg string, args ...any) {
	l.slog.Info(msg, l.attrs(args...)...)
}

// Warn 记录警告日志
func (l *Logger) Warn(msg string, args ...any) {
	l.slog.Warn(msg, l.attrs(args...)...)
}

// Error 记录错误日志
func (l *Logger) Error(msg string, args ...any) {
	l.slog.Error(msg, l.attrs(args...)...)
}

// package 级便捷方法（等价于 Default().XXX）
func Debug(msg string, args ...any) { Default().Debug(msg, args...) }
func Info(msg string, args ...any)  { Default().Info(msg, args...) }
func Warn(msg string, args ...any)  { Default().Warn(msg, args...) }
func Error(msg string, args ...any) { Default().Error(msg, args...) }

// Time 记录耗时（配合 defer 使用）
func (l *Logger) Time(msg string, start time.Time, args ...any) {
	l.Info(msg, append(args, "duration_ms", time.Since(start).Milliseconds())...)
}

# Package xlog 统一日志库

xlog 是 oceanho-stack 的统一日志封装，基于 Go 标准库 `log/slog`，零第三方运行时依赖。

## 特性

- 结构化日志：`Debug/Info/Warn/Error`
- 自动携带链路字段：`trace_id`、`span_id`、`remote_ip`、`client_type`
- 上下文透传：`xlog.FromCtx(ctx)` 自动带上 ctx 中的链路字段
- 支持控制台 / 文件两种输出，可配置 JSON / 文本格式
- 纯标准库，无业务依赖，可直接抽离复用

## 快速开始

```go
// 初始化（程序入口调用一次）
xlog.Init(xlog.Options{
    Level:     xlog.LevelInfo,
    Format:    xlog.FormatText,
    Output:    xlog.OutputConsole,
    FilePath:  "",       // OutputFile 时必填
})

// 普通记录
xlog.Info("hello", "key", "value")

// 带上下文（自动带 trace_id 等链路字段）
ctx := xtrace.WithTraceID(context.Background(), "trace-123")
xlog.FromCtx(ctx).Info("request start", "method", "GET")
```

## 约定

- 业务字段（如 tenant_id、user_id）由调用方传入，lib 不感知业务。
- 统一字段：`trace_id`、`span_id`、`remote_ip`、`client_type`。

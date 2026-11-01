# Package xtrace 链路追踪库

xtrace 是 oceanho-stack 的轻量链路追踪封装，基于 context 透传 `trace_id` / `span_id`，预留 OpenTelemetry 接入点。

## 特性

- 生成 / 提取 / 透传 `trace_id`、`span_id`
- `WithTraceID(ctx, id)`、`WithSpanID(ctx, id)` 写入 context
- `FromContext(ctx)` 读取链路字段
- HTTP Header 注入 / 提取（跨服务传递）
- 无业务依赖；预留 otel 钩子（当前不强制依赖 otel 库）

## 快速开始

```go
// 服务入口：有上游 header 则提取，否则新建
traceID := xtrace.ExtractTraceID(header)
if traceID == "" {
    traceID = xtrace.NewTraceID()
}
ctx := xtrace.WithTraceID(context.Background(), traceID)

// 内部传递（日志等）
xlog.FromCtx(ctx).Info("...") // 自动带 trace_id

// 下游调用：注入 header
req.Header.Set("X-Trace-Id", xtrace.FromContext(ctx).TraceID)
```

## 配套

- `xlog.FromCtx` 自动读取本库写入的链路字段

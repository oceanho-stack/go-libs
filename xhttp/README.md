# Package xhttp 通用 HTTP 客户端

xhttp 是 oceanho-stack 的 HTTP 客户端封装，基于标准库 `net/http`，无第三方依赖、不绑定 web 框架。

## 特性

- `Client`：超时、基础 URL、默认 header、trace 注入
- `Get` / `PostJSON` / `Do` 便捷方法
- JSON 序列化 / 反序列化
- 可选简单重试（对可重试的临时错误）
- 不依赖 gin/echo，可作为独立客户端库复用

## 快速开始

```go
c := xhttp.NewClient(xhttp.Options{
    BaseURL:   "https://api.example.com",
    Timeout:   10 * time.Second,
    Headers:   map[string]string{"Authorization": "Bearer xxx"},
})

// 带 context 和 trace
ctx := xtrace.WithTraceID(context.Background(), "trace-1")
var result Resp
err := c.PostJSON(ctx, "/v1/tenants", req, &result)
```

## 设计边界

- 纯客户端能力，**不包含 gin 中间件**（recovery/CORS/trace 等 web 中间件按业务需要放各服务项目 pkg）
- 不定义业务响应结构（Resp 是产品协议，留业务项目）

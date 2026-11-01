# Package xvalidator 参数校验库

xvalidator 是 oceanho-stack 的统一参数校验封装，基于 `go-playground/validator/v10`。

## 特性

- 结构体 tag 校验（`validate:"required"`）
- 统一错误翻译为中文消息
- 内置自定义 tag：`id`（雪花ID/正整数）、`phone`（手机号，可选）
- 错误聚合返回

## 快速开始

```go
type Req struct {
    Name  string `json:"name" validate:"required,min=2,max=32"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"gte=0,lte=150"`
}

err := xvalidator.Validate(Req{Name: "a", Email: "bad", Age: 200})
// 返回聚合错误，可用 errors.As 取出 xvalidator.Errors
```

## 错误结构

- `xvalidator.Errors`：实现 error 接口，包含多个 `FieldError`
- 每项：字段名（json tag）、tag、消息

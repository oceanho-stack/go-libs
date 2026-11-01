# Package xtime 统一时间处理库

xtime 收拢业务侧反复出现的时间痛点：统一的序列化/解析格式、多布局容错解析、业务默认时区。**不重新发明 time 的功能**，只约定业务侧的使用方式。

## 特性

- **统一布局常量**：RFC3339 / 日期 / 时间 / 紧凑格式
- **容错解析**：`ParseFlexible` 依次尝试多个常见布局
- **业务默认时区**：默认 `Asia/Shanghai`，`SetBusinessTZ` 可覆盖
- **格式化助手**：日期/时间/完整时间/紧凑格式（自动按业务时区）
- **时间戳互转**：毫秒/微秒 ⇄ `time.Time`
- **日期运算**：`TruncateDay`、`DaysBetween`（自然日差）

## 快速开始

```go
// 业务默认时区（可选，默认 Asia/Shanghai）
if err := xtime.SetBusinessTZ("Asia/Shanghai"); err != nil { panic(err) }

// 容错解析（多种输入格式都接受）
t, err := xtime.ParseFlexible("2026-11-02 10:30:05")
t2, _ := xtime.Parse("2026-11-02T10:30:00Z") // 严格 RFC3339

// 格式化（自动转业务时区）
xtime.FormatDateTimeIn(t) // 2026-11-02T10:30:05+08:00
xtime.FormatDateOnly(t)   // 2026-11-02
xtime.FormatCompact(t)    // 20261102T183005（文件名/日志安全）

// 时间戳
ms := xtime.UnixMilli(t)
back := xtime.FromUnixMilli(ms)

// 日期运算
days := xtime.DaysBetween(start, end) // 自然日差
today := xtime.TruncateDay(now)       // 当天零点
```

## 设计边界

- **不封装原生活动**：`Add`/`Sub`/`Duration` 直接用 `time`，不做一层薄壳
- **序列化约定**：对外 API（JSON 响应）建议用 RFC3339（`FormatDateTime`），数据库层交给 GORM
- **时区单一源**：所有 `Format*`/`TruncateDay` 统一走 `BusinessTZ()`，避免散落各处的手写 `+8h`

// Package xtime 统一时间处理库：业务侧序列化/解析/时区约定。
//
// 设计原则：不重新发明 time 的功能，只收拢业务侧反复出现的痛点——
// 统一的 JSON/字符串格式、多布局容错解析、业务默认时区。
package xtime

import (
	"errors"
	"fmt"
	"time"
)

// 业务常用时间布局（Go 参考时间 2006-01-02T15:04:05Z07:00）
const (
	// LayoutDateTime 完整日期时间，兼容 RFC3339（带时区偏移）
	LayoutDateTime = time.RFC3339
	// LayoutDateOnly 仅日期
	LayoutDateOnly = "2006-01-02"
	// LayoutTimeOnly 仅时间
	LayoutTimeOnly = "15:04:05"
	// LayoutDateTimeNano 纳秒精度日期时间
	LayoutDateTimeNano = time.RFC3339Nano
	// LayoutDateTimeCompact 紧凑无分隔符（文件名/日志安全）
	LayoutDateTimeCompact = "20060102T150405"
	// LayoutDateCompact 紧凑日期
	LayoutDateCompact = "20060102"
)

// 可选布局，ParseFlexible 会按序尝试
var flexibleLayouts = []string{
	time.RFC3339Nano,
	time.RFC3339,
	LayoutDateTime,
	LayoutDateTimeCompact,
	LayoutDateOnly,
	LayoutDateCompact,
	"2006-01-02 15:04:05",
	"2006-01-02 15:04",
	"2006/01/02 15:04:05",
	"2006/01/02 15:04",
	"2006/01/02",
	"01/02/2006",
}

var (
	// ErrParse 时间字符串解析失败（所有布局均不匹配）
	ErrParse = errors.New("xtime: 无法解析时间字符串")
	// ErrInvalidTZ 无效时区名
	ErrInvalidTZ = errors.New("xtime: 无效的时区名")
)

// businessTZ 业务默认时区，可用 SetBusinessTZ 覆盖（默认 Asia/Shanghai）。
var businessTZ = time.FixedZone("CST", 8*3600) // 简化默认：UTC+8，无 DST 波动

// SetBusinessTZ 设置业务默认时区（按 IANA 名，如 "Asia/Shanghai"）。
func SetBusinessTZ(name string) error {
	loc, err := time.LoadLocation(name)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidTZ, name)
	}
	businessTZ = loc
	return nil
}

// BusinessTZ 返回当前业务默认时区。
func BusinessTZ() *time.Location { return businessTZ }

// Now 返回当前时间（保持 UTC 值，仅标记业务时区位置）。
func Now() time.Time { return time.Now().In(businessTZ) }

// Parse 按业务默认时区解析，格式为 LayoutDateTime（RFC3339）。
// 若字符串无时区信息，则按业务时区解释。
func Parse(s string) (time.Time, error) { return ParseIn(s, businessTZ) }

// ParseIn 按指定时区解析 RFC3339 格式。
func ParseIn(s string, loc *time.Location) (time.Time, error) {
	t, err := time.ParseInLocation(time.RFC3339, s, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: %q: %v", ErrParse, s, err)
	}
	return t, nil
}

// ParseFlexible 依次尝试多个常见布局解析时间字符串，宽容处理不同输入格式。
// 返回的时间无时区信息时按业务默认时区解释。
func ParseFlexible(s string) (time.Time, error) { return ParseFlexibleIn(s, businessTZ) }

// ParseFlexibleIn 在指定时区下多布局解析。
func ParseFlexibleIn(s string, loc *time.Location) (time.Time, error) {
	for _, layout := range flexibleLayouts {
		if t, err := time.ParseInLocation(layout, s, loc); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("%w: %q", ErrParse, s)
}

// ParseDateOnly 仅解析日期部分（"2006-01-02"），时分秒归零。
func ParseDateOnly(s string) (time.Time, error) {
	t, err := time.ParseInLocation(LayoutDateOnly, s, businessTZ)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: %q: %v", ErrParse, s, err)
	}
	return t, nil
}

// FormatDateTime 格式化完整日期时间（RFC3339）。
func FormatDateTime(t time.Time) string { return t.UTC().Format(LayoutDateTime) }

// FormatDateTimeIn 按业务时区格式化完整日期时间。
func FormatDateTimeIn(t time.Time) string { return t.In(businessTZ).Format(LayoutDateTime) }

// FormatDateOnly 格式化仅日期（"2006-01-02"），按业务时区。
func FormatDateOnly(t time.Time) string { return t.In(businessTZ).Format(LayoutDateOnly) }

// FormatTimeOnly 格式化仅时间（"15:04:05"），按业务时区。
func FormatTimeOnly(t time.Time) string { return t.In(businessTZ).Format(LayoutTimeOnly) }

// FormatCompact 格式化紧凑日期时间（"20060102T150405"），适合文件名/日志。
func FormatCompact(t time.Time) string { return t.In(businessTZ).Format(LayoutDateTimeCompact) }

// UnixMilli 返回毫秒时间戳。
func UnixMilli(t time.Time) int64 { return t.UnixMilli() }

// UnixMicro 返回微秒时间戳。
func UnixMicro(t time.Time) int64 { return t.UnixMicro() }

// FromUnixMilli 毫秒时间戳 → time.Time（按业务时区）。
func FromUnixMilli(ms int64) time.Time {
	return time.UnixMilli(ms).In(businessTZ)
}

// FromUnixMicro 微秒时间戳 → time.Time（按业务时区）。
func FromUnixMicro(us int64) time.Time {
	return time.UnixMicro(us).In(businessTZ)
}

// TruncateDay 将时间截断到当天零点（按业务时区）。
func TruncateDay(t time.Time) time.Time {
	y, m, d := t.In(businessTZ).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, businessTZ)
}

// DaysBetween 返回 a 与 b 之间的自然日差（按业务时区，忽略时分秒）。
func DaysBetween(a, b time.Time) int {
	da := TruncateDay(a)
	db := TruncateDay(b)
	return int(db.Sub(da).Hours() / 24)
}

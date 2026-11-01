package xtime

import (
	"testing"
	"time"
)

func TestParseAndFormat(t *testing.T) {
	s := "2026-11-02T10:30:00Z"
	tm, err := Parse(s)
	if err != nil {
		t.Fatalf("Parse 失败: %v", err)
	}
	// RFC3339 往返
	back := FormatDateTime(tm)
	if back != "2026-11-02T10:30:00Z" {
		t.Errorf("FormatDateTime = %q, want RFC3339", back)
	}
}

func TestParseInBusinessTZ(t *testing.T) {
	// 无时区字符串按业务时区（UTC+8）解释
	tm, err := ParseFlexible("2026-11-02 10:30:00")
	if err != nil {
		t.Fatalf("ParseFlexible 失败: %v", err)
	}
	if tm.Location().String() != "CST" {
		t.Errorf("时区 = %v, want CST", tm.Location())
	}
	// 小时应为 10（未加时区偏移）
	if tm.Hour() != 10 {
		t.Errorf("Hour = %d, want 10", tm.Hour())
	}
}

func TestParseFlexible(t *testing.T) {
	layouts := []string{
		"2026-11-02T10:30:00Z",
		"2026-11-02 10:30:05",
		"2026/11/02 10:30",
		"2026/01/02",
		"11/02/2026",
	}
	for _, s := range layouts {
		if _, err := ParseFlexible(s); err != nil {
			t.Errorf("ParseFlexible(%q) 应成功, got %v", s, err)
		}
	}
}

func TestParseFlexibleInvalid(t *testing.T) {
	if _, err := ParseFlexible("not-a-date"); err == nil {
		t.Error("非法字符串应返回错误")
	}
}

func TestParseDateOnly(t *testing.T) {
	tm, err := ParseDateOnly("2026-11-02")
	if err != nil {
		t.Fatalf("ParseDateOnly 失败: %v", err)
	}
	if tm.Hour() != 0 || tm.Minute() != 0 || tm.Second() != 0 {
		t.Errorf("日期部分应归零, got %v", tm)
	}
}

func TestFormatHelpers(t *testing.T) {
	base := time.Date(2026, 11, 2, 10, 30, 5, 0, time.UTC)
	if got := FormatDateOnly(base); got != "2026-11-02" {
		t.Errorf("FormatDateOnly = %q", got)
	}
	// 业务时区 +8：10:30:05 UTC → 18:30:05 CST
	if got := FormatTimeOnly(base); got != "18:30:05" {
		t.Errorf("FormatTimeOnly = %q, want 18:30:05", got)
	}
	// 业务时区 +8
	if got := FormatCompact(base); got != "20261102T183005" {
		t.Errorf("FormatCompact = %q, want 20261102T183005", got)
	}
}

func TestUnixMilliRoundTrip(t *testing.T) {
	now := time.Now()
	ms := UnixMilli(now)
	back := FromUnixMilli(ms)
	if back.UnixMilli() != ms {
		t.Errorf("毫秒往返不一致: %d vs %d", back.UnixMilli(), ms)
	}
	us := UnixMicro(now)
	if FromUnixMicro(us).UnixMicro() != us {
		t.Errorf("微秒往返不一致")
	}
}

func TestTruncateDay(t *testing.T) {
	tm := time.Date(2026, 11, 2, 23, 59, 59, 999, time.UTC)
	day := TruncateDay(tm)
	if day.Hour() != 0 || day.Minute() != 0 || day.Second() != 0 {
		t.Errorf("TruncateDay 未归零: %v", day)
	}
}

func TestDaysBetween(t *testing.T) {
	loc := BusinessTZ()
	a := time.Date(2026, 11, 2, 1, 0, 0, 0, loc)
	b := time.Date(2026, 11, 5, 23, 0, 0, 0, loc)
	if d := DaysBetween(a, b); d != 3 {
		t.Errorf("DaysBetween = %d, want 3", d)
	}
	// 反向应为负数
	if d := DaysBetween(b, a); d != -3 {
		t.Errorf("DaysBetween 反向 = %d, want -3", d)
	}
}

func TestSetBusinessTZ(t *testing.T) {
	if err := SetBusinessTZ("Asia/Tokyo"); err != nil {
		t.Fatalf("SetBusinessTZ 失败: %v", err)
	}
	if BusinessTZ().String() != "Asia/Tokyo" {
		t.Errorf("BusinessTZ = %v", BusinessTZ())
	}
	// 恢复默认，避免影响其他测试
	if err := SetBusinessTZ("Asia/Shanghai"); err != nil {
		t.Fatalf("恢复失败: %v", err)
	}
}

func TestErrParseOnBadInput(t *testing.T) {
	if _, err := Parse("bad"); err == nil {
		t.Error("Parse(bad) 应返回错误")
	}
}

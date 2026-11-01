package xid

import (
	"testing"
)

func TestSnowflakeUnique(t *testing.T) {
	gen, err := NewSnowflake(1)
	if err != nil {
		t.Fatal(err)
	}

	seen := make(map[uint64]bool)
	for i := 0; i < 100000; i++ {
		id := gen.NextID()
		if seen[id] {
			t.Fatalf("重复 ID: %d", id)
		}
		seen[id] = true
	}
}

func TestSnowflakeOrdering(t *testing.T) {
	gen, _ := NewSnowflake(2)
	prev := gen.NextID()
	for i := 0; i < 10000; i++ {
		cur := gen.NextID()
		if cur <= prev {
			t.Fatalf("ID 应趋势递增: prev=%d cur=%d", prev, cur)
		}
		prev = cur
	}
}

func TestSnowflakeMachineIDBounds(t *testing.T) {
	if _, err := NewSnowflake(-1); err == nil {
		t.Error("machineID=-1 应报错")
	}
	if _, err := NewSnowflake(1024); err == nil {
		t.Error("machineID=1024 应报错")
	}
	if _, err := NewSnowflake(0); err != nil {
		t.Errorf("machineID=0 应合法: %v", err)
	}
	if _, err := NewSnowflake(1023); err != nil {
		t.Errorf("machineID=1023 应合法: %v", err)
	}
}

func TestSnowflakeDifferentMachines(t *testing.T) {
	g1, _ := NewSnowflake(1)
	g2, _ := NewSnowflake(2)

	seen := make(map[uint64]bool)
	for i := 0; i < 1000; i++ {
		if seen[g1.NextID()] || seen[g2.NextID()] {
			t.Fatal("不同机器ID产生重复ID")
		}
		seen[g1.NextID()] = true
		seen[g2.NextID()] = true
	}
}

func TestNextIDString(t *testing.T) {
	gen, _ := NewSnowflake(3)
	s := gen.NextIDString()
	if s == "" {
		t.Error("NextIDString 为空")
	}
	// 字符串应为纯数字
	for _, c := range s {
		if c < '0' || c > '9' {
			t.Errorf("NextIDString 含非数字: %q", s)
		}
	}
}

// 接口兼容性编译期断言
var _ Generator = (*Snowflake)(nil)

func TestInterface(t *testing.T) {
	var gen Generator
	gen, _ = NewSnowflake(5)
	if gen.NextID() == 0 {
		t.Error("接口调用失败")
	}
}

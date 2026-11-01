package xvalidator

import (
	"errors"
	"strings"
	"testing"
)

type userReq struct {
	Name  string `json:"name" validate:"required,min=2,max=32"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"gte=0,lte=150"`
	ID    uint64 `json:"id" validate:"id"`
}

func TestValid(t *testing.T) {
	req := userReq{Name: "ocean", Email: "a@b.com", Age: 30, ID: 123}
	if err := Validate(req); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

func TestMissingRequired(t *testing.T) {
	req := userReq{Age: 30, ID: 123}
	err := Validate(req)
	if err == nil {
		t.Fatal("expected error")
	}

	var es *Errors
	if !errors.As(err, &es) {
		t.Fatalf("expected *Errors, got %T", err)
	}
	if len(es.Errors) != 2 { // name + email 都缺失
		t.Errorf("errors len = %d, want 2: %v", len(es.Errors), es.Errors)
	}
}

func TestInvalidEmail(t *testing.T) {
	req := userReq{Name: "ocean", Email: "not-an-email", Age: 30, ID: 123}
	err := Validate(req)
	if err == nil {
		t.Fatal("expected error")
	}
	var es *Errors
	if !errors.As(err, &es) {
		t.Fatal("expected *Errors")
	}
	found := false
	for _, fe := range es.Errors {
		if fe.Field == "email" {
			found = true
			if !strings.Contains(fe.Message, "格式不正确") {
				t.Errorf("email msg = %q", fe.Message)
			}
		}
	}
	if !found {
		t.Error("未找到 email 字段错误")
	}
}

func TestCustomIDTag(t *testing.T) {
	// id = 0 不通过
	if err := Validate(userReq{Name: "ocean", Email: "a@b.com", Age: 30, ID: 0}); err == nil {
		t.Error("id=0 应校验失败")
	}
	// id = 正数通过
	if err := Validate(userReq{Name: "ocean", Email: "a@b.com", Age: 30, ID: 999}); err != nil {
		t.Errorf("id=999 应通过: %v", err)
	}
}

func TestRangeValidation(t *testing.T) {
	// age 超范围
	err := Validate(userReq{Name: "ocean", Email: "a@b.com", Age: 200, ID: 1})
	if err == nil {
		t.Fatal("age=200 应失败")
	}
	var es *Errors
	if !errors.As(err, &es) {
		t.Fatal("expected *Errors")
	}
	for _, fe := range es.Errors {
		if fe.Field == "age" && !strings.Contains(fe.Message, "不能大于") {
			t.Errorf("age msg = %q", fe.Message)
		}
	}
}

func TestErrorString(t *testing.T) {
	req := userReq{Name: "", Email: "", Age: 30, ID: 1}
	err := Validate(req)
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "必填") {
		t.Errorf("Error() = %q, 应含'必填'", msg)
	}
}

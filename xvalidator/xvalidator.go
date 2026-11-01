// Package xvalidator 统一参数校验：基于 go-playground/validator，带中文错误翻译。
package xvalidator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	// v 全局 validator 实例，注册自定义 tag
	v *validator.Validate
)

func init() {
	v = validator.New()

	// 使用 json tag 作为字段名（便于前端对齐字段）
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// 自定义校验：id（正整数 / 雪花ID）
	_ = v.RegisterValidation("id", func(fl validator.FieldLevel) bool {
		switch fl.Field().Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return fl.Field().Int() > 0
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return fl.Field().Uint() > 0
		case reflect.String:
			s := fl.Field().String()
			if s == "" {
				return false
			}
			for _, c := range s {
				if c < '0' || c > '9' {
					return false
				}
			}
			return true
		default:
			return false
		}
	})
}

// Validate 校验结构体，返回聚合错误；nil 表示通过。
func Validate(s any) error {
	if err := v.Struct(s); err != nil {
		return translate(err)
	}
	return nil
}

// FieldError 单个字段错误
type FieldError struct {
	Field   string `json:"field"`   // 字段名（json tag）
	Tag     string `json:"tag"`     // 校验规则
	Message string `json:"message"` // 中文消息
}

// Errors 聚合错误，实现 error 接口
type Errors struct {
	Errors []FieldError `json:"errors"`
}

func (e *Errors) Error() string {
	msgs := make([]string, 0, len(e.Errors))
	for _, fe := range e.Errors {
		msgs = append(msgs, fe.Message)
	}
	return strings.Join(msgs, "; ")
}

// Translate validator 错误为中文聚合错误
func translate(err error) error {
	verrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return &Errors{Errors: []FieldError{{
			Tag:     "unknown",
			Message: err.Error(),
		}}}
	}

	es := &Errors{Errors: make([]FieldError, 0, len(verrs))}
	for _, fe := range verrs {
		es.Errors = append(es.Errors, FieldError{
			Field:   fe.Field(),
			Tag:     fe.Tag(),
			Message: msg(fe.Field(), fe.Tag(), fe.Param()),
		})
	}
	return es
}

// msg 根据校验 tag 生成中文消息
func msg(field, tag, param string) string {
	switch tag {
	case "required":
		return fmt.Sprintf("%s 为必填项", field)
	case "email":
		return fmt.Sprintf("%s 格式不正确", field)
	case "min":
		return fmt.Sprintf("%s 长度/值不能小于 %s", field, param)
	case "max":
		return fmt.Sprintf("%s 长度/值不能大于 %s", field, param)
	case "gte":
		return fmt.Sprintf("%s 不能小于 %s", field, param)
	case "lte":
		return fmt.Sprintf("%s 不能大于 %s", field, param)
	case "id":
		return fmt.Sprintf("%s 必须为正整数", field)
	case "phone":
		return fmt.Sprintf("%s 手机号格式不正确", field)
	case "oneof":
		return fmt.Sprintf("%s 只能是 %s 之一", field, param)
	default:
		return fmt.Sprintf("%s 校验失败(%s)", field, tag)
	}
}

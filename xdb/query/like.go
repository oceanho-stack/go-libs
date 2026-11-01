package query

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// LikeMode like 匹配模式
type LikeMode string

const (
	LikeFull   LikeMode = "full"   // %val%
	LikePrefix LikeMode = "prefix" // val%
	LikeSuffix LikeMode = "suffix" // %val
)

// Like 模糊查询条件。
// Field 必须为数据库列名；实现 FieldRef，受 Spec 列白名单校验。
type Like struct {
	Field string
	Value string
	Mode  LikeMode
}

func NewLikeCond(field, value string, mode LikeMode) Like {
	return Like{Field: field, Value: value, Mode: mode}
}

// FieldName 返回引用的列名（供白名单校验）
func (l Like) FieldName() string { return l.Field }

func (l Like) Apply(db *gorm.DB) (*gorm.DB, error) {
	if l.Field == "" {
		return nil, fmt.Errorf("query: like field cannot be empty")
	}
	if l.Value == "" {
		return db, nil // 空值不加条件
	}
	pattern := "%" + escapeLike(l.Value) + "%"
	switch l.Mode {
	case LikePrefix:
		pattern = escapeLike(l.Value) + "%"
	case LikeSuffix:
		pattern = "%" + escapeLike(l.Value)
	}
	// ESCAPE '\' 让转义后的 %/_ 作为字面量（SQLite/MySQL/PostgreSQL 均支持）
	return db.Where(l.Field+" LIKE ? ESCAPE '\\'", pattern), nil
}

// escapeLike 转义 LIKE 通配符 % _
func escapeLike(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}

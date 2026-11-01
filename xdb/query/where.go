package query

import "gorm.io/gorm"

// Where 原生 where 条件（参数化，防 SQL 注入）。
// ⚠️ 逃生舱：Query 为裸 SQL 片段，不受 Spec 列白名单校验，调用方必须保证字段安全（禁止透传用户输入做列名）。
type Where struct {
	Query string
	Args  []any
}

func (w Where) Apply(db *gorm.DB) (*gorm.DB, error) {
	return db.Where(w.Query, w.Args...), nil
}

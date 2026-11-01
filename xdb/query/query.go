// Package query 通用查询条件构造器（xdb 子包）。
//
// 设计目标（相对传统 repo 查询封装的优化）：
//  1. Spec 不持有 db：构造时不绑定数据库连接，Build(db) 时才应用 → 可跨表、跨连接复用。
//  2. 列白名单强制：结构化条件（Like 等）声明引用的列，Spec 构建时校验，
//     不在白名单直接报错，把"手动自律"变成"机制强制"，杜绝字段注入。
//  3. Where 是逃生舱：支持任意参数化 SQL，但不受白名单约束（由调用方保证安全）。
package query

import "gorm.io/gorm"

// Condition 查询条件接口。
// Apply 将条件应用到 *gorm.DB 链式对象，返回处理后的 db 与错误。
type Condition interface {
	Apply(db *gorm.DB) (*gorm.DB, error)
}

// FieldRef 可选接口：声明条件引用的列名，供 Spec 做列白名单校验。
// 结构化条件（如 Like）实现它；原生 Where 不实现（逃生舱，不做白名单校验）。
type FieldRef interface {
	FieldName() string
}

// applyAll 逐个应用条件；若某个条件校验失败则整体失败。
func applyAll(db *gorm.DB, conds []Condition) (*gorm.DB, error) {
	chain := db
	var err error
	for _, c := range conds {
		chain, err = c.Apply(chain)
		if err != nil {
			return nil, err
		}
	}
	return chain, nil
}

package query

import (
	"fmt"

	"gorm.io/gorm"
)

// Spec 查询规约：收纳一组 Condition + 排序 + 预加载。
//
// 优化要点：
//  - 不持有 db：NewSpec() 无参，Build(db) 时才应用 → 同一个 Spec 可跨表、跨连接复用。
//  - 列白名单强制：WithAllowFields 设定允许引用的列，Build 时对实现 FieldRef 的条件校验，
//    越白名单直接报错（把"自律"变"机制"，防字段注入）。
//  - 分页不在 Spec 内（ListPage 入参给 page/size），Spec 只管条件 + 排序 + 预加载。
type Spec struct {
	conds   []Condition
	order   string
	preload []string
	allowed map[string]bool // 允许引用的列；nil 表示不校验
}

// NewSpec 创建查询规约（不绑定 db）
func NewSpec() *Spec {
	return &Spec{}
}

// AddCond 追加条件（Where/Like/And/Or/自定义 Condition）
func (s *Spec) AddCond(c Condition) *Spec {
	s.conds = append(s.conds, c)
	return s
}

// Order 设置排序表达式，如 "created_at DESC"
func (s *Spec) Order(expr string) *Spec {
	s.order = expr
	return s
}

// Preload 设置预加载关联（可多次）
func (s *Spec) Preload(fields ...string) *Spec {
	s.preload = append(s.preload, fields...)
	return s
}

// AllowFields 设置列白名单。传入的列名允许被结构化条件引用。
// 不调用本方法则不做白名单校验（但强烈建议业务层调用，防字段注入）。
func (s *Spec) AllowFields(fields ...string) *Spec {
	s.allowed = make(map[string]bool, len(fields))
	for _, f := range fields {
		s.allowed[f] = true
	}
	return s
}

// Build 把规约应用到 db（db 需已 Model 绑定对应表），返回组装完成的 *gorm.DB。
// 返回 error 表示条件校验失败（如越白名单），此时不会执行数据库 IO。
func (s *Spec) Build(db *gorm.DB) (*gorm.DB, error) {
	// 白名单校验
	if s.allowed != nil {
		if err := s.validate(db); err != nil {
			return nil, err
		}
	}

	chain, err := applyAll(db, s.conds)
	if err != nil {
		return nil, err
	}
	for _, p := range s.preload {
		chain = chain.Preload(p)
	}
	if s.order != "" {
		chain = chain.Order(s.order)
	}
	return chain, nil
}

// validate 递归校验所有条件引用的列是否在白名单内
func (s *Spec) validate(db *gorm.DB) error {
	for _, c := range s.conds {
		if err := s.checkCond(c); err != nil {
			return err
		}
	}
	return nil
}

func (s *Spec) checkCond(c Condition) error {
	switch v := c.(type) {
	case And:
		for _, sc := range v.conds {
			if err := s.checkCond(sc); err != nil {
				return err
			}
		}
	case Or:
		for _, sc := range v.conds {
			if err := s.checkCond(sc); err != nil {
				return err
			}
		}
	case FieldRef:
		if !s.allowed[v.FieldName()] {
			return fmt.Errorf("query: 列 %q 不在白名单内（注入防护）", v.FieldName())
		}
	}
	return nil
}

// Conds 返回当前条件数（便于测试/诊断）
func (s *Spec) Conds() []Condition { return s.conds }

package query

import "gorm.io/gorm"

// And 逻辑与：内部所有条件都必须满足。
// 递归应用子条件，且内部条件同样参与白名单校验（见 Spec.Build）。
type And struct {
	conds []Condition
}

func NewAnd(conds ...Condition) And {
	return And{conds: conds}
}

func (a And) Apply(db *gorm.DB) (*gorm.DB, error) {
	return applyAll(db, a.conds)
}

// Or 逻辑或：内部任一条件满足即可。
type Or struct {
	conds []Condition
}

func NewOr(conds ...Condition) Or {
	return Or{conds: conds}
}

func (o Or) Apply(db *gorm.DB) (*gorm.DB, error) {
	if len(o.conds) == 0 {
		return db, nil
	}
	// 每个子条件在独立 session 上构造成表达式（不执行），再在同一链上 OR 叠加：
	// db.Where(c1).Or(c2).Or(c3)...
	exprs := make([]*gorm.DB, 0, len(o.conds))
	for _, c := range o.conds {
		sub, err := c.Apply(db.Session(&gorm.Session{}))
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, sub)
	}

	chain := db.Where(exprs[0])
	for _, e := range exprs[1:] {
		chain = chain.Or(e)
	}
	return chain, nil
}

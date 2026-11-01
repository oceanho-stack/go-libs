package xdb

import "gorm.io/gorm"

// PageQuery 通用查询条件构造器（分页/排序/预加载）。
//
// 设计边界：本结构体**不含租户过滤等业务逻辑**。
// 业务项目如需租户隔离，应组合/内嵌本结构体扩展（如 TenantScopedQuery）。
type PageQuery struct {
	page     int
	pageSize int
	sort     string
	order    string
	preloads []string
}

// NewPageQuery 创建空查询条件
func NewPageQuery() *PageQuery {
	return &PageQuery{}
}

// WithPage 设置分页。page 从 1 开始；page/pageSize<=0 时不分页。
func (q *PageQuery) WithPage(page, pageSize int) *PageQuery {
	cp := *q
	cp.page = page
	cp.pageSize = pageSize
	return &cp
}

// WithSort 设置排序。sort 为字段名，order 为 asc/desc。
// 注意：调用方必须自行对 sort 做白名单校验，防止 SQL 注入。
func (q *PageQuery) WithSort(sort, order string) *PageQuery {
	cp := *q
	cp.sort = sort
	cp.order = order
	return &cp
}

// WithPreload 设置预加载关联。支持链式多次调用。
func (q *PageQuery) WithPreload(fields ...string) *PageQuery {
	cp := *q
	cp.preloads = append(cp.preloads, fields...)
	return &cp
}

// Apply 把条件应用到 *gorm.DB（不修改传入的 tx，返回新链）
func (q *PageQuery) Apply(tx *gorm.DB) *gorm.DB {
	for _, p := range q.preloads {
		tx = tx.Preload(p)
	}
	if q.sort != "" {
		tx = tx.Order(q.sort + " " + q.order)
	}
	if q.page > 0 && q.pageSize > 0 {
		tx = tx.Limit(q.pageSize).Offset((q.page - 1) * q.pageSize)
	}
	return tx
}

// Page 分页结果包装
type Page[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

// NewPage 构造分页结果
func NewPage[T any](list []T, total int64, page, pageSize int) Page[T] {
	return Page[T]{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
}

// Count 执行总数统计（不带分页条件）
func (q *PageQuery) Count(tx *gorm.DB) (int64, error) {
	var total int64
	err := tx.Count(&total).Error
	return total, err
}

// List 执行分页查询，返回列表（带分页条件）
func (q *PageQuery) List[T any](tx *gorm.DB) ([]T, error) {
	var list []T
	err := q.Apply(tx).Find(&list).Error
	return list, err
}

// ListAndCount 一次执行总数 + 分页查询
func (q *PageQuery) ListAndCount[T any](tx *gorm.DB) ([]T, int64, error) {
	var list []T
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Apply(tx).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

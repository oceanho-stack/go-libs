package xdb

import (
	"context"
	"fmt"
	"reflect"

	"github.com/oceanho-stack/go-libs/xdb/query"
	"gorm.io/gorm"
)

// IDType 支持的主键类型：uint64（BaseModel 默认）、int64/uint/string 等。
type IDType interface {
	int | int64 | uint | uint64 | string
}

// BaseRepo 通用泛型仓库基类。所有业务 Repo 通过嵌入复用，只写特有方法。
//
// 使用契约（比传统设计更收敛）：
//  1. 主键：默认约定主键列名为 id（对应 BaseModel.ID）。若表主键非 id，
//     请勿复用 GetByID/DeleteByID 等按主键方法，业务 Repo 自行手写。
//  2. 删除：Delete 是软删除（Model 需含 gorm.DeletedAt，BaseModel 已自带）；
//     物理删除必须显式调用 DeleteHard。杜绝默认物理删。
//  3. 更新：Update 是显式字段集（map/列白名单），不提供"更新所有非零字段"的隐式入口，
//     避免零值语义歧义与越权更新。
//  4. 事务：事务操作一律走 *Tx 方法 + 外部 xdb.Tx 编排；不要自行 Begin/Commit。
type BaseRepo[T any, ID IDType] struct {
	db *gorm.DB
}

// NewBaseRepo 创建基础仓库
func NewBaseRepo[T any, ID IDType](db *gorm.DB) *BaseRepo[T, ID] {
	return &BaseRepo[T, ID]{db: db}
}

// NewUint64BaseRepo 创建主键为 uint64 的仓库（BaseModel 默认）
func NewUint64BaseRepo[T any](db *gorm.DB) *BaseRepo[T, uint64] {
	return NewBaseRepo[T, uint64](db)
}

// NewInt64BaseRepo 创建主键为 int64 的仓库
func NewInt64BaseRepo[T any](db *gorm.DB) *BaseRepo[T, int64] {
	return NewBaseRepo[T, int64](db)
}

// NewStringBaseRepo 创建主键为 string 的仓库
func NewStringBaseRepo[T any](db *gorm.DB) *BaseRepo[T, string] {
	return NewBaseRepo[T, string](db)
}

// Model 返回绑定 T 类型的 *gorm.DB（供扩展查询）
func (r *BaseRepo[T, ID]) Model(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Model(new(T))
}

// GetByID 按主键查询单条；返回 gorm.ErrRecordNotFound 表示不存在
func (r *BaseRepo[T, ID]) GetByID(ctx context.Context, id ID) (*T, error) {
	var m T
	err := r.db.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// FindByIDs 按主键列表批量查询（IN）。空列表直接返回空切片不查库。
func (r *BaseRepo[T, ID]) FindByIDs(ctx context.Context, ids []ID) ([]T, error) {
	if len(ids) == 0 {
		return []T{}, nil
	}
	var list []T
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

// Create 新增单条
func (r *BaseRepo[T, ID]) Create(ctx context.Context, m *T) error {
	return r.db.WithContext(ctx).Create(m).Error
}

// CreateTx 事务内新增单条
func (r *BaseRepo[T, ID]) CreateTx(ctx context.Context, tx *gorm.DB, m *T) error {
	return tx.WithContext(ctx).Create(m).Error
}

// BatchCreateTx 事务内批量新增。注意：大切片请上层自行分片。
func (r *BaseRepo[T, ID]) BatchCreateTx(ctx context.Context, tx *gorm.DB, list []*T) error {
	return tx.WithContext(ctx).Create(list).Error
}

// Update 按主键更新显式字段集（map，key 为列名）。
// ⚠️ 只更新 map 中出现的列，禁止直接透传 HTTP 原始请求 map；
//    调用方必须手工构造允许更新的字段集合（防越权更新）。
func (r *BaseRepo[T, ID]) Update(ctx context.Context, id ID, data map[string]any) error {
	var m T
	return r.db.WithContext(ctx).Model(&m).Where("id = ?", id).Updates(data).Error
}

// UpdateTx 事务内按主键更新显式字段集
func (r *BaseRepo[T, ID]) UpdateTx(ctx context.Context, tx *gorm.DB, id ID, data map[string]any) error {
	var m T
	return tx.WithContext(ctx).Model(&m).Where("id = ?", id).Updates(data).Error
}

// Delete 按主键软删除（Model 需含 gorm.DeletedAt）。
// 若模型不含 DeletedAt 仍调用本方法，会触发物理删除 —— 因此这里显式拦截：
// 检测到无软删字段时直接报错，强制调用方改用 DeleteHard。
func (r *BaseRepo[T, ID]) Delete(ctx context.Context, id ID) error {
	if !hasDeletedAt[T]() {
		return fmt.Errorf("xdb: 模型 %T 无软删字段，Delete 已拒绝（防止意外物理删除），如需物理删除请用 DeleteHard", new(T))
	}
	var m T
	return r.db.WithContext(ctx).Delete(&m, id).Error
}

// DeleteTx 事务内软删除
func (r *BaseRepo[T, ID]) DeleteTx(ctx context.Context, tx *gorm.DB, id ID) error {
	if !hasDeletedAt[T]() {
		return fmt.Errorf("xdb: 模型 %T 无软删字段，DeleteTx 已拒绝，如需物理删除请用 DeleteHardTx", new(T))
	}
	var m T
	return tx.WithContext(ctx).Delete(&m, id).Error
}

// DeleteHard 显式物理删除（永久删除，不可恢复）。仅在业务明确需要时调用。
func (r *BaseRepo[T, ID]) DeleteHard(ctx context.Context, id ID) error {
	var m T
	return r.db.WithContext(ctx).Unscoped().Delete(&m, id).Error
}

// ListPage 分页查询（单表），返回 (list, total, err)。
// spec 需已 Model 绑定（通过 r.Model(ctx) 或 spec 内条件的 db 已绑定）。
// 注意：仅单表普通条件，不支持 JOIN / GROUP BY / DISTINCT。
func (r *BaseRepo[T, ID]) ListPage(ctx context.Context, spec *query.Spec, page, size int) ([]T, int64, error) {
	var list []T
	var total int64

	if size < 1 {
		size = 1
	}
	if page < 1 {
		page = 1
	}

	dbQuery, err := spec.Build(r.Model(ctx))
	if err != nil {
		return nil, 0, err
	}

	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err = dbQuery.Limit(size).Offset((page - 1) * size).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// Count 按规约统计总数
func (r *BaseRepo[T, ID]) Count(ctx context.Context, spec *query.Spec) (int64, error) {
	var total int64
	dbQuery, err := spec.Build(r.Model(ctx))
	if err != nil {
		return 0, err
	}
	err = dbQuery.Count(&total).Error
	return total, err
}

// hasDeletedAt 编译期检查类型 T 是否含 gorm.DeletedAt 字段
func hasDeletedAt[T any]() bool {
	var m T
	typ := reflect.TypeOf(m)
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return false
	}
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		// 直接字段或嵌套匿名结构体里的 DeletedAt
		if f.Type == reflect.TypeOf(gorm.DeletedAt{}) {
			return true
		}
		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			for j := 0; j < f.Type.NumField(); j++ {
				if f.Type.Field(j).Type == reflect.TypeOf(gorm.DeletedAt{}) {
					return true
				}
			}
		}
	}
	return false
}

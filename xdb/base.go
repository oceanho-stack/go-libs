package xdb

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 所有表的基础模型。
//
// 设计约定（优化 my-repo-syncer 的教训）：
//  - 主键固定 uint64（配合 xid 雪花 ID，避免 GORM 自增与雪花冲突时的手动主键）。
//  - 自带 CreatedAt/UpdatedAt（GORM 自动维护）。
//  - 自带 DeletedAt（软删字段）→ BaseRepo.Delete 天然是软删除；硬删需显式 DeleteHard。
//    这样把"是否软删"由 Model 结构决定，而非靠注释约定，杜绝意外物理删除。
type BaseModel struct {
	ID        uint64         `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

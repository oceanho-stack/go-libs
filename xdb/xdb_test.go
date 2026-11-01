package xdb

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// testModel 测试模型
type testModel struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `json:"name"`
	GroupID   uint           `json:"group_id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// setupDB 用纯 Go 内存 sqlite（无 cgo，无需外部数据库）
// 每个测试使用唯一内存库名，避免测试间数据累积
func setupDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s_%d?mode=memory&cache=shared", t.Name(), time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	if err := db.AutoMigrate(&testModel{}); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	// 造数据：3 组，每组 3 条
	for g := uint(1); g <= 3; g++ {
		for i := 0; i < 3; i++ {
			m := testModel{Name: "m", GroupID: g}
			if err := db.Create(&m).Error; err != nil {
				t.Fatal(err)
			}
		}
	}
	return db
}

func TestOpenAndPing(t *testing.T) {
	db := setupDB(t)
	if err := Ping(db); err != nil {
		t.Fatalf("Ping failed: %v", err)
	}
	sqlDB, err := RawDB(db)
	if err != nil || sqlDB == nil {
		t.Fatalf("RawDB failed: %v", err)
	}
}

func TestPageQueryListAndCount(t *testing.T) {
	db := setupDB(t)

	// 全量统计
	q := NewPageQuery()
	total, err := q.Count(db.Model(&testModel{}))
	if err != nil {
		t.Fatal(err)
	}
	if total != 9 {
		t.Errorf("total = %d, want 9", total)
	}

	// 分页：第 2 页，每页 4 条
	q2 := NewPageQuery().WithPage(2, 4).WithSort("id", "asc")
	list, total2, err := q2.ListAndCount[testModel](db.Model(&testModel{}))
	if err != nil {
		t.Fatal(err)
	}
	if total2 != 9 {
		t.Errorf("total2 = %d, want 9", total2)
	}
	if len(list) != 4 {
		t.Errorf("page2 list len = %d, want 4", len(list))
	}
	if list[0].ID != 5 {
		t.Errorf("page2 first id = %d, want 5", list[0].ID)
	}
}

func TestPageQueryWithFilter(t *testing.T) {
	db := setupDB(t)

	// 业务条件：group_id = 2，第 1 页 10 条
	tx := db.Model(&testModel{}).Where("group_id = ?", 2)
	q := NewPageQuery().WithPage(1, 10)
	list, total, err := q.ListAndCount[testModel](tx)
	if err != nil {
		t.Fatal(err)
	}
	if total != 3 {
		t.Errorf("filter total = %d, want 3", total)
	}
	if len(list) != 3 {
		t.Errorf("filter list len = %d, want 3", len(list))
	}
	for _, m := range list {
		if m.GroupID != 2 {
			t.Errorf("item group = %d, want 2", m.GroupID)
		}
	}
}

func TestQueryImmutable(t *testing.T) {
	db := setupDB(t)

	base := NewPageQuery()
	// 链式返回副本，base 不被修改
	_ = base.WithPage(1, 3).WithSort("id", "asc")
	_ = base.WithPage(2, 3)

	// base 应仍无分页条件：查询全量 9 条
	tx := db.Model(&testModel{})
	total, err := base.Count(tx)
	if err != nil {
		t.Fatal(err)
	}
	if total != 9 {
		t.Errorf("base count = %d, want 9（说明原对象被修改）", total)
	}
}

func TestTxRollback(t *testing.T) {
	db := setupDB(t)

	ctx := context.Background()
	// 事务内插入 1 条，然后返回错误 → 应回滚
	err := Tx(ctx, db, func(tx *gorm.DB) error {
		if err := tx.Create(&testModel{Name: "rollback", GroupID: 9}).Error; err != nil {
			return err
		}
		return gorm.ErrInvalidTransaction // 触发回滚
	})
	if err == nil {
		t.Fatal("expected tx error")
	}

	var count int64
	db.Model(&testModel{}).Where("name = ?", "rollback").Count(&count)
	if count != 0 {
		t.Errorf("rollback 后 count = %d, want 0", count)
	}
}

func TestTxCommit(t *testing.T) {
	db := setupDB(t)

	ctx := context.Background()
	err := Tx(ctx, db, func(tx *gorm.DB) error {
		return tx.Create(&testModel{Name: "commit", GroupID: 9}).Error
	})
	if err != nil {
		t.Fatalf("tx failed: %v", err)
	}

	var count int64
	db.Model(&testModel{}).Where("name = ?", "commit").Count(&count)
	if count != 1 {
		t.Errorf("commit 后 count = %d, want 1", count)
	}
}

package xdb

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/oceanho-stack/go-libs/xdb/query"
	"gorm.io/gorm"
)

// testUser 嵌入 BaseModel 的示例模型（主键 uint64，软删）
type testUser struct {
	BaseModel
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// testHard 无软删字段的模型（用于验证 Delete 保护）
type testHard struct {
	ID   uint64 `gorm:"primaryKey" json:"id"`
	Name string `json:"name"`
}

func setupUserDB(t *testing.T) (*gorm.DB, *BaseRepo[testUser, uint64]) {
	t.Helper()
	dsn := fmt.Sprintf("file:%s_%d?mode=memory&cache=shared", t.Name(), time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	if err := db.AutoMigrate(&testUser{}, &testHard{}); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	repo := NewUint64BaseRepo[testUser](db)
	// 造数据：3 条
	for i := 1; i <= 3; i++ {
		u := testUser{Name: fmt.Sprintf("user%d", i), Email: fmt.Sprintf("u%d@x.com", i), Age: 20 + i}
		if err := repo.Create(context.Background(), &u); err != nil {
			t.Fatal(err)
		}
	}
	return db, repo
}

func TestBaseRepoCRUD(t *testing.T) {
	db, repo := setupUserDB(t)
	ctx := context.Background()

	// GetByID
	u, err := repo.GetByID(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if u.Name != "user1" {
		t.Errorf("GetByID name = %q", u.Name)
	}

	// 不存在
	if _, err := repo.GetByID(ctx, 999); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("GetByID missing err = %v, want ErrRecordNotFound", err)
	}

	// Update（显式字段集）
	if err := repo.Update(ctx, 1, map[string]any{"name": "renamed"}); err != nil {
		t.Fatal(err)
	}
	u2, _ := repo.GetByID(ctx, 1)
	if u2.Name != "renamed" {
		t.Errorf("after update name = %q", u2.Name)
	}

	// FindByIDs
	list, err := repo.FindByIDs(ctx, []uint64{1, 3})
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Errorf("FindByIDs len = %d, want 2", len(list))
	}

	// 空列表不查库
	empty, _ := repo.FindByIDs(ctx, nil)
	if len(empty) != 0 {
		t.Errorf("FindByIDs(nil) len = %d, want 0", len(empty))
	}

	// 事务
	err = Tx(ctx, db, func(tx *gorm.DB) error {
		if err := repo.CreateTx(ctx, tx, &testUser{Name: "txuser", Age: 1}); err != nil {
			return err
		}
		return gorm.ErrInvalidTransaction
	})
	if err == nil {
		t.Fatal("expected tx err")
	}
	var cnt int64
	db.Model(&testUser{}).Where("name = ?", "txuser").Count(&cnt)
	if cnt != 0 {
		t.Errorf("rollback 后 txuser count = %d, want 0", cnt)
	}
}

func TestBaseRepoDeleteSoft(t *testing.T) {
	_, repo := setupUserDB(t)
	ctx := context.Background()

	// 软删除：Delete 后 GetByID 应找不到
	if err := repo.Delete(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetByID(ctx, 1); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("soft delete 后 GetByID err = %v, want ErrRecordNotFound", err)
	}

	// 但物理记录仍在（软删）
	var total int64
	repo.db.Unscoped().Model(&testUser{}).Count(&total)
	if total != 3 {
		t.Errorf("物理行数 = %d, want 3（应为软删）", total)
	}
}

func TestBaseRepoDeleteHard(t *testing.T) {
	_, repo := setupUserDB(t)
	ctx := context.Background()

	// 显式物理删除
	if err := repo.DeleteHard(ctx, 1); err != nil {
		t.Fatal(err)
	}
	var total int64
	repo.db.Unscoped().Model(&testUser{}).Count(&total)
	if total != 2 {
		t.Errorf("硬删后物理行数 = %d, want 2", total)
	}
}

func TestBaseRepoDeleteProtection(t *testing.T) {
	dsn := fmt.Sprintf("file:hard_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, _ := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	_ = db.AutoMigrate(&testHard{})

	repo := NewUint64BaseRepo[testHard](db)
	ctx := context.Background()

	// 无软删字段的模型调用 Delete 应被拒绝
	if err := repo.Delete(ctx, 1); err == nil {
		t.Error("无软删字段的 Delete 应被拒绝")
	}
	// 显式 DeleteHard 允许
	if err := repo.DeleteHard(ctx, 1); err != nil {
		t.Errorf("DeleteHard 应允许: %v", err)
	}
}

func TestBaseRepoListPage(t *testing.T) {
	_, repo := setupUserDB(t)
	ctx := context.Background()

	// 条件过滤 + 分页 + 排序
	spec := query.NewSpec().
		AddCond(query.NewLikeCond("name", "user", query.LikeFull)).
		AllowFields("name").
		Order("age desc")

	list, total, err := repo.ListPage(ctx, spec, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if len(list) != 2 {
		t.Errorf("list len = %d, want 2", len(list))
	}
	// 按 age desc：user3(23) 在最前
	if list[0].Name != "user3" {
		t.Errorf("first = %q, want user3", list[0].Name)
	}
}

func TestBaseRepoListPageWhitelist(t *testing.T) {
	_, repo := setupUserDB(t)
	ctx := context.Background()

	// 白名单只允许 name，查询 email 应报错（注入防护）
	spec := query.NewSpec().
		AddCond(query.NewLikeCond("email", "u1", query.LikeFull)).
		AllowFields("name")

	_, _, err := repo.ListPage(ctx, spec, 1, 10)
	if err == nil {
		t.Error("越白名单列应报错")
	}
}

func TestBaseRepoCount(t *testing.T) {
	_, repo := setupUserDB(t)
	ctx := context.Background()

	spec := query.NewSpec().AddCond(query.NewLikeCond("name", "user", query.LikePrefix))
	total, err := repo.Count(ctx, spec)
	if err != nil {
		t.Fatal(err)
	}
	if total != 3 {
		t.Errorf("count = %d, want 3", total)
	}
}

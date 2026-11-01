package query

import (
	"fmt"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type qModel struct {
	ID      uint64 `gorm:"primaryKey" json:"id"`
	Name    string `json:"name"`
	GroupID int    `json:"group_id"`
}

func setupQDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:q_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&qModel{}); err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 6; i++ {
		gid := i%2 + 1 // 1,2,1,2,1,2
		db.Create(&qModel{Name: fmt.Sprintf("n%d", i), GroupID: gid})
	}
	return db
}

func TestSpecBuildAndOrder(t *testing.T) {
	db := setupQDB(t)

	spec := NewSpec().AddCond(Where{Query: "group_id = ?", Args: []any{1}}).Order("id asc")
	q, err := spec.Build(db.Model(&qModel{}))
	if err != nil {
		t.Fatal(err)
	}
	var list []qModel
	if err := q.Find(&list).Error; err != nil {
		t.Fatal(err)
	}
	if len(list) != 3 {
		t.Errorf("len = %d, want 3", len(list))
	}
	for _, m := range list {
		if m.GroupID != 1 {
			t.Errorf("group = %d, want 1", m.GroupID)
		}
	}
}

func TestSpecWhitelist(t *testing.T) {
	db := setupQDB(t)

	// 白名单只允许 name，用 Like 引用 group_id 应被拦（结构化条件受白名单约束）
	spec := NewSpec().AddCond(NewLikeCond("group_id", "1", LikeFull)).AllowFields("name")
	if _, err := spec.Build(db.Model(&qModel{})); err == nil {
		t.Error("Like 引用未白名单列应报错")
	}

	// 白名单允许 group_id 则通过
	spec2 := NewSpec().AddCond(NewLikeCond("group_id", "1", LikeFull)).AllowFields("name", "group_id")
	if _, err := spec2.Build(db.Model(&qModel{})); err != nil {
		t.Errorf("应通过: %v", err)
	}

	// Where 是逃生舱：不受白名单约束（调用方自行保证安全）
	spec3 := NewSpec().AddCond(Where{Query: "group_id = ?", Args: []any{1}}).AllowFields("name")
	if _, err := spec3.Build(db.Model(&qModel{})); err != nil {
		t.Errorf("Where 逃生舱应放行: %v", err)
	}
}

func TestSpecWhitelistNested(t *testing.T) {
	db := setupQDB(t)

	// And 内嵌条件引用越界列，应被递归校验拦截
	spec := NewSpec().AddCond(NewAnd(NewLikeCond("name", "n", LikeFull))).
		AddCond(NewLikeCond("group_id", "1", LikeFull)).
		AllowFields("name")
	if _, err := spec.Build(db.Model(&qModel{})); err == nil {
		t.Error("And 内越白名单应报错")
	}
}

func TestLikeEscape(t *testing.T) {
	db := setupQDB(t)
	db.Create(&qModel{Name: "100%真", GroupID: 1})
	db.Create(&qModel{Name: "100假", GroupID: 1})

	spec := NewSpec().AddCond(NewLikeCond("name", "100%", LikeFull)).AllowFields("name")
	q, err := spec.Build(db.Model(&qModel{}))
	if err != nil {
		t.Fatal(err)
	}
	var list []qModel
	if err := q.Find(&list).Error; err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Errorf("转义后命中 %d 条, want 1（%% 不应当通配符）", len(list))
	}
	if list[0].Name != "100%真" {
		t.Errorf("命中 = %q, want 100%%真", list[0].Name)
	}
}

func TestOrCondition(t *testing.T) {
	db := setupQDB(t)

	spec := NewSpec().AddCond(NewOr(Where{Query: "name = ?", Args: []any{"n1"}}, Where{Query: "name = ?", Args: []any{"n2"}}))
	q, err := spec.Build(db.Model(&qModel{}))
	if err != nil {
		t.Fatal(err)
	}
	var list []qModel
	if err := q.Find(&list).Error; err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Errorf("Or 命中 %d 条, want 2", len(list))
	}
}

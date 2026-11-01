# Package xdb GORM 基础封装

xdb 是 oceanho-stack 的 GORM 统一底座封装，提供连接、连接池、慢日志、事务、**通用泛型仓库基类**、查询规约等能力。

## 特性

- 统一 `gorm.DB` 初始化：DSN、连接池（MaxOpen/MaxIdle/Lifetime）、慢 SQL 日志
- 日志对接 `xlog`（慢查询自动记录 trace_id）
- 事务抽象：`xdb.Tx(ctx, func(tx *gorm.DB) error)`
- **`BaseModel`**：基础模型（主键 `uint64` + 时间戳 + 软删字段），配合 xid 雪花 ID
- **`BaseRepo[T, ID]`**：泛型仓库基类（GetByID/FindByIDs/Create/Update/Delete/ListPage/Count），业务 Repo 嵌入复用
- **`query` 子包**：查询规约 `Spec` + 条件 `Where/Like/And/Or`，列白名单强制，跨表复用
- 分页助手：`Page[T]` 统一包装
- 软删除默认启用（gorm.DeletedAt），物理删除必须显式 `DeleteHard`

## 快速开始

```go
db, err := xdb.Open(xdb.Options{
    DSN:          "postgres://user:pass@localhost:5432/db",
    MaxOpenConns: 50,
    MaxIdleConns: 10,
    LogLevel:     "warn",
})
if err != nil { panic(err) }

// 定义模型（嵌入 BaseModel）
type User struct {
    xdb.BaseModel
    Name  string
    Email string
}

// 定义仓库（嵌入泛型基类，只需写特有方法）
type UserRepo struct {
    *xdb.NewUint64BaseRepo[User]
}
func (r *UserRepo) GetByName(ctx, name string) (*User, error) {
    var u User
    err := r.Model(ctx).Where("name = ?", name).First(&u).Error
    return &u, err
}

// 使用
repo := UserRepo{NewUint64BaseRepo[User](db)}
u, err := repo.GetByID(ctx, 123456)

// 分页查询（query 规约 + 列白名单）
spec := xdbquery.NewSpec().
    AddCond(xdbquery.NewLikeCond("name", "张", xdbquery.LikeFull)).
    AllowFields("name", "email").
    Order("created_at desc")
list, total, err := repo.ListPage(ctx, spec, 1, 20)
```

## 设计决策（相对传统 repo 封装的优化）

- **BaseRepo 泛型基类**：通用 CRUD 一次写好，每张表嵌入复用，只写特有方法（不是每张表一个完整 Repo）。
- **删除保护**：`Delete` 默认软删除（依赖 BaseModel 的 DeletedAt）；模型无软删字段时 `Delete` 会**被拒绝**，物理删除必须显式 `DeleteHard`，杜绝意外物理删除。
- **更新显式化**：`Update` 只更新传入的字段集（map），禁止透传 HTTP 原始请求，防越权更新；不提供"更新所有非零字段"的隐式入口，消灭零值语义歧义。
- **Spec 不持有 db**：`NewSpec()` 无参，`Build(db)` 时才应用 → 同一 Spec 可跨表、跨连接复用。
- **列白名单强制**：`AllowFields` 设定允许引用的列，结构化条件（Like 等）越白名单直接报错，把"手动自律"变成"机制强制"，防字段注入。`Where` 是逃生舱（支持任意参数化 SQL，不受白名单约束，由调用方保证安全）。

## 设计边界（重要）

- **不包含租户过滤、数据权限等业务规则** —— 这些在业务项目的 repository 层通过继承/组合扩展
- 不写数据库方言 SQL，全部走 GORM 链式 API
- 只依赖 `gorm.io/gorm` + 驱动 + `xlog`

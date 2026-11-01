# Package xdb GORM 基础封装

xdb 是 oceanho-stack 的 GORM 统一底座封装，提供连接、连接池、慢日志、事务、通用查询构造等能力。

## 特性

- 统一 `gorm.DB` 初始化：DSN、连接池（MaxOpen/MaxIdle/Lifetime）、慢 SQL 日志
- 日志对接 `xlog`（慢查询自动记录 trace_id）
- 事务抽象：`xdb.Tx(ctx, func(tx *gorm.DB) error)`
- 通用查询构造：`BasePageQuery`（分页/排序/预加载），**不含租户逻辑**（租户是业务，由业务项目扩展 `TenantScopedQuery`）
- 分页助手：`Page[T]` 统一包装
- 软删除默认启用（gorm.DeletedAt）

## 快速开始

```go
db, err := xdb.Open(xdb.Options{
    DSN:          "postgres://user:pass@localhost:5432/db",
    MaxOpenConns: 50,
    MaxIdleConns: 10,
    ConnMaxLifetime: time.Hour,
    SlowThreshold: 200 * time.Millisecond,
    LogLevel:     "warn",
})
if err != nil { panic(err) }

// 查询条件构造（业务 repo 使用）
q := xdb.NewPageQuery().
    WithPage(1, 20).
    WithSort("created_at", "desc").
    WithPreload("Members")
```

## 设计边界（重要）

- **不包含租户过滤、数据权限等业务规则** —— 这些在业务项目的 repository 层通过继承/组合扩展
- 不写数据库方言 SQL，全部走 GORM 链式 API
- 只依赖 `gorm.io/gorm` + 驱动 + `xlog`

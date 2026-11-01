# go-libs 公共 Go 工具库（monorepo）

oceanho-stack 的公共 Go 工具库集合，独立 git 仓库，无业务依赖，可跨项目复用。

## 仓库地址

- GitHub: https://github.com/oceanho-stack/go-libs
- Module: `github.com/oceanho-stack/go-libs`

## 包含的库

| 库 | 说明 | 状态 |
|---|---|---|
| xlog | 统一日志（slog 封装，链路字段透传） | ✅ 已实现 |
| xcfg | 配置加载（默认值→yaml→env→命令行） | 待实现 |
| xtrace | 链路追踪（context 透传 trace/span） | 待实现 |
| xdb | GORM 基础封装（连接池/慢日志/事务/BasePageQuery） | 待实现 |
| xvalidator | 参数校验包装 | 待实现 |
| xid | ID 生成器 | 待实现 |
| xcrypto | 加密工具 | 待实现 |
| xcache | 缓存抽象 | 待实现 |
| xhttp | HTTP 客户端与基础中间件 | 待实现 |

## 实现顺序（依赖方向）

```
xlog → xcfg → xtrace → xdb → xvalidator/xid/xcrypto → xcache → xhttp
```

## 依赖规则（硬性约束）

- **禁止反向依赖业务项目**（iam-server 等）
- 只依赖标准库 + 稳定第三方库（gorm、viper 等）
- 不定义业务规则（租户、权限、业务错误码等一律不进 lib）
- 每个库独立目录、独立 package，按需引入

## 使用方式（在业务项目中）

```go
import (
    "github.com/oceanho-stack/go-libs/xlog"
    "github.com/oceanho-stack/go-libs/xcfg"
)
```

# Package xcfg 统一配置加载库

xcfg 是 oceanho-stack 的统一配置加载封装，基于 `spf13/viper`。

## 特性

- 加载顺序：**默认值 → 配置文件(yaml) → 环境变量 → 显式覆盖**
- 结构体自动绑定（`mapstructure` tag）
- 配置校验（可选，使用 validator tag）
- 环境变量前缀自动映射
- 无业务依赖，可直接抽离复用

## 快速开始

```go
// 定义配置结构体（业务项目自定义）
type Config struct {
    App struct {
        Name string `mapstructure:"name" validate:"required"`
        Port int    `mapstructure:"port"`
    } `mapstructure:"app"`
    DB struct {
        DSN string `mapstructure:"dsn"`
    } `mapstructure:"db"`
}

// 加载
var cfg Config
err := xcfg.Load(xcfg.Options{
    File:     "config.yaml",
    EnvPrefix: "OCEAN",
    Defaults: map[string]any{
        "app.name": "iam-server",
        "app.port": 8080,
    },
    Target: &cfg,
})
```

## 环境变量映射规则

前缀 + 路径大写 + 下划线分隔：
- `OCEAN_APP_NAME` → `app.name`
- `OCEAN_DB_DSN` → `db.dsn`

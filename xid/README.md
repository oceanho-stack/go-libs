# Package xid ID 生成器

xid 是 oceanho-stack 的 ID 生成库，提供接口化、可切换实现的唯一 ID 生成能力。

## 特性

- `Generator` 接口：`NextID() uint64`、`NextIDStr() string`
- 雪花算法实现（`Snowflake`）：时间戳 + 机器ID + 序列号，趋势递增
- 机器ID自动从节点名/配置获取，避免多实例冲突
- 无第三方依赖，标准库实现

## 快速开始

```go
// 雪花算法（指定机器ID）
gen, err := xid.NewSnowflake(1)
if err != nil { panic(err) }

id := gen.NextID()      // uint64
idStr := gen.NextIDStr() // 十进制字符串
```

## 设计

- 接口化：后续可加 UUID、自增等实现，业务代码依赖接口不依赖具体实现
- 雪花：41位时间戳 + 10位机器ID + 12位序列号

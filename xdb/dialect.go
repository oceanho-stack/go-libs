package xdb

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// dialector 根据驱动名返回对应 gorm dialector
func dialector(driver, dsn string) gorm.Dialector {
	switch driver {
	case "mysql":
		return mysql.Open(dsn)
	case "sqlite":
		return sqlite.Open(dsn)
	case "postgres", "":
		return postgres.Open(dsn)
	default:
		// 未知驱动：panic 让调用方立刻发现配置错误
		panic(fmt.Sprintf("xdb: 未知数据库驱动 %q，支持 postgres/mysql/sqlite", driver))
	}
}

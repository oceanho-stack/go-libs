// Package xdb GORM 统一底座：连接/连接池/慢日志/事务/通用查询构造。
package xdb

import (
	"context"
	"database/sql"
	"time"

	"github.com/oceanho-stack/go-libs/xlog"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Options 数据库初始化配置
type Options struct {
	DSN             string        // 连接串
	Driver          string        // 驱动：postgres / mysql / sqlite（默认 postgres）
	MaxOpenConns    int           // 最大打开连接数
	MaxIdleConns    int           // 最大空闲连接数
	ConnMaxLifetime time.Duration // 连接最大存活时间
	ConnMaxIdleTime time.Duration // 连接最大空闲时间
	SlowThreshold   time.Duration // 慢 SQL 阈值（默认 200ms）
	LogLevel        string        // silent / error / warn / info（默认 warn）
}

// Open 初始化数据库连接
func Open(opts Options) (*gorm.DB, error) {
	if opts.Driver == "" {
		opts.Driver = "postgres"
	}
	if opts.SlowThreshold == 0 {
		opts.SlowThreshold = 200 * time.Millisecond
	}
	if opts.LogLevel == "" {
		opts.LogLevel = "warn"
	}

	gormLog := newGormLogger(parseLogLevel(opts.LogLevel), opts.SlowThreshold)

	db, err := gorm.Open(dialector(opts.Driver, opts.DSN), &gorm.Config{
		Logger: gormLog,
	})
	if err != nil {
		return nil, err
	}

	// 连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if opts.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(opts.MaxOpenConns)
	}
	if opts.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(opts.MaxIdleConns)
	}
	if opts.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(opts.ConnMaxLifetime)
	}
	if opts.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(opts.ConnMaxIdleTime)
	}

	return db, nil
}

// Tx 事务封装：fn 返回 error 则回滚，否则提交
func Tx(ctx context.Context, db *gorm.DB, fn func(tx *gorm.DB) error) error {
	return db.WithContext(ctx).Transaction(fn)
}

// Ping 健康检查
func Ping(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// RawDB 获取底层 *sql.DB（用于连接池状态、统计等，不用于业务查询）
func RawDB(db *gorm.DB) (*sql.DB, error) {
	return db.DB()
}

// ---------- 内部：gorm logger 对接 xlog ----------

func parseLogLevel(s string) logger.LogLevel {
	switch s {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "info":
		return logger.Info
	default:
		return logger.Warn
	}
}

type gormLogger struct {
	logger.Writer
	level logger.LogLevel
}

func newGormLogger(level logger.LogLevel, slow time.Duration) logger.Interface {
	return &gormLogger{level: level}
}

func (l *gormLogger) LogMode(level logger.LogLevel) logger.Interface {
	l.level = level
	return l
}

func (l *gormLogger) Info(ctx context.Context, msg string, args ...any) {
	if l.level >= logger.Info {
		xlog.FromCtx(ctx).Info(msg, "gorm", args)
	}
}

func (l *gormLogger) Warn(ctx context.Context, msg string, args ...any) {
	if l.level >= logger.Warn {
		xlog.FromCtx(ctx).Warn(msg, "gorm", args)
	}
}

func (l *gormLogger) Error(ctx context.Context, msg string, args ...any) {
	if l.level >= logger.Error {
		xlog.FromCtx(ctx).Error(msg, "gorm", args)
	}
}

func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sqlStr, rows := fc()

	fields := []any{
		"sql", sqlStr,
		"rows", rows,
		"duration_ms", elapsed.Milliseconds(),
	}

	switch {
	case err != nil && l.level >= logger.Error:
		xlog.FromCtx(ctx).Error("sql error", append(fields, "error", err)...)
	case l.level >= logger.Warn && err == gorm.ErrRecordNotFound:
		xlog.FromCtx(ctx).Warn("record not found", fields...)
	default:
		if l.level == logger.Info {
			xlog.FromCtx(ctx).Info("sql", fields...)
		}
	}
}

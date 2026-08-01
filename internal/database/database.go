package database

import (
	"fmt"
	"time"

	"forum/internal/config"
	"forum/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)
//连接i
func Open(cfg config.DatabaseConfig, debug bool) (*gorm.DB, error) {
	var dialector gorm.Dialector
	switch cfg.Driver {
	case "mysql":
		dialector = mysql.Open(cfg.DSN)
	default:
		return nil, fmt.Errorf("unsupported database driver %q", cfg.Driver)
	}

	logLevel := logger.Warn
	if debug {
		logLevel = logger.Info
	}
	db, err := gorm.Open(dialector, &gorm.Config{Logger: logger.Default.LogMode(logLevel)})
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get database pool: %w", err)
	}
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConnections)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConnections)
	lifetime := cfg.ConnectionMaxLifetime
	if lifetime <= 0 {
		lifetime = time.Hour
	}
	sqlDB.SetConnMaxLifetime(lifetime)
	return db, nil
}

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.User{},
		&model.Post{},
		&model.Comment{},
		&model.PostLike{},
	); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}
	return nil
}

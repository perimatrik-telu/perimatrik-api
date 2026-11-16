package database

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"perimatrik-api-go/internal/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(cfg *config.Config) (*gorm.DB, error) {
	dsn, err := connectionDSN(cfg)
	if err != nil {
		return nil, err
	}

	return connect(cfg, dsn)
}

func connectionDSN(cfg *config.Config) (string, error) {
	tlsMode := strings.ToLower(strings.TrimSpace(cfg.DB.TLS))
	switch tlsMode {
	case "", "auto":
		host := strings.TrimSuffix(strings.ToLower(cfg.DB.Host), ".")
		tlsMode = "false"
		if strings.HasSuffix(host, ".tidbcloud.com") {
			tlsMode = "true"
		}
	case "true", "false":
	default:
		return "", fmt.Errorf("invalid DB_TLS: expected auto, true, or false")
	}

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=true&loc=%s&tls=%s",
		cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Name,
		url.QueryEscape(cfg.Location.String()), tlsMode,
	)
	return dsn, nil
}

func connect(cfg *config.Config, dsn string) (*gorm.DB, error) {
	logLevel := logger.Warn
	if cfg.AppDebug {
		logLevel = logger.Info
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:  logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time { return time.Now().In(cfg.Location) },
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("database ping: %w", err)
	}
	return db, nil
}

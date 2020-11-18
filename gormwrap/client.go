package gormwrap

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/DataWorkbench/glog"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MySQLConfig struct {
	// Hosts sample "127.0.0.1:3306,127.0.0.1:3307,127.0.0.1:3308"
	Hosts       string `json:"hosts"         yaml:"hosts"         env:"HOSTS"                     validate:"required"`
	Users       string `json:"users"         yaml:"users"         env:"USERS"                     validate:"required"`
	Password    string `json:"password"      yaml:"password"      env:"PASSWORD"                  validate:"required"`
	Database    string `json:"database"      yaml:"database"      env:"DATABASE"                  validate:"required"`
	MaxIdleConn int    `json:"max_idle_conn" yaml:"max_idle_conn" env:"MAX_IDLE_CONN,default=16"  validate:"required"`
	MaxOpenConn int    `json:"max_open_conn" yaml:"max_open_conn" env:"MAX_OPEN_CONN,default=128" validate:"required"`
	// gorm log level: 1 => Silent, 2 => Error, 3 => Warn, 4 => Info
	LogLevel int `json:"log_level" yaml:"log_level" env:"LOG_LEVEL,default=3" validate:"gte=1,lte=4"`
	// ConnMaxLifetime unit seconds
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime" yaml:"conn_max_lifetime" env:"CONN_MAX_LIFETIME,default=10m" validate:"required"`
	// SlowThreshold time 0 indicates disabled
	SlowThreshold time.Duration `json:"slow_threshold" yaml:"slow_threshold" env:"SLOW_THRESHOLD,default=2s" validate:"gte=0"`
}

// NewMySQLConn return a grom.DB by mysql driver
// NOTICE: Must set glog.Logger into the ctx by glow.WithContext
func NewMySQLConn(ctx context.Context, cfg *MySQLConfig) (db *gorm.DB, err error) {
	lp := glog.FromContext(ctx)

	defer func() {
		if err != nil {
			lp.Error().Error("create mysql connection error", err).Fire()
		}
	}()

	lp.Info().Msg("connecting to mysql").String("hosts", cfg.Hosts).String("database", cfg.Database).Fire()

	hosts := strings.Split(strings.ReplaceAll(cfg.Hosts, " ", ""), ",")
	if len(hosts) == 0 {
		err = fmt.Errorf("invalid hosts %s", cfg.Hosts)
		return
	}

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Users, cfg.Password, hosts[0], cfg.Database,
	)

	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: &Logger{
			Level:         LogLevel(cfg.LogLevel),
			SlowThreshold: cfg.SlowThreshold,
			Output:        lp,
		},
	})
	if err != nil {
		return
	}

	// Set connection pool
	var sqlDB *sql.DB
	if sqlDB, err = db.DB(); err != nil {
		return
	}
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConn)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConn)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	//// TODO: Adds multiple databases if necessary
	//// import gorm.io/plugin/dbresolver
	return
}

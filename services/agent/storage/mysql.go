package storage

import (
	"context"
	"fmt"
	"log"

	"codev42/services/agent/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DBConnection : MySQL 연동 구조체 (GORM 이용)
type RDBConnection struct {
	DB *gorm.DB
}

func NewRDBConnection(dsn string) (*RDBConnection, error) {
	gormDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open gorm db: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB from gorm: %w", err)
	}
	if err = sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping mysql: %w", err)
	}

	return &RDBConnection{DB: gormDB}, nil
}

// Close : DB 종료
func (c *RDBConnection) Close() error {
	if c.DB != nil {
		if sqlDB, err := c.DB.DB(); err == nil {
			if errClose := sqlDB.Close(); errClose != nil {
				log.Printf("failed to close gorm db: %v", errClose)
			}
		}
	}
	return nil
}

// AutoMigrate : GORM의 AutoMigrate (필요에 따라 사용)
func (c *RDBConnection) AutoMigrate() error {
	return c.DB.AutoMigrate(&model.Project{}, &model.FileStruct{}, &model.Code{})
}

// Example: 트랜잭션 예시
// 필요 시 트랜잭션을 사용하는 헬퍼 메서드를 제공할 수도 있음
func (c *RDBConnection) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return c.DB.WithContext(ctx).Transaction(fn)
}

package model

import (
	"time"
)

// Project : projects 테이블에 대응
type Project struct {
	ID          int64  `gorm:"primaryKey"`
	Name        string `gorm:"size:100;not null"`
	Description string `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	FileStruct  []FileStruct `gorm:"foreignKey:ProjectID"`
}

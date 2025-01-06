package model

import "time"

// Function : functions 테이블에 대응
type Code struct {
	ID          int64  `gorm:"primaryKey"`
	FileID      int64  `gorm:"not null;index"`
	FuncName    string `gorm:"size:255;not null"`
	Description string `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

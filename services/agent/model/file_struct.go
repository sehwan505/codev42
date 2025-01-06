package model

import "time"

// File : files 테이블에 대응
type FileStruct struct {
	ID        int64  `gorm:"primaryKey"`
	ProjectID int64  `gorm:"not null;index"`
	FilePath  string `gorm:"size:255;not null"`
	Content   string `gorm:"type:text"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Functions []Code `gorm:"foreignKey:FileID"`
}

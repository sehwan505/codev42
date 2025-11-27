package model

import "time"

// Function : functions 테이블에 대응
type Code struct {
	ID              int64     `gorm:"primaryKey"`
	FileID          int64     `gorm:"not null;index"`
	FuncDeclaration string    `gorm:"size:255;not null"`
	Plan            string    `gorm:"type:text"`
	CodeChunk       string    `gorm:"type:text;not null"`
	ChunkHash       string    `gorm:"type:varchar(64);not null;unique"`
	CreatedAt       time.Time `gorm:"autoCreateTime"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"`
}

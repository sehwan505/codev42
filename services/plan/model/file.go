package model

import "time"

type File struct {
	ID            int64     `gorm:"primaryKey"`
	ProjectID     string    `gorm:"type:varchar(255)"`
	ProjectBranch string    `gorm:"type:varchar(100)"`
	FilePath      string    `gorm:"not null"`
	Directory     string    `gorm:"not null"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
	Codes         []Code    `gorm:"foreignKey:FileID"`
}

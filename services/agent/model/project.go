package model

import (
	"time"
)

type Project struct {
	ID          string    `gorm:"primaryKey;type:varchar(255)"`
	Branch      string    `gorm:"primaryKey;type:varchar(100)"`
	Name        string    `gorm:"type:varchar(255);not null"`
	Description string    `gorm:"type:text"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
	Files       []File    `gorm:"foreignKey:ProjectID,ProjectBranch;references:ID,Branch"`
}

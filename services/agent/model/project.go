package model

import (
	"time"
)

type Project struct {
	ID          int64  `gorm:"primaryKey"`
	Name        string `gorm:"size:100;not null"`
	Description string `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Files       []File `gorm:"foreignKey:ProjectID"`
}

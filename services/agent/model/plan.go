package model

import (
	"time"
)

type Annotation struct {
	ID          int64     `gorm:"primaryKey"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
	Name        string    `gorm:"type:varchar(255);not null"`
	Params      string    `gorm:"type:text"`
	Returns     string    `gorm:"type:text"`
	Description string    `gorm:"type:text"`
	PlanID      int64     `gorm:"not null"`
	Plan        Plan      `gorm:"foreignKey:PlanID;references:ID"`
}

type Plan struct {
	ID          int64        `gorm:"primaryKey"`
	DevPlanID   int64        `gorm:"not null"`
	DevPlan     DevPlan      `gorm:"foreignKey:DevPlanID"`
	CreatedAt   time.Time    `gorm:"autoCreateTime"`
	UpdatedAt   time.Time    `gorm:"autoUpdateTime"`
	ClassName   string       `gorm:"type:varchar(255);not null"`
	Annotations []Annotation `gorm:"foreignKey:PlanID;references:ID"`
}

type DevPlan struct {
	ID        int64     `gorm:"primaryKey"`
	ProjectID string    `gorm:"type:varchar(255);not null"`
	Branch    string    `gorm:"type:varchar(100);not null"`
	Project   Project   `gorm:"foreignKey:ProjectID,Branch;references:ID,Branch"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	Language  string    `gorm:"type:varchar(255);not null"`
	Plans     []Plan    `gorm:"foreignKey:DevPlanID"`
}

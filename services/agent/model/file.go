package model

type File struct {
	ID        int64  `gorm:"primaryKey"`
	ProjectID int64  `gorm:"not null;index"`
	FilePath  string `gorm:"not null"`
	Directory string `gorm:"not null"`
	CreatedAt uint64 `gorm:"autoCreateTime"`
	UpdatedAt uint64 `gorm:"autoUpdateTime"`
	Codes     []Code `gorm:"foreignKey:FileID"`
}

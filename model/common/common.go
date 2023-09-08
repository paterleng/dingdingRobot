package common

import (
	"gorm.io/gorm"
	"time"
)

type Base struct {
	CreatedAt time.Time      `gorm:"column:create_at"`
	UpdatedAt time.Time      `gorm:"column:update_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
}

package models

import (
	"gorm.io/gorm"
)

type Task struct {
	gorm.Model
	Synopsis    string `gorm:"uniqueindex"`
	Description string
}

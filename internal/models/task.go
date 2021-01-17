package models

import (
	"fmt"
	"gorm.io/gorm"
)

type Task struct {
	gorm.Model
	Synopsis    string `gorm:"uniqueindex"`
	Description string
}

func (t *Task) String() string {
	str := fmt.Sprintf("%d. %s", t.ID, t.Synopsis)
	if t.Description != "" {
		str = fmt.Sprintf("%s (%s)", str, t.Description)
	}
	str = fmt.Sprintf("%s [C:%s U:%s]", str, t.CreatedAt, t.UpdatedAt)
	if t.DeletedAt.Valid {
		str = fmt.Sprintf("%s {DELETED}", str)
	}
	return str
}

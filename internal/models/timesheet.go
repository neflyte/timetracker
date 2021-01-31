package models

import (
	"database/sql"
	"gorm.io/gorm"
	"time"
)

type Timesheet struct {
	gorm.Model
	Task      Task
	TaskID    uint
	StartTime time.Time `gorm:"not null"`
	StopTime  sql.NullTime
}

package models

import (
	"database/sql"
	"fmt"
	"gorm.io/gorm"
	"time"
)

type Timesheet struct {
	gorm.Model
	TaskID    uint
	StartTime time.Time `gorm:"not null"`
	StopTime  sql.NullTime
}

func (ts *Timesheet) String() string {
	str := fmt.Sprintf("%d. %s", ts.TaskID, ts.StartTime)
	if ts.StopTime.Valid {
		str = fmt.Sprintf("%s -> %s", str, ts.StopTime.Time)
	}
	return fmt.Sprintf("%s (%d)", str, ts.ID)
}

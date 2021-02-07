package models

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
	"testing"
)

const (
	TestDSN = "file:test.db?cache=shared&mode=memory"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func MustOpenTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(TestDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		t.Fatalf("error opening test db: %s", err)
	}
	err = db.AutoMigrate(new(TaskData), new(TimesheetData))
	if err != nil {
		t.Fatalf("error automigrating test db schema: %s", err)
	}
	return db
}

func CloseTestDB(t *testing.T, db *gorm.DB) {
	if db != nil {
		sqldb, err := db.DB()
		if err != nil {
			t.Logf("error getting sql.DB handle: %s\n", err)
		} else {
			err = sqldb.Close()
			if err != nil {
				t.Logf("error closing DB handle: %s\n", err)
			}
		}
	}
}

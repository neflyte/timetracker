package models

import (
	"errors"
	"fmt"
	"github.com/bluele/factory-go/factory"
	"github.com/neflyte/timetracker/internal/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"os"
	"testing"
)

const (
	TestDSN           = "file:test.db?cache=shared&mode=memory"
	numberOfTestTasks = 10
)

var (
	TaskFactory = factory.NewFactory(
		&TaskData{
			log:      logger.GetStructLogger("TaskData"),
			testMode: true,
		},
	).SeqInt("ID", func(n int) (interface{}, error) {
		return n, nil
	}).Attr("Synopsis", func(args factory.Args) (interface{}, error) {
		task, ok := args.Instance().(*TaskData)
		if !ok {
			return nil, errors.New("args.Instance() is not a *TaskData")
		}
		return fmt.Sprintf("Task-%d", task.ID), nil
	}).Attr("Description", func(args factory.Args) (interface{}, error) {
		task, ok := args.Instance().(*TaskData)
		if !ok {
			return nil, errors.New("args.Instance() is not a *TaskData")
		}
		return fmt.Sprintf("This is the description for task %s", task.Synopsis), nil
	}).OnCreate(func(args factory.Args) error {
		taskData, ok := args.Instance().(*TaskData)
		if !ok {
			return errors.New("args.Instance() is not a *TaskData")
		}
		task := Task(taskData)
		return task.Create()
	})
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func MustOpenTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(TestDSN), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Warn),
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

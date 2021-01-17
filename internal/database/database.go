package database

import (
	"database/sql"
	"fmt"
	"github.com/neflyte/timetracker/internal/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
)

var (
	DB         *gorm.DB
	gormConfig = &gorm.Config{
		Logger: gormlog.Default.LogMode(gormlog.Silent),
	}
)

func Open(fileName string) (*gorm.DB, error) {
	log := logger.GetLogger("Open")
	dsn := fmt.Sprintf("file:%s?_foreign_keys=1&_journal_mode=WAL&_mode=rwc", fileName)
	log.Printf("opening sqlite db at %s\n", dsn)
	return gorm.Open(sqlite.Open(dsn), gormConfig)
}

func Close(db *gorm.DB) {
	log := logger.GetLogger("Close")
	if db != nil {
		sqldb, err := db.DB()
		if err != nil {
			log.Printf("error getting sql.DB handle: %s\n", err)
			return
		}
		log.Printf("closing sqlite db")
		err = sqldb.Close()
		if err != nil {
			log.Printf("error closing DB handle: %s\n", err)
		}
	}
}

func CloseRows(rows *sql.Rows) {
	log := logger.GetLogger("CloseRows")
	if rows != nil {
		err := rows.Close()
		if err != nil {
			log.Printf("error closing rows: %s\n", err)
		}
	}
}

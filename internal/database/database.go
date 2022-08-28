package database

import (
	"database/sql"
	"fmt"

	"github.com/neflyte/timetracker/internal/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	// dbInstance is the singleton database handle
	dbInstance *gorm.DB

	gormConfig = &gorm.Config{
		Logger: newGormLogger(),
	}
	databaseLog = logger.GetPackageLogger("database")
)

// Open opens a new database connection to the specified SQLite database file
func Open(fileName string) (*gorm.DB, error) {
	log := logger.GetFuncLogger(databaseLog, "Open")
	dsn := fmt.Sprintf("file:%s?_foreign_keys=1&_journal_mode=WAL&_mode=rwc", fileName)
	log.Printf("opening sqlite db at %s\n", dsn)
	return gorm.Open(sqlite.Open(dsn), gormConfig)
}

// Close closes an open database connection
func Close(db *gorm.DB) {
	log := logger.GetFuncLogger(databaseLog, "Close")
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

// Get returns the singleton database connection
func Get() *gorm.DB {
	return dbInstance
}

// Set sets the singleton database connection
func Set(db *gorm.DB) {
	dbInstance = db
}

// CloseRows closes a sql.Rows object and logs any errors that occurred
func CloseRows(rows *sql.Rows) {
	log := logger.GetFuncLogger(databaseLog, "CloseRows")
	if rows != nil {
		err := rows.Close()
		if err != nil {
			log.Err(err).Msg("error closing sql rows")
		}
	}
}

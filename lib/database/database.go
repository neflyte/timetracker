package database

import (
	"database/sql"
	"fmt"

	"github.com/neflyte/timetracker/lib/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	// dbInstance is the singleton database handle
	dbInstance *gorm.DB
	// dbLogger is the database logger
	dbLogger = newGormLogger()
	// gormConfig is the GORM config struct
	gormConfig = &gorm.Config{
		Logger:                                   dbLogger,
		DisableForeignKeyConstraintWhenMigrating: true,
	}
	databaseLog = logger.GetPackageLogger("database")
)

// Open opens a new database connection to the specified SQLite database file
func Open(fileName string) (*gorm.DB, error) {
	log := logger.GetFuncLogger(databaseLog, "Open")
	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_mode=rwc", fileName)
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

// checkForeignKeys returns the enabled/disabled status of foreign key support
func checkForeignKeys(db *gorm.DB) (bool, error) {
	log := logger.GetFuncLogger(databaseLog, "CheckForeignKeys")
	checkResult := db.Raw("PRAGMA foreign_keys")
	if checkResult.Error != nil {
		log.Err(checkResult.Error).
			Msgf("error checking foreign keys status")
		return false, checkResult.Error
	}
	foreignKeysStatus := 0
	err := checkResult.Scan(&foreignKeysStatus).Error
	if err != nil {
		log.Err(err).
			Msg("error scanning foreign keys status")
		return false, err
	}
	if foreignKeysStatus == 1 {
		return true, nil
	}
	return false, nil
}

// EnableForeignKeys toggles foreign key support
func EnableForeignKeys(db *gorm.DB, enable bool) {
	log := logger.GetFuncLogger(databaseLog, "EnableForeignKeys")
	enableClause := "OFF"
	if enable {
		enableClause = "ON"
	}
	result := db.Exec(fmt.Sprintf("PRAGMA foreign_keys = %s", enableClause))
	if result.Error != nil {
		log.Err(result.Error).
			Msgf("error setting foreign keys to %s", enableClause)
		return
	}
	checkStatus, err := checkForeignKeys(db)
	if err != nil {
		log.Err(err).
			Msgf("error checking foreign keys status")
		return
	}
	log.Debug().
		Bool("checkStatus", checkStatus).
		Msg("status after update")
}

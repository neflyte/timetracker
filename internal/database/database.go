package database

import (
	"fmt"
	"github.com/neflyte/timetracker/internal/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	// dbInstance is the singleton database handle
	dbInstance *gorm.DB

	gormConfig = &gorm.Config{
		Logger: NewGormLogger(),
	}
	databaseLog = logger.GetPackageLogger("database")
)

func Open(fileName string) (*gorm.DB, error) {
	log := logger.GetFuncLogger(databaseLog, "Open")
	dsn := fmt.Sprintf("file:%s?_foreign_keys=1&_journal_mode=WAL&_mode=rwc", fileName)
	log.Printf("opening sqlite db at %s\n", dsn)
	return gorm.Open(sqlite.Open(dsn), gormConfig)
}

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

func Get() *gorm.DB {
	return dbInstance
}

func Set(db *gorm.DB) {
	dbInstance = db
}

// CloseRows isn't needed yet
/*func CloseRows(rows *sql.Rows) {
	log := databaseLog.With().Str("func", "CloseRows").Logger()
	if rows != nil {
		err := rows.Close()
		if err != nil {
			log.Printf("error closing rows: %s\n", err)
		}
	}
}*/

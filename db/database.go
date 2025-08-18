package db

import (
	"github.com/sa7mon/s3scanner/bucket"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

var db *gorm.DB

func Connect(dbConn string, migrate bool) error {
	// Connect to the database and run migrations if needed

	// We've already connected
	// TODO: Replace this with a sync.Once pattern
	if db != nil {
		return nil
	}

	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Hour,    // Slow SQL threshold
			LogLevel:                  logger.Error, // Log level
			IgnoreRecordNotFoundError: true,         // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,         // Enable color
		},
	)

	// https://github.com/go-gorm/postgres
	database, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dbConn,
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{
		Logger: gormLogger,
	})

	if err != nil {
		return err
	}

	if migrate {
		err = database.AutoMigrate(&bucket.Bucket{}, &bucket.Object{})
		if err != nil {
			return err
		}
	}

	db = database

	return nil
}
func StoreBucket(b *bucket.Bucket) error {
	if b.Exists == bucket.BucketNotExist {
		return nil
	}
	return db.Session(&gorm.Session{CreateBatchSize: 1000, FullSaveAssociations: true}).Create(&b).Error
}

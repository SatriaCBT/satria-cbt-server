package database

import (
	"fmt"
	"log"
	"os"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db   *gorm.DB
	once sync.Once
)

func Connect() *gorm.DB {
	once.Do(func() {
		dsn := fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_SSLMODE"),
			os.Getenv("DB_TIMEZONE"),
		)
		var err error
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		log.Println("Connected to database")
	})
	return db
}

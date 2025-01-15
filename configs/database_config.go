package configs

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)


func Database() *gorm.DB {
	if err := godotenv.Load(); err != nil {
        log.Fatalf("Error loading .env file")
    }
	dbHost := os.Getenv("DB_HOST")
    dbUser := os.Getenv("DB_USER")
    dbPassword := os.Getenv("DB_PASSWORD")
    dbName := os.Getenv("DB_NAME")
    dbPort := os.Getenv("DB_PORT")
    dbSslMode := os.Getenv("DB_SSLMODE")
    dbTimezone := os.Getenv("DB_TIMEZONE")

	fmt.Println(dbName)

    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
        dbHost, dbUser, dbPassword, dbName, dbPort, dbSslMode, dbTimezone)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}else {
		fmt.Println("Connected to database........")
	}

	return db
}
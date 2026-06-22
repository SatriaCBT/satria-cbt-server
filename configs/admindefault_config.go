package configs

import (
	"log"
	"math/rand"
	"github.com/Satria-CBT/satria-cbt-server/models"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func InitAdminDefault(db *gorm.DB) {
	var count int64
	db.Model(&models.Admins{}).Count(&count)
	if count == 0 {
		password := generateRandomPassword(8)

		admin := models.Admins{
			Name:     "Default Admin",
			Username: "admin",
			Email:    "admin@example.com",
			Password: hashPassword(password),
		}

		if err := db.Create(&admin).Error; err != nil {
			log.Fatalf("Failed to create default admin: %v", err)
		}

		log.Printf("Default admin created with username: %s and password: %s", admin.Username, password)
	}
}


func generateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	password := make([]byte, length)
	for i := range password {
		password[i] = charset[rand.Intn(len(charset))]
	}

	return string(password)
}

func hashPassword(password string) string {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	return string(hashed)
}
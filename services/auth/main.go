package main

import (
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/Satria-CBT/satria-cbt-server/pkg/auth"
	"github.com/Satria-CBT/satria-cbt-server/pkg/database"
	authmodels "github.com/Satria-CBT/satria-cbt-server/services/auth/models"

	"github.com/Satria-CBT/satria-cbt-server/services/auth/handlers"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func main() {
	godotenv.Load()

	db := database.Connect()
	db.AutoMigrate(&authmodels.Admin{}, &authmodels.Teacher{}, &authmodels.Student{})

	initDefaultAdmin(db)

	adminH := handlers.NewAdminHandler(db)
	teacherH := handlers.NewTeacherHandler(db)
	studentH := handlers.NewStudentHandler(db)

	app := fiber.New(fiber.Config{
		AppName:      "Satria Auth Service",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	api := app.Group("/api")

	admin := api.Group("/admin")
	admin.Post("/login", adminH.Login)
	admin.Post("/register", adminH.Register)
	admin.Get("/profile", auth.AuthenticateToken([]string{"admin"}), adminH.Profile)
	admin.Put("/:id", auth.AuthenticateToken([]string{"admin"}), adminH.Update)
	admin.Delete("/:id", auth.AuthenticateToken([]string{"admin"}), adminH.Delete)

	teacher := api.Group("/teacher")
	teacher.Post("/login", teacherH.Login)
	teacher.Post("/register", auth.AuthenticateToken([]string{"admin"}), teacherH.Register)
	teacher.Get("/profile", auth.AuthenticateToken([]string{"admin", "teacher"}), teacherH.Profile)
	teacher.Put("/:id", auth.AuthenticateToken([]string{"admin", "teacher"}), teacherH.Update)
	teacher.Delete("/:id", auth.AuthenticateToken([]string{"admin"}), teacherH.Delete)

	student := api.Group("/student")
	student.Post("/login", studentH.Login)
	student.Post("/register", auth.AuthenticateToken([]string{"admin", "teacher"}), studentH.Register)
	student.Get("/profile", auth.AuthenticateToken([]string{"admin", "teacher", "student"}), studentH.Profile)
	student.Put("/:id", auth.AuthenticateToken([]string{"admin", "teacher", "student"}), studentH.Update)
	student.Delete("/:id", auth.AuthenticateToken([]string{"admin", "teacher"}), studentH.Delete)

	port := os.Getenv("AUTH_SERVICE_PORT")
	if port == "" {
		port = "3001"
	}
	log.Printf("Auth service running on :%s", port)
	app.Listen(":" + port)
}

func initDefaultAdmin(db *gorm.DB) {
	var count int64
	db.Model(&authmodels.Admin{}).Count(&count)
	if count > 0 {
		return
	}

	password := generatePassword(8)
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	db.Create(&authmodels.Admin{
		Name:     "Default Admin",
		Username: "admin",
		Email:    "admin@example.com",
		Password: string(hashed),
	})

	log.Printf("Default admin created — username: admin, password: %s", password)
}

func generatePassword(length int) string {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	password := make([]byte, length)
	for i := range password {
		password[i] = charset[rand.Intn(len(charset))]
	}
	return string(password)
}

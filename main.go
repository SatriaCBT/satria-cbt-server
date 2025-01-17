package main

import (
	"fmt"
	"log"
	"satriacbtserver/configs"
	"satriacbtserver/models"
	"time"
	"satriacbtserver/controllers"
	"satriacbtserver/routers"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
        log.Fatalf("Error loading .env file")
    }
	app := fiber.New(
		fiber.Config{
			AppName: "Satria CBT Server",
			IdleTimeout:  time.Second * 10,
			ReadTimeout:  time.Second * 10,
			WriteTimeout: time.Second * 10,
			Prefork:      false,
			ServerHeader:  "Satria CBT Server",
		
		},
	)
	database := configs.Database()
	models.MigrationAdmin(database)
	models.MigrationStudents(database)
	models.MigrationTeachers(database)
	models.MigrationClass(database)
	configs.InitAdminDefault(database)

	adminController := controllers.NewAdminController()
	studentController := controllers.NewStudentController()
	teacherController := controllers.NewTeacherController()
	routers.NewRoutesAdmins(app, adminController)
	routers.NewRoutesStudents(app, studentController)
	routers.NewRoutesTeachers(app, teacherController)
	app.Listen(":3000")
	fmt.Println("Server running on port 3000")
}
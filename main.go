package main

import (
	"fmt"
	"log"
	"github.com/Satria-CBT/satria-cbt-server/configs"
	"github.com/Satria-CBT/satria-cbt-server/models"
	"time"
	"github.com/Satria-CBT/satria-cbt-server/controllers"
	"github.com/Satria-CBT/satria-cbt-server/routers"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	_ "github.com/Satria-CBT/satria-cbt-server/docs"
	"github.com/gofiber/swagger"
)

// @title Satria CBT Server API
// @version 1.0
// @description REST API server for Computer-Based Test (CBT) system
// @host localhost:3000
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @schemes http

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

	adminController := controllers.NewAdminController(database)
	studentController := controllers.NewStudentController(database)
	teacherController := controllers.NewTeacherController(database)
	classController := controllers.NewClassController(database, studentController, teacherController)

	routers.NewRoutesClass(app, classController)
	routers.NewRoutesAdmins(app, adminController)
	routers.NewRoutesStudents(app, studentController)
	routers.NewRoutesTeachers(app, teacherController)

	app.Get("/swagger/*", swagger.HandlerDefault)

	app.Listen(":3000")
	fmt.Println("Server running on port 3000")
}
package routers

import (
	"github.com/Satria-CBT/satria-cbt-server/controllers"
	"github.com/Satria-CBT/satria-cbt-server/middleware"

	"github.com/gofiber/fiber/v2"
)

func NewRoutesStudents(router fiber.Router, studentcontroller *controllers.StudentController,) {
	app := router.Group("/student")
	app.Post("/register", middleware.AuthenticateToken([]string{"admin", "teacher"}), studentcontroller.RegisterStudent)
	app.Post("/login", studentcontroller.LoginStudent)
	app.Get("/profile", middleware.AuthenticateToken([]string{"admin", "teacher", "student"}), studentcontroller.GetSessionProfileStudent)
	app.Put("/update/:id", middleware.AuthenticateToken([]string{"admin", "teacher", "student"}), studentcontroller.UpdateStudent)
	app.Delete("/delete/:id", middleware.AuthenticateToken([]string{"admin", "teacher"}), studentcontroller.DeleteStudent)
}
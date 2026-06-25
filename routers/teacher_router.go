package routers

import (
	"github.com/Satria-CBT/satria-cbt-server/controllers"
	"github.com/Satria-CBT/satria-cbt-server/middleware"

	"github.com/gofiber/fiber/v2"
)

func NewRoutesTeachers(router fiber.Router, teachercontroller *controllers.TeacherController,) {
	app := router.Group("/teacher")
	app.Post("/register", middleware.AuthenticateToken([]string{"admin"}), teachercontroller.RegisterTeacher)
	app.Post("/login", teachercontroller.LoginTeacher)
	app.Get("/profile", middleware.AuthenticateToken([]string{"admin", "teacher"}), teachercontroller.GetSessionProfileTeacher)
	app.Put("/update/:id", middleware.AuthenticateToken([]string{"admin", "teacher"}), teachercontroller.UpdateTeacher)
	app.Delete("/delete/:id", middleware.AuthenticateToken([]string{"admin"}), teachercontroller.DeleteTeacher)
}
package routers

import (
	"satriacbtserver/controllers"
	"github.com/gofiber/fiber/v2"
)

func NewRoutesTeachers(router fiber.Router, teachercontroller *controllers.TeacherController,) {
	app := router.Group("/student")
	app.Post("/register", teachercontroller.RegisterTeacher)
	app.Post("/login", teachercontroller.LoginTeacher)
	app.Get("/profile", teachercontroller.GetSessionProfileTeacher)
	app.Put("/update/:id", teachercontroller.UpdateTeacher)
	app.Delete("/delete/:id", teachercontroller.DeleteTeacher)
}
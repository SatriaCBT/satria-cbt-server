package routers

import (
	"satriacbtserver/controllers"
	"github.com/gofiber/fiber/v2"
)

func NewRoutesStudents(router fiber.Router, studentcontroller *controllers.StudentController,) {
	app := router.Group("/student")
	app.Post("/register", studentcontroller.RegisterStudent)
	app.Post("/login", studentcontroller.LoginStudent)
	app.Get("/profile", studentcontroller.GetSessionProfileStudent)
	app.Put("/update/:id", studentcontroller.UpdateStudent)
	app.Delete("/delete/:id", studentcontroller.DeleteStudent)
}
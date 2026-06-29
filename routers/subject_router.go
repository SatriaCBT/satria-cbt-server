package routers

import (
	"github.com/Satria-CBT/satria-cbt-server/controllers"
	"github.com/Satria-CBT/satria-cbt-server/middleware"

	"github.com/gofiber/fiber/v2"
)

func NewRoutesSubject(router fiber.Router, subjectController *controllers.SubjectController) {
	app := router.Group("/subjects")

	app.Post("/", middleware.AuthenticateToken([]string{"admin", "teacher"}), subjectController.Create)
	app.Get("/", middleware.AuthenticateToken([]string{"admin", "teacher", "student"}), subjectController.GetAll)
	app.Get("/:id", middleware.AuthenticateToken([]string{"admin", "teacher", "student"}), subjectController.GetByID)
	app.Put("/:id", middleware.AuthenticateToken([]string{"admin", "teacher"}), subjectController.Update)
	app.Delete("/:id", middleware.AuthenticateToken([]string{"admin"}), subjectController.Delete)
}

package routers

import (
	"github.com/Satria-CBT/satria-cbt-server/controllers"
	"github.com/Satria-CBT/satria-cbt-server/middleware"

	"github.com/gofiber/fiber/v2"
)

func NewRoutesQuestion(router fiber.Router, questionController *controllers.QuestionController) {
	app := router.Group("/questions")

	app.Post("/", middleware.AuthenticateToken([]string{"teacher"}), questionController.Create)
	app.Get("/", middleware.AuthenticateToken([]string{"admin", "teacher", "student"}), questionController.GetAll)
	app.Get("/:id", middleware.AuthenticateToken([]string{"admin", "teacher", "student"}), questionController.GetByID)
	app.Put("/:id", middleware.AuthenticateToken([]string{"teacher"}), questionController.Update)
	app.Delete("/:id", middleware.AuthenticateToken([]string{"teacher"}), questionController.Delete)
}

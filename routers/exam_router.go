package routers

import (
	"github.com/Satria-CBT/satria-cbt-server/controllers"
	"github.com/Satria-CBT/satria-cbt-server/middleware"

	"github.com/gofiber/fiber/v2"
)

func NewRoutesExam(router fiber.Router, examController *controllers.ExamController) {
	app := router.Group("/exams")

	app.Post("/", middleware.AuthenticateToken([]string{"teacher"}), examController.Create)
	app.Get("/", middleware.AuthenticateToken([]string{"admin", "teacher", "student"}), examController.GetAll)
	app.Get("/:id", middleware.AuthenticateToken([]string{"admin", "teacher", "student"}), examController.GetByID)
	app.Put("/:id", middleware.AuthenticateToken([]string{"teacher"}), examController.Update)
	app.Delete("/:id", middleware.AuthenticateToken([]string{"teacher"}), examController.Delete)
	app.Post("/:id/publish", middleware.AuthenticateToken([]string{"teacher"}), examController.Publish)
	app.Post("/:id/questions", middleware.AuthenticateToken([]string{"teacher"}), examController.AddQuestions)
	app.Delete("/:id/questions/:questionId", middleware.AuthenticateToken([]string{"teacher"}), examController.RemoveQuestion)
}

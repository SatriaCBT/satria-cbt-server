package routers

import (
	"github.com/Satria-CBT/satria-cbt-server/controllers"
	"github.com/Satria-CBT/satria-cbt-server/middleware"

	"github.com/gofiber/fiber/v2"
)

func NewRoutesExport(router fiber.Router, exportController *controllers.ExportController) {
	app := router.Group("/export")

	app.Get("/exams/:examId/results", middleware.AuthenticateToken([]string{"admin", "teacher"}), exportController.ExportExamResults)
	app.Get("/students/:studentId/results", middleware.AuthenticateToken([]string{"admin", "teacher"}), exportController.ExportStudentResults)
}

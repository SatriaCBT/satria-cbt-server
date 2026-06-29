package routers

import (
	"github.com/Satria-CBT/satria-cbt-server/controllers"
	"github.com/Satria-CBT/satria-cbt-server/middleware"

	"github.com/gofiber/fiber/v2"
)

func NewRoutesExamAttempt(router fiber.Router, attemptController *controllers.ExamAttemptController) {
	app := router.Group("/attempts")

	app.Post("/exams/:examId/start", middleware.AuthenticateToken([]string{"student"}), attemptController.Start)
	app.Post("/:attemptId/submit", middleware.AuthenticateToken([]string{"student"}), attemptController.Submit)
	app.Get("/:attemptId", middleware.AuthenticateToken([]string{"admin", "teacher", "student"}), attemptController.GetAttempt)
	app.Get("/exams/:examId", middleware.AuthenticateToken([]string{"admin", "teacher"}), attemptController.GetAttemptsByExam)
	app.Get("/my", middleware.AuthenticateToken([]string{"student"}), attemptController.GetMyAttempts)
	app.Put("/:attemptId/grade", middleware.AuthenticateToken([]string{"teacher"}), attemptController.GradeEssay)
}

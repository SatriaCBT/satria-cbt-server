package routers

import (
	"github.com/Satria-CBT/satria-cbt-server/controllers"
	"github.com/Satria-CBT/satria-cbt-server/middleware"

	"github.com/gofiber/fiber/v2"
)

func NewRoutesDashboard(router fiber.Router, dashboardController *controllers.DashboardController) {
	app := router.Group("/dashboard")

	app.Get("/stats", middleware.AuthenticateToken([]string{"admin", "teacher"}), dashboardController.Stats)
	app.Get("/exams/:examId/stats", middleware.AuthenticateToken([]string{"admin", "teacher"}), dashboardController.ExamStats)
	app.Get("/students/:studentId/performance", middleware.AuthenticateToken([]string{"admin", "teacher"}), dashboardController.StudentPerformance)
	app.Get("/classes/:classId/performance", middleware.AuthenticateToken([]string{"admin", "teacher"}), dashboardController.ClassPerformance)
	app.Get("/recent-activity", middleware.AuthenticateToken([]string{"admin", "teacher"}), dashboardController.RecentActivity)
}

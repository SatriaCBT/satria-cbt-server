package main

import (
	"log"
	"os"

	"github.com/Satria-CBT/satria-cbt-server/pkg/database"
	"github.com/Satria-CBT/satria-cbt-server/services/exam/handlers"
	"github.com/Satria-CBT/satria-cbt-server/services/exam/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	db := database.Connect()
	db.AutoMigrate(
		&models.Subject{},
		&models.Question{},
		&models.Exam{},
		&models.ExamQuestion{},
		&models.ExamAttempt{},
		&models.ExamAnswer{},
		&models.Class{},
	)

	h := handlers.NewExamHandler(db)

	app := fiber.New()
	app.Use(cors.New())

	subjects := app.Group("/api/subjects")
	subjects.Post("/", h.CreateSubject)
	subjects.Get("/", h.GetSubjects)
	subjects.Put("/:id", h.UpdateSubject)
	subjects.Delete("/:id", h.DeleteSubject)

	questions := app.Group("/api/questions")
	questions.Post("/", h.CreateQuestion)
	questions.Get("/", h.GetQuestions)
	questions.Put("/:id", h.UpdateQuestion)
	questions.Delete("/:id", h.DeleteQuestion)

	exams := app.Group("/api/exams")
	exams.Post("/", h.CreateExam)
	exams.Get("/", h.GetExams)
	exams.Get("/:id", h.GetExam)
	exams.Put("/:id", h.UpdateExam)
	exams.Delete("/:id", h.DeleteExam)
	exams.Post("/:id/publish", h.PublishExam)
	exams.Post("/:id/questions", h.AddQuestions)
	exams.Delete("/:id/questions/:questionId", h.RemoveQuestion)
	exams.Get("/:id/attempts", h.GetAttemptsByExam)
	exams.Get("/:id/results/export", h.ExportExamResults)

	attempts := app.Group("/api/attempts")
	attempts.Post("/exams/:examId/start", h.StartAttempt)
	attempts.Post("/:attemptId/submit", h.SubmitAttempt)
	attempts.Post("/:attemptId/save", h.SaveProgress)
	attempts.Get("/:attemptId/resume", h.ResumeAttempt)
	attempts.Get("/:attemptId/review", h.ReviewAttempt)
	attempts.Post("/:attemptId/tab-switch", h.LogTabSwitch)
	attempts.Post("/:attemptId/grade", h.GradeEssay)
	attempts.Get("/me", h.GetMyAttempts)

	classes := app.Group("/api/classes")
	classes.Get("/", h.GetClasses)
	classes.Post("/", h.CreateClass)

	dashboard := app.Group("/api/dashboard")
	dashboard.Get("/stats", h.DashboardStats)
	dashboard.Get("/exams/:examId/stats", h.ExamStats)

	handlers.InitWS(app, h)

	log.Println("Exam service running on :3002")
	log.Fatal(app.Listen(":" + getPort("3002")))
}

func getPort(def string) string {
	if p := os.Getenv("EXAM_PORT"); p != "" {
		return p
	}
	return def
}

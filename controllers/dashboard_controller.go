package controllers

import (
	"github.com/Satria-CBT/satria-cbt-server/models"
	"github.com/Satria-CBT/satria-cbt-server/res"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

type DashboardController struct {
	db *gorm.DB
}

func NewDashboardController(db *gorm.DB) *DashboardController {
	return &DashboardController{db: db}
}

func (d *DashboardController) Stats(c *fiber.Ctx) error {
	claims := c.Locals("userID").(jwt.MapClaims)
	role := claims["role"].(string)
	userID := uint(claims["id"].(float64))

	var stats res.DashboardStatsResponse

	if role == "teacher" {
		d.db.Model(&models.Exam{}).Where("created_by_id = ?", userID).Count(&stats.TotalExams)
		d.db.Model(&models.Subject{}).Count(&stats.TotalSubjects)
		d.db.Model(&models.ExamAttempt{}).
			Joins("JOIN exams ON exams.id = exam_attempts.exam_id").
			Where("exams.created_by_id = ?", userID).
			Count(&stats.TotalAttempts)
		d.db.Model(&models.ExamAttempt{}).
			Joins("JOIN exams ON exams.id = exam_attempts.exam_id").
			Where("exams.created_by_id = ? AND exam_attempts.status = ?", userID, models.AttemptInProgress).
			Count(&stats.OngoingAttempts)
	} else {
		d.db.Model(&models.Exam{}).Count(&stats.TotalExams)
		d.db.Model(&models.Subject{}).Count(&stats.TotalSubjects)
		d.db.Model(&models.Students{}).Count(&stats.TotalStudents)
		d.db.Model(&models.Teachers{}).Count(&stats.TotalTeachers)
		d.db.Model(&models.ExamAttempt{}).Count(&stats.TotalAttempts)
		d.db.Model(&models.ExamAttempt{}).Where("status = ?", models.AttemptInProgress).Count(&stats.OngoingAttempts)
	}

	return c.JSON(res.ResponseCode{Code: fiber.StatusOK, Message: "Dashboard stats retrieved", Data: stats})
}

func (d *DashboardController) ExamStats(c *fiber.Ctx) error {
	examID := c.Params("examId")

	var result []struct {
		Score int
	}
	d.db.Model(&models.ExamAttempt{}).Select("COALESCE(score, 0) as score").Where("exam_id = ? AND status != ?", examID, models.AttemptInProgress).Scan(&result)

	var exam models.Exam
	if err := d.db.First(&exam, examID).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusNotFound, Message: "Exam not found"}
	}

	if len(result) == 0 {
		return c.JSON(res.ResponseCode{Code: fiber.StatusOK, Message: "No attempts yet", Data: res.ExamStatResponse{
			ExamID:    exam.ID,
			ExamTitle: exam.Title,
		}})
	}

	var totalScore int
	highest := 0
	lowest := -1
	passCount := 0

	for _, r := range result {
		totalScore += r.Score
		if r.Score > highest {
			highest = r.Score
		}
		if lowest == -1 || r.Score < lowest {
			lowest = r.Score
		}
		if exam.PassingScore > 0 && r.Score >= exam.PassingScore {
			passCount++
		}
	}

	avgScore := float64(totalScore) / float64(len(result))
	failCount := len(result) - passCount
	passRate := 0.0
	if len(result) > 0 {
		passRate = float64(passCount) / float64(len(result)) * 100
	}

	return c.JSON(res.ResponseCode{Code: fiber.StatusOK, Message: "Exam stats retrieved", Data: res.ExamStatResponse{
		ExamID:        exam.ID,
		ExamTitle:     exam.Title,
		TotalAttempts: len(result),
		AverageScore:  avgScore,
		HighestScore:  highest,
		LowestScore:   lowest,
		PassCount:     passCount,
		FailCount:     failCount,
		PassRate:      passRate,
	}})
}

func (d *DashboardController) StudentPerformance(c *fiber.Ctx) error {
	studentID := c.Params("studentId")

	var attempts []models.ExamAttempt
	if err := d.db.Where("student_id = ? AND status != ?", studentID, models.AttemptInProgress).Find(&attempts).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusInternalServerError, Message: err.Error()}
	}

	var student models.Students
	d.db.First(&student, studentID)

	if len(attempts) == 0 {
		return c.JSON(res.ResponseCode{Code: fiber.StatusOK, Message: "No attempts yet", Data: res.StudentPerformanceResponse{
			StudentID:   student.ID,
			StudentName: student.Name,
		}})
	}

	totalScore := 0
	highest := 0
	lowest := -1
	passCount := 0

	for _, a := range attempts {
		if a.Score != nil {
			s := *a.Score
			totalScore += s
			if s > highest {
				highest = s
			}
			if lowest == -1 || s < lowest {
				lowest = s
			}
			var exam models.Exam
			d.db.First(&exam, a.ExamID)
			if exam.PassingScore > 0 && s >= exam.PassingScore {
				passCount++
			}
		}
	}

	avgScore := float64(totalScore) / float64(len(attempts))
	failCount := len(attempts) - passCount

	return c.JSON(res.ResponseCode{Code: fiber.StatusOK, Message: "Student performance retrieved", Data: res.StudentPerformanceResponse{
		StudentID:    student.ID,
		StudentName:  student.Name,
		TotalExams:   len(attempts),
		AverageScore: avgScore,
		HighestScore: highest,
		LowestScore:  lowest,
		TotalPass:    passCount,
		TotalFail:    failCount,
	}})
}

func (d *DashboardController) ClassPerformance(c *fiber.Ctx) error {
	classID := c.Params("classId")

	var class models.Class
	if err := d.db.Preload("Students").First(&class, classID).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusNotFound, Message: "Class not found"}
	}

	studentIDs := make([]uint, len(class.Students))
	for i, s := range class.Students {
		studentIDs[i] = s.ID
	}

	var attempts []models.ExamAttempt
	d.db.Where("student_id IN ? AND status != ?", studentIDs, models.AttemptInProgress).Find(&attempts)

	if len(attempts) == 0 {
		return c.JSON(res.ResponseCode{Code: fiber.StatusOK, Message: "No attempts yet", Data: res.ClassPerformanceResponse{
			ClassID:       class.ID,
			ClassName:     class.Name,
			TotalStudents: len(class.Students),
		}})
	}

	totalScore := 0
	passCount := 0
	for _, a := range attempts {
		if a.Score != nil {
			totalScore += *a.Score
			var exam models.Exam
			d.db.First(&exam, a.ExamID)
			if exam.PassingScore > 0 && *a.Score >= exam.PassingScore {
				passCount++
			}
		}
	}
	avgScore := float64(totalScore) / float64(len(attempts))
	failCount := len(attempts) - passCount

	return c.JSON(res.ResponseCode{Code: fiber.StatusOK, Message: "Class performance retrieved", Data: res.ClassPerformanceResponse{
		ClassID:       class.ID,
		ClassName:     class.Name,
		TotalStudents: len(class.Students),
		TotalAttempts: len(attempts),
		AverageScore:  avgScore,
		PassCount:     passCount,
		FailCount:     failCount,
	}})
}

func (d *DashboardController) RecentActivity(c *fiber.Ctx) error {
	var attempts []models.ExamAttempt
	query := d.db.Preload("Exam").Preload("Student").Order("updated_at DESC").Limit(10)

	claims := c.Locals("userID").(jwt.MapClaims)
	if claims["role"].(string) == "teacher" {
		teacherID := uint(claims["id"].(float64))
		query = query.Joins("JOIN exams ON exams.id = exam_attempts.exam_id").
			Where("exams.created_by_id = ?", teacherID)
	}

	if err := query.Find(&attempts).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusInternalServerError, Message: err.Error()}
	}

	var response []res.RecentAttemptResponse
	for _, a := range attempts {
		submittedAt := a.EndTime
		response = append(response, res.RecentAttemptResponse{
			AttemptID:   a.ID,
			ExamTitle:   a.Exam.Title,
			StudentName: a.Student.Name,
			Score:       a.Score,
			Status:      string(a.Status),
			SubmittedAt: submittedAt,
		})
	}

	return c.JSON(res.ResponseCode{Code: fiber.StatusOK, Message: "Recent activity retrieved", Data: response})
}

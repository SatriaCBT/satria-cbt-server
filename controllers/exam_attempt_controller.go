package controllers

import (
	"time"

	"github.com/Satria-CBT/satria-cbt-server/models"
	"github.com/Satria-CBT/satria-cbt-server/res"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

type ExamAttemptController struct {
	db *gorm.DB
}

func NewExamAttemptController(db *gorm.DB) *ExamAttemptController {
	return &ExamAttemptController{db: db}
}

func (e *ExamAttemptController) Start(c *fiber.Ctx) error {
	examID := c.Params("examId")
	studentID, err := getUserID(c)
	if err != nil {
		return err
	}

	var exam models.Exam
	if err := e.db.First(&exam, examID).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusNotFound, Message: "Exam not found"}
	}

	if exam.Status != models.ExamPublished {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: "Exam is not available"}
	}

	if exam.StartTime != nil && time.Now().Before(*exam.StartTime) {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: "Exam has not started yet"}
	}
	if exam.EndTime != nil && time.Now().After(*exam.EndTime) {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: "Exam has ended"}
	}

	var attemptCount int64
	e.db.Model(&models.ExamAttempt{}).Where("exam_id = ? AND student_id = ?", examID, studentID).Count(&attemptCount)
	if int(attemptCount) >= exam.MaxAttempts {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: "Maximum attempts reached"}
	}

	attempt := models.ExamAttempt{
		ExamID:    exam.ID,
		StudentID: studentID,
		StartTime: time.Now(),
		Status:    models.AttemptInProgress,
	}

	if err := e.db.Create(&attempt).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusInternalServerError, Message: err.Error()}
	}

	var examQuestions []struct {
		QuestionID uint
	}
	if exam.ShuffleQuestions {
		e.db.Raw("SELECT question_id FROM exam_questions WHERE exam_id = ? ORDER BY RANDOM()", examID).Scan(&examQuestions)
	} else {
		e.db.Raw("SELECT question_id FROM exam_questions WHERE exam_id = ? ORDER BY order_index ASC", examID).Scan(&examQuestions)
	}

	for _, eq := range examQuestions {
		e.db.Create(&models.ExamAnswer{
			AttemptID:  attempt.ID,
			QuestionID: eq.QuestionID,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(res.ResponseCode{
		Code:    fiber.StatusCreated,
		Message: "Exam started successfully",
		Data:    toAttemptResponse(attempt),
	})
}

func (e *ExamAttemptController) Submit(c *fiber.Ctx) error {
	attemptID := c.Params("attemptId")
	studentID, err := getUserID(c)
	if err != nil {
		return err
	}

	var req struct {
		Answers []struct {
			QuestionID uint   `json:"questionId"`
			Answer     string `json:"answer"`
		} `json:"answers"`
	}
	if err := c.BodyParser(&req); err != nil {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: err.Error()}
	}

	var attempt models.ExamAttempt
	if err := e.db.Preload("Exam").First(&attempt, attemptID).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusNotFound, Message: "Attempt not found"}
	}

	if attempt.StudentID != studentID {
		return &fiber.Error{Code: fiber.StatusForbidden, Message: "Unauthorized to submit this attempt"}
	}
	if attempt.Status != models.AttemptInProgress {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: "Attempt is already completed"}
	}

	elapsed := time.Since(attempt.StartTime)
	if elapsed > time.Duration(attempt.Exam.Duration)*time.Minute {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: "Time is up"}
	}

	totalCorrect := 0
	totalWrong := 0
	totalScore := 0

	for _, answer := range req.Answers {
		var question models.Question
		if err := e.db.First(&question, answer.QuestionID).Error; err != nil {
			continue
		}

		isCorrect := false
		if question.Type == models.QuestionMultipleChoice || question.Type == models.QuestionTrueFalse {
			if question.CorrectAnswer == answer.Answer {
				isCorrect = true
			}
		}

		points := 0
		if isCorrect {
			points = question.Points
			totalCorrect++
		} else {
			totalWrong++
		}
		totalScore += points

		e.db.Model(&models.ExamAnswer{}).Where("attempt_id = ? AND question_id = ?", attemptID, answer.QuestionID).Updates(map[string]interface{}{
			"answer":     answer.Answer,
			"is_correct": isCorrect,
			"points":     points,
		})
	}

	now := time.Now()
	score := totalScore
	attempt.EndTime = &now
	attempt.Score = &score
	attempt.TotalPoints = attempt.Exam.TotalPoints
	attempt.TotalCorrect = totalCorrect
	attempt.TotalWrong = totalWrong
	attempt.Status = models.AttemptCompleted

	if err := e.db.Save(&attempt).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusInternalServerError, Message: err.Error()}
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Exam submitted successfully",
		Data:    toAttemptResponse(attempt),
	})
}

func (e *ExamAttemptController) GetAttempt(c *fiber.Ctx) error {
	attemptID := c.Params("attemptId")
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	claims := c.Locals("userID").(jwt.MapClaims)
	role := claims["role"].(string)

	var attempt models.ExamAttempt
	query := e.db.Preload("Exam").Preload("Answers")
	if err := query.First(&attempt, attemptID).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusNotFound, Message: "Attempt not found"}
	}

	if role == "student" && attempt.StudentID != userID {
		return &fiber.Error{Code: fiber.StatusForbidden, Message: "Unauthorized"}
	}

	answers := make([]res.ExamAnswerResponse, len(attempt.Answers))
	for i, ans := range attempt.Answers {
		answerResp := res.ExamAnswerResponse{
			ID:         ans.ID,
			AttemptID:  ans.AttemptID,
			QuestionID: ans.QuestionID,
			Answer:     ans.Answer,
		}

		if attempt.Status == models.AttemptCompleted || attempt.Status == models.AttemptGraded {
			answerResp.IsCorrect = ans.IsCorrect
		}

		answers[i] = answerResp
	}

	response := toAttemptResponse(attempt)
	response.Answers = answers

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Attempt retrieved successfully",
		Data:    response,
	})
}

func (e *ExamAttemptController) GetAttemptsByExam(c *fiber.Ctx) error {
	examID := c.Params("examId")

	var attempts []models.ExamAttempt
	if err := e.db.Preload("Student").Where("exam_id = ?", examID).Find(&attempts).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusInternalServerError, Message: err.Error()}
	}

	var response []res.ExamAttemptResponse
	for _, a := range attempts {
		response = append(response, toAttemptResponse(a))
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Attempts retrieved successfully",
		Data:    response,
	})
}

func (e *ExamAttemptController) GetMyAttempts(c *fiber.Ctx) error {
	studentID, err := getUserID(c)
	if err != nil {
		return err
	}

	var attempts []models.ExamAttempt
	if err := e.db.Preload("Exam").Where("student_id = ?", studentID).Order("created_at DESC").Find(&attempts).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusInternalServerError, Message: err.Error()}
	}

	var response []res.ExamAttemptResponse
	for _, a := range attempts {
		response = append(response, toAttemptResponse(a))
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "My attempts retrieved successfully",
		Data:    response,
	})
}

func (e *ExamAttemptController) GradeEssay(c *fiber.Ctx) error {
	attemptID := c.Params("attemptId")

	var req struct {
		QuestionID uint `json:"questionId"`
		Points     int  `json:"points"`
	}
	if err := c.BodyParser(&req); err != nil {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: err.Error()}
	}

	if err := e.db.Model(&models.ExamAnswer{}).Where("attempt_id = ? AND question_id = ?", attemptID, req.QuestionID).Update("points", req.Points).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusInternalServerError, Message: err.Error()}
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Essay graded successfully",
	})
}

func toAttemptResponse(a models.ExamAttempt) res.ExamAttemptResponse {
	return res.ExamAttemptResponse{
		ID:           a.ID,
		ExamID:       a.ExamID,
		StudentID:    a.StudentID,
		StartTime:    a.StartTime,
		EndTime:      a.EndTime,
		Score:        a.Score,
		TotalPoints:  a.TotalPoints,
		TotalCorrect: a.TotalCorrect,
		TotalWrong:   a.TotalWrong,
		Status:       string(a.Status),
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
	}
}

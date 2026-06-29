package controllers

import (
	"time"

	"github.com/Satria-CBT/satria-cbt-server/models"
	"github.com/Satria-CBT/satria-cbt-server/res"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

type ExamController struct {
	db *gorm.DB
}

func NewExamController(db *gorm.DB) *ExamController {
	return &ExamController{db: db}
}

func (e *ExamController) Create(c *fiber.Ctx) error {
	var req models.Exam
	if err := c.BodyParser(&req); err != nil {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: err.Error()}
	}

	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	if req.Title == "" || req.Duration == 0 {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: "Title and duration are required"}
	}
	if req.SubjectID == 0 {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: "Subject ID is required"}
	}

	req.CreatedByID = userID
	req.Status = models.ExamDraft

	if err := e.db.Create(&req).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusInternalServerError, Message: err.Error()}
	}

	return c.Status(fiber.StatusCreated).JSON(res.ResponseCode{
		Code:    fiber.StatusCreated,
		Message: "Exam created successfully",
		Data:    toExamResponse(req),
	})
}

func (e *ExamController) GetAll(c *fiber.Ctx) error {
	claims := c.Locals("userID").(jwt.MapClaims)
	role := claims["role"].(string)

	var exams []models.Exam
	query := e.db.Preload("Subject")

	if role == "teacher" {
		teacherID := claims["id"].(float64)
		query = query.Where("created_by_id = ?", teacherID)
	} else if role == "student" {
		studentID := claims["id"].(float64)
		var student models.Students
		if err := e.db.Preload("Classes").First(&student, studentID).Error; err != nil {
			return &fiber.Error{Code: fiber.StatusNotFound, Message: "Student not found"}
		}

		classIDs := make([]uint, len(student.Classes))
		for i, cls := range student.Classes {
			classIDs[i] = cls.ID
		}

		query = query.Where("status = ? AND class_id IN ? AND (start_time IS NULL OR start_time <= ?) AND (end_time IS NULL OR end_time >= ?)",
			models.ExamPublished, classIDs, time.Now(), time.Now())
	}

	if err := query.Find(&exams).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusInternalServerError, Message: err.Error()}
	}

	var response []res.ExamResponse
	for _, exam := range exams {
		resp := toExamResponse(exam)
		resp.QuestionCount = int(e.db.Model(&exam).Association("Questions").Count())
		response = append(response, resp)
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Exams retrieved successfully",
		Data:    response,
	})
}

func (e *ExamController) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")

	var exam models.Exam
	if err := e.db.Preload("Subject").Preload("Questions").First(&exam, id).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusNotFound, Message: "Exam not found"}
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Exam retrieved successfully",
		Data:    toExamDetailResponse(exam),
	})
}

func (e *ExamController) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	var req models.Exam
	if err := c.BodyParser(&req); err != nil {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: err.Error()}
	}

	var exam models.Exam
	if err := e.db.First(&exam, id).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusNotFound, Message: "Exam not found"}
	}

	if exam.Status != models.ExamDraft {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: "Can only update draft exams"}
	}

	userID, err := getUserID(c)
	if err != nil {
		return err
	}
	if exam.CreatedByID != userID {
		return &fiber.Error{Code: fiber.StatusForbidden, Message: "Unauthorized to update this exam"}
	}

	if req.Title != "" {
		exam.Title = req.Title
	}
	if req.Duration > 0 {
		exam.Duration = req.Duration
	}
	if req.SubjectID > 0 {
		exam.SubjectID = req.SubjectID
	}
	if req.Description != "" {
		exam.Description = req.Description
	}
	if req.PassingScore > 0 {
		exam.PassingScore = req.PassingScore
	}
	if req.ClassID > 0 {
		exam.ClassID = req.ClassID
	}
	exam.ShuffleQuestions = req.ShuffleQuestions
	exam.ShowResult = req.ShowResult
	exam.StartTime = req.StartTime
	exam.EndTime = req.EndTime

	if err := e.db.Save(&exam).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusInternalServerError, Message: err.Error()}
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Exam updated successfully",
		Data:    toExamResponse(exam),
	})
}

func (e *ExamController) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	var exam models.Exam
	if err := e.db.First(&exam, id).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusNotFound, Message: "Exam not found"}
	}

	if exam.Status != models.ExamDraft {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: "Can only delete draft exams"}
	}

	if err := e.db.Delete(&exam).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusInternalServerError, Message: err.Error()}
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Exam deleted successfully",
	})
}

func (e *ExamController) Publish(c *fiber.Ctx) error {
	id := c.Params("id")

	var exam models.Exam
	if err := e.db.Preload("Questions").First(&exam, id).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusNotFound, Message: "Exam not found"}
	}

	if exam.Status != models.ExamDraft {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: "Exam is not in draft status"}
	}

	if len(exam.Questions) == 0 {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: "Exam must have at least one question"}
	}

	var totalPoints int
	for _, q := range exam.Questions {
		totalPoints += q.Points
	}
	exam.TotalPoints = totalPoints
	exam.Status = models.ExamPublished

	if exam.StartTime == nil {
		now := time.Now()
		exam.StartTime = &now
	}

	if err := e.db.Save(&exam).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusInternalServerError, Message: err.Error()}
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Exam published successfully",
		Data:    toExamResponse(exam),
	})
}

func (e *ExamController) AddQuestions(c *fiber.Ctx) error {
	examID := c.Params("id")

	var req struct {
		QuestionIDs []uint `json:"questionIds"`
	}
	if err := c.BodyParser(&req); err != nil {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: err.Error()}
	}

	var exam models.Exam
	if err := e.db.First(&exam, examID).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusNotFound, Message: "Exam not found"}
	}

	if exam.Status != models.ExamDraft {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: "Can only modify draft exams"}
	}

	var questions []models.Question
	if err := e.db.Where("id IN ?", req.QuestionIDs).Find(&questions).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusInternalServerError, Message: err.Error()}
	}

	if err := e.db.Model(&exam).Association("Questions").Append(questions); err != nil {
		return &fiber.Error{Code: fiber.StatusInternalServerError, Message: err.Error()}
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Questions added to exam successfully",
	})
}

func (e *ExamController) RemoveQuestion(c *fiber.Ctx) error {
	examID := c.Params("id")
	questionID := c.Params("questionId")

	var exam models.Exam
	if err := e.db.First(&exam, examID).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusNotFound, Message: "Exam not found"}
	}

	if exam.Status != models.ExamDraft {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: "Can only modify draft exams"}
	}

	if err := e.db.Model(&exam).Association("Questions").Delete(&models.Question{ID: parseUint(questionID)}); err != nil {
		return &fiber.Error{Code: fiber.StatusInternalServerError, Message: err.Error()}
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Question removed from exam successfully",
	})
}

func toExamResponse(exam models.Exam) res.ExamResponse {
	return res.ExamResponse{
		ID:               exam.ID,
		Title:            exam.Title,
		Description:      exam.Description,
		SubjectID:        exam.SubjectID,
		Duration:         exam.Duration,
		StartTime:        exam.StartTime,
		EndTime:          exam.EndTime,
		TotalPoints:      exam.TotalPoints,
		PassingScore:     exam.PassingScore,
		ShuffleQuestions: exam.ShuffleQuestions,
		ShowResult:       exam.ShowResult,
		MaxAttempts:      exam.MaxAttempts,
		Status:           string(exam.Status),
		CreatedByID:      exam.CreatedByID,
		ClassID:          exam.ClassID,
		CreatedAt:        exam.CreatedAt,
		UpdatedAt:        exam.UpdatedAt,
	}
}

func toExamDetailResponse(exam models.Exam) res.ExamDetailResponse {
	base := toExamResponse(exam)
	questions := make([]res.QuestionResponseWithAnswer, len(exam.Questions))
	for i, q := range exam.Questions {
		questions[i] = toQuestionResponseWithAnswer(q)
	}
	return res.ExamDetailResponse{
		ExamResponse: base,
		Questions:    questions,
	}
}

func parseUint(s string) uint {
	var id uint
	for _, c := range s {
		if c >= '0' && c <= '9' {
			id = id*10 + uint(c-'0')
		}
	}
	return id
}

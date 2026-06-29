package controllers

import (
	"github.com/Satria-CBT/satria-cbt-server/models"
	"github.com/Satria-CBT/satria-cbt-server/res"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

type QuestionController struct {
	db *gorm.DB
}

func NewQuestionController(db *gorm.DB) *QuestionController {
	return &QuestionController{db: db}
}

func (q *QuestionController) Create(c *fiber.Ctx) error {
	var req models.Question
	if err := c.BodyParser(&req); err != nil {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: err.Error()}
	}

	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	if req.Question == "" || req.CorrectAnswer == "" {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: "Question and correct answer are required"}
	}
	if req.SubjectID == 0 {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: "Subject ID is required"}
	}
	if req.Points == 0 {
		req.Points = 1
	}

	req.CreatedByID = userID

	if err := q.db.Create(&req).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusInternalServerError, Message: err.Error()}
	}

	return c.Status(fiber.StatusCreated).JSON(res.ResponseCode{
		Code:    fiber.StatusCreated,
		Message: "Question created successfully",
		Data:    toQuestionResponse(req),
	})
}

func (q *QuestionController) GetAll(c *fiber.Ctx) error {
	subjectID := c.Query("subjectId")

	var questions []models.Question
	query := q.db.Preload("Subject")

	if subjectID != "" {
		query = query.Where("subject_id = ?", subjectID)
	}

	claims := c.Locals("userID").(jwt.MapClaims)
	role := claims["role"].(string)

	if role == "teacher" {
		teacherID := claims["id"].(float64)
		query = query.Where("created_by_id = ?", teacherID)
	}

	if err := query.Find(&questions).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusInternalServerError, Message: err.Error()}
	}

	var response []res.QuestionResponseWithAnswer
	for _, qs := range questions {
		response = append(response, toQuestionResponseWithAnswer(qs))
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Questions retrieved successfully",
		Data:    response,
	})
}

func (q *QuestionController) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")

	var question models.Question
	if err := q.db.Preload("Subject").First(&question, id).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusNotFound, Message: "Question not found"}
	}

	claims := c.Locals("userID").(jwt.MapClaims)
	role := claims["role"].(string)

	if role == "student" {
		return c.JSON(res.ResponseCode{
			Code:    fiber.StatusOK,
			Message: "Question retrieved successfully",
			Data:    toQuestionResponse(question),
		})
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Question retrieved successfully",
		Data:    toQuestionResponseWithAnswer(question),
	})
}

func (q *QuestionController) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	var req models.Question
	if err := c.BodyParser(&req); err != nil {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: err.Error()}
	}

	var question models.Question
	if err := q.db.First(&question, id).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusNotFound, Message: "Question not found"}
	}

	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	if question.CreatedByID != userID {
		return &fiber.Error{Code: fiber.StatusForbidden, Message: "Unauthorized to update this question"}
	}

	if req.Question != "" {
		question.Question = req.Question
	}
	if req.CorrectAnswer != "" {
		question.CorrectAnswer = req.CorrectAnswer
	}
	if req.Type != "" {
		question.Type = req.Type
	}
	if req.Options != "" {
		question.Options = req.Options
	}
	if req.Points > 0 {
		question.Points = req.Points
	}
	if req.SubjectID > 0 {
		question.SubjectID = req.SubjectID
	}

	if err := q.db.Save(&question).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusInternalServerError, Message: err.Error()}
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Question updated successfully",
		Data:    toQuestionResponseWithAnswer(question),
	})
}

func (q *QuestionController) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	var question models.Question
	if err := q.db.First(&question, id).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusNotFound, Message: "Question not found"}
	}

	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	if question.CreatedByID != userID {
		return &fiber.Error{Code: fiber.StatusForbidden, Message: "Unauthorized to delete this question"}
	}

	if err := q.db.Delete(&question).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusInternalServerError, Message: err.Error()}
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Question deleted successfully",
	})
}

func toQuestionResponse(q models.Question) res.QuestionResponse {
	return res.QuestionResponse{
		ID:        q.ID,
		SubjectID: q.SubjectID,
		Type:      string(q.Type),
		Question:  q.Question,
		Options:   q.Options,
		Points:    q.Points,
		CreatedAt: q.CreatedAt,
		UpdatedAt: q.UpdatedAt,
	}
}

func toQuestionResponseWithAnswer(q models.Question) res.QuestionResponseWithAnswer {
	return res.QuestionResponseWithAnswer{
		ID:            q.ID,
		SubjectID:     q.SubjectID,
		Type:          string(q.Type),
		Question:      q.Question,
		Options:       q.Options,
		CorrectAnswer: q.CorrectAnswer,
		Points:        q.Points,
		CreatedByID:   q.CreatedByID,
		CreatedAt:     q.CreatedAt,
		UpdatedAt:     q.UpdatedAt,
	}
}

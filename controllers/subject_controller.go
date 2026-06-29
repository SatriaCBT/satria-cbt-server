package controllers

import (
	"github.com/Satria-CBT/satria-cbt-server/models"
	"github.com/Satria-CBT/satria-cbt-server/res"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type SubjectController struct {
	db *gorm.DB
}

func NewSubjectController(db *gorm.DB) *SubjectController {
	return &SubjectController{db: db}
}

func (s *SubjectController) Create(c *fiber.Ctx) error {
	var req models.Subject
	if err := c.BodyParser(&req); err != nil {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: err.Error()}
	}

	if req.Name == "" || req.Code == "" {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: "Name and code are required"}
	}

	if err := s.db.Create(&req).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusInternalServerError, Message: err.Error()}
	}

	return c.Status(fiber.StatusCreated).JSON(res.ResponseCode{
		Code:    fiber.StatusCreated,
		Message: "Subject created successfully",
		Data:    toSubjectResponse(req),
	})
}

func (s *SubjectController) GetAll(c *fiber.Ctx) error {
	var subjects []models.Subject
	if err := s.db.Find(&subjects).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusInternalServerError, Message: err.Error()}
	}

	var response []res.SubjectResponse
	for _, sub := range subjects {
		response = append(response, toSubjectResponse(sub))
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Subjects retrieved successfully",
		Data:    response,
	})
}

func (s *SubjectController) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	var subject models.Subject
	if err := s.db.First(&subject, id).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusNotFound, Message: "Subject not found"}
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Subject retrieved successfully",
		Data:    toSubjectResponse(subject),
	})
}

func (s *SubjectController) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	var req models.Subject
	if err := c.BodyParser(&req); err != nil {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: err.Error()}
	}

	var subject models.Subject
	if err := s.db.First(&subject, id).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusNotFound, Message: "Subject not found"}
	}

	if req.Name != "" {
		subject.Name = req.Name
	}
	if req.Code != "" {
		subject.Code = req.Code
	}
	if req.Description != "" {
		subject.Description = req.Description
	}

	if err := s.db.Save(&subject).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusInternalServerError, Message: err.Error()}
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Subject updated successfully",
		Data:    toSubjectResponse(subject),
	})
}

func (s *SubjectController) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	var subject models.Subject
	if err := s.db.First(&subject, id).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusNotFound, Message: "Subject not found"}
	}

	if err := s.db.Delete(&subject).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusInternalServerError, Message: err.Error()}
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Subject deleted successfully",
	})
}

func toSubjectResponse(s models.Subject) res.SubjectResponse {
	return res.SubjectResponse{
		ID:          s.ID,
		Name:        s.Name,
		Code:        s.Code,
		Description: s.Description,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}

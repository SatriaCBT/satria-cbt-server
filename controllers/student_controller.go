package controllers

import (
	"os"
	"time"

	"github.com/Satria-CBT/satria-cbt-server/models"
	"github.com/Satria-CBT/satria-cbt-server/res"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type StudentController struct {
	db *gorm.DB
}

func NewStudentController(db *gorm.DB) *StudentController {
	return &StudentController{db: db}
}

func (s *StudentController) RegisterStudent(c *fiber.Ctx) error {
	var req models.Students
	if err := c.BodyParser(&req); err != nil {
		return &fiber.Error{
			Code:    fiber.StatusBadRequest,
			Message: err.Error(),
		}
	}

	if !isValidEmail(req.Email) {
		return &fiber.Error{
			Code:    fiber.StatusBadRequest,
			Message: "Invalid email format",
		}
	}

	if err := validatePassword(req.Password); err != nil {
		return err
	}

	encoded, err := hashPassword(req.Password)
	if err != nil {
		return err
	}

	student := models.Students{
		Name:      req.Name,
		Username:  req.Username,
		Email:     req.Email,
		Password:  encoded,
		Classes:   req.Classes,
		CreatedAt: time.Now(),
	}

	if err := s.db.Create(&student).Error; err != nil {
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	classSummaries := make([]res.ClassSummaryResponse, len(student.Classes))
	for i, class := range student.Classes {
		classSummaries[i] = res.ClassSummaryResponse{
			ID:   class.ID,
			Name: class.Name,
			Code: class.Code,
		}
	}

	response := res.StudentResponse{
		ID:        student.ID,
		Name:      student.Name,
		Username:  student.Username,
		Email:     student.Email,
		CreatedAt: student.CreatedAt,
		Classes:   classSummaries,
	}

	if student.UpdatedAt.IsZero() {
		response.UpdatedAt = nil
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Student registered successfully",
		Data:    response,
	})
}

func (s *StudentController) LoginStudent(c *fiber.Ctx) error {
	var req models.StudentsRequest
	if err := c.BodyParser(&req); err != nil {
		return &fiber.Error{
			Code:    fiber.StatusBadRequest,
			Message: err.Error(),
		}
	}

	var student models.Students
	if err := s.db.Where("email = ?", req.Email).Preload("Classes").First(&student).Error; err != nil {
		return &fiber.Error{
			Code:    fiber.StatusNotFound,
			Message: err.Error(),
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(student.Password), []byte(req.Password)); err != nil {
		return &fiber.Error{
			Code:    fiber.StatusUnauthorized,
			Message: "Invalid password",
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    student.ID,
		"email": student.Email,
		"role":  "student",
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	})

	secretKey := os.Getenv("ADMIN_TOKEN")
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(res.ResponseCode{
			Code:    fiber.StatusInternalServerError,
			Message: "Failed to generate token",
		})
	}

	classSummaries := make([]res.ClassSummaryResponse, len(student.Classes))
	for i, class := range student.Classes {
		classSummaries[i] = res.ClassSummaryResponse{
			ID:   class.ID,
			Name: class.Name,
			Code: class.Code,
		}
	}

	response := res.StudentLoginResponse{
		ID:        student.ID,
		Name:      student.Name,
		Username:  student.Username,
		Email:     student.Email,
		CreatedAt: student.CreatedAt,
		Classes:   classSummaries,
		Token:     tokenString,
	}

	if student.UpdatedAt.IsZero() {
		response.UpdatedAt = nil
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Student logged in successfully",
		Data:    response,
	})
}

func (s *StudentController) GetSessionProfileStudent(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	var student models.Students
	if err := s.db.Where("id = ?", userID).Preload("Classes").First(&student).Error; err != nil {
		return &fiber.Error{
			Code:    fiber.StatusNotFound,
			Message: err.Error(),
		}
	}

	classSummaries := make([]res.ClassSummaryResponse, len(student.Classes))
	for i, class := range student.Classes {
		classSummaries[i] = res.ClassSummaryResponse{
			ID:   class.ID,
			Name: class.Name,
			Code: class.Code,
		}
	}

	data := res.StudentResponse{
		ID:        student.ID,
		Name:      student.Name,
		Username:  student.Username,
		Email:     student.Email,
		CreatedAt: student.CreatedAt,
		Classes:   classSummaries,
	}

	if student.UpdatedAt.IsZero() {
		data.UpdatedAt = nil
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Profile fetched successfully",
		Data:    data,
	})
}

func (s *StudentController) UpdateStudent(c *fiber.Ctx) error {
	var req map[string]interface{}
	var id = c.Params("id")

	if err := c.BodyParser(&req); err != nil {
		return &fiber.Error{
			Code:    fiber.ErrBadRequest.Code,
			Message: err.Error(),
		}
	}

	allowfields := map[string]bool{
		"name":     true,
		"username": true,
		"password": true,
		"email":    true,
		"classes":  true,
	}

	for key := range req {
		if !allowfields[key] {
			delete(req, key)
		}
	}

	var student models.Students
	if err := s.db.Where("id = ?", id).First(&student).Error; err != nil {
		return &fiber.Error{
			Code:    fiber.StatusNotFound,
			Message: err.Error(),
		}
	}

	if name, ok := req["name"].(string); ok && name != "" {
		student.Name = name
	}

	if username, ok := req["username"].(string); ok && username != "" {
		student.Username = username
	}

	if password, ok := req["password"].(string); ok && password != "" {
		if err := validatePassword(password); err != nil {
			return err
		}
		encoded, err := hashPassword(password)
		if err != nil {
			return err
		}
		student.Password = encoded
	}

	if email, ok := req["email"].(string); ok && email != "" {
		if !isValidEmail(email) {
			return &fiber.Error{
				Code:    fiber.StatusBadRequest,
				Message: "Invalid email format",
			}
		}
		student.Email = email
	}

	if classes, ok := req["classes"].([]interface{}); ok && len(classes) > 0 {
		var parsedClasses []models.Class
		for _, c := range classes {
			if classMap, isMap := c.(map[string]interface{}); isMap {
				if id, exists := classMap["id"]; exists {
					parsedClasses = append(parsedClasses, models.Class{ID: id.(uint)})
				}
			}
		}
		student.Classes = parsedClasses
	}

	student.UpdatedAt = time.Now()

	if err := s.db.Save(&student).Error; err != nil {
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	classSummaries := make([]res.ClassSummaryResponse, len(student.Classes))
	for i, class := range student.Classes {
		classSummaries[i] = res.ClassSummaryResponse{
			ID:   class.ID,
			Name: class.Name,
			Code: class.Code,
		}
	}

	response := res.StudentResponse{
		ID:        student.ID,
		Name:      student.Name,
		Username:  student.Username,
		Email:     student.Email,
		CreatedAt: student.CreatedAt,
		Classes:   classSummaries,
	}

	if student.UpdatedAt.IsZero() {
		response.UpdatedAt = nil
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Student updated successfully",
		Data:    response,
	})
}

func (s *StudentController) DeleteStudent(c *fiber.Ctx) error {
	userID := c.Params("id")

	var student models.Students
	if err := s.db.Where("id = ?", userID).First(&student).Error; err != nil {
		return &fiber.Error{
			Code:    fiber.StatusNotFound,
			Message: err.Error(),
		}
	}

	if err := s.db.Delete(&student).Error; err != nil {
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Student deleted successfully",
	})
}

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

type TeacherController struct {
	db *gorm.DB
}

func NewTeacherController(db *gorm.DB) *TeacherController {
	return &TeacherController{db: db}
}

func (t *TeacherController) RegisterTeacher(c *fiber.Ctx) error {
	var req models.Teachers
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

	teacher := models.Teachers{
		Name:      req.Name,
		Username:  req.Username,
		Email:     req.Email,
		Password:  encoded,
		Classes:   req.Classes,
		CreatedAt: time.Now(),
	}

	if err := t.db.Create(&teacher).Error; err != nil {
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	response := res.AdminResponse{
		ID:        teacher.ID,
		Name:      teacher.Name,
		Username:  teacher.Username,
		Email:     teacher.Email,
		CreatedAt: teacher.CreatedAt,
	}

	if teacher.UpdatedAt.IsZero() {
		response.UpdatedAt = nil
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Teacher registered successfully",
		Data:    response,
	})
}

func (t *TeacherController) LoginTeacher(c *fiber.Ctx) error {
	var req models.TeachersRequest
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

	var teacher models.Teachers
	if err := t.db.Where("email = ?", req.Email).First(&teacher).Error; err != nil {
		return &fiber.Error{
			Code:    fiber.StatusNotFound,
			Message: err.Error(),
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(teacher.Password), []byte(req.Password)); err != nil {
		return &fiber.Error{
			Code:    fiber.StatusUnauthorized,
			Message: "Invalid password",
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    teacher.ID,
		"email": teacher.Email,
		"role":  "teacher",
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

	response := res.TeacherLoginResponse{
		ID:        teacher.ID,
		Name:      teacher.Name,
		Username:  teacher.Username,
		Email:     teacher.Email,
		CreatedAt: teacher.CreatedAt,
		Token:     tokenString,
	}

	if teacher.UpdatedAt.IsZero() {
		response.UpdatedAt = nil
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Teacher logged in successfully",
		Data:    response,
	})
}

func (t *TeacherController) GetSessionProfileTeacher(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	var teacher models.Teachers
	if err := t.db.Where("id = ?", userID).Preload("Classes").First(&teacher).Error; err != nil {
		return &fiber.Error{
			Code:    fiber.StatusNotFound,
			Message: err.Error(),
		}
	}

	classSummaries := make([]res.ClassSummaryResponse, len(teacher.Classes))
	for i, class := range teacher.Classes {
		classSummaries[i] = res.ClassSummaryResponse{
			ID:   class.ID,
			Name: class.Name,
			Code: class.Code,
		}
	}

	data := res.TeacherResponse{
		ID:        teacher.ID,
		Name:      teacher.Name,
		Username:  teacher.Username,
		Email:     teacher.Email,
		CreatedAt: teacher.CreatedAt,
		Classes:   classSummaries,
	}

	if teacher.UpdatedAt.IsZero() {
		data.UpdatedAt = nil
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Profile fetched successfully",
		Data:    data,
	})
}

func (t *TeacherController) UpdateTeacher(c *fiber.Ctx) error {
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

	var teacher models.Teachers
	if err := t.db.Where("id = ?", id).First(&teacher).Error; err != nil {
		return &fiber.Error{
			Code:    fiber.StatusNotFound,
			Message: err.Error(),
		}
	}

	if name, ok := req["name"].(string); ok && name != "" {
		teacher.Name = name
	}

	if username, ok := req["username"].(string); ok && username != "" {
		teacher.Username = username
	}

	if password, ok := req["password"].(string); ok && password != "" {
		if err := validatePassword(password); err != nil {
			return err
		}
		encoded, err := hashPassword(password)
		if err != nil {
			return err
		}
		teacher.Password = encoded
	}

	if email, ok := req["email"].(string); ok && email != "" {
		if !isValidEmail(email) {
			return &fiber.Error{
				Code:    fiber.StatusBadRequest,
				Message: "Invalid email format",
			}
		}
		teacher.Email = email
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
		teacher.Classes = parsedClasses
	}

	teacher.UpdatedAt = time.Now()

	if err := t.db.Save(&teacher).Error; err != nil {
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	classSummaries := make([]res.ClassSummaryResponse, len(teacher.Classes))
	for i, class := range teacher.Classes {
		classSummaries[i] = res.ClassSummaryResponse{
			ID:   class.ID,
			Name: class.Name,
			Code: class.Code,
		}
	}

	response := res.TeacherResponse{
		ID:        teacher.ID,
		Name:      teacher.Name,
		Username:  teacher.Username,
		Email:     teacher.Email,
		CreatedAt: teacher.CreatedAt,
		Classes:   classSummaries,
	}

	if teacher.UpdatedAt.IsZero() {
		response.UpdatedAt = nil
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Teacher updated successfully",
		Data:    response,
	})
}

func (t *TeacherController) DeleteTeacher(c *fiber.Ctx) error {
	userID := c.Params("id")

	var teacher models.Teachers
	if err := t.db.Where("id = ?", userID).First(&teacher).Error; err != nil {
		return &fiber.Error{
			Code:    fiber.StatusNotFound,
			Message: err.Error(),
		}
	}

	if err := t.db.Delete(&teacher).Error; err != nil {
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Teacher deleted successfully",
	})
}

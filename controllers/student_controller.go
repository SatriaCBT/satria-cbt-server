package controllers

import (
	"os"
	"regexp"
	"satriacbtserver/configs"
	"satriacbtserver/models"
	"satriacbtserver/res"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type StudentController struct {}

func NewStudentController() *StudentController {
	return &StudentController{}
}


func (s *StudentController) RegisterStudent(c *fiber.Ctx) error {
	var req models.Students
	if err := c.BodyParser(&req); err != nil {
		return &fiber.Error{
			Code: fiber.StatusBadRequest,
			Message: err.Error(),
		}
	}

	if len(req.Password) <= 5 || len(req.Password) >= 12 {
		return &fiber.Error{
			Code: fiber.StatusBadRequest,
			Message: "Password must be between 5 and 12 characters",
		}
	}

	if !isValidEmail(req.Email) {
		return &fiber.Error{
			Code: fiber.StatusBadRequest,
			Message: "Invalid email format",
		}
	}

	hashletter := regexp.MustCompile(`[a-zA-Z]`).MatchString(req.Password)
	hashnumber := regexp.MustCompile(`\d`).MatchString(req.Password)
	hashspecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(req.Password)

	if !hashletter || !hashnumber || !hashspecial {
		return &fiber.Error{
			Code: fiber.StatusBadRequest,
			Message: "Password must contain at least one letter, one number, and one special character",
		}
	}

	encode, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return &fiber.Error{
			Code: fiber.StatusInternalServerError,
			Message: err.Error(),
		}
	}
	student := models.Students{
		Name: req.Name,
		Username: req.Username,
		Email: req.Email,
		Password: string(encode),
		Classes: req.Classes,
		CreatedAt: time.Now(),
	}

	err = configs.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&student).Error; err != nil {
			return &fiber.Error{
				Code: fiber.StatusInternalServerError,
				Message: err.Error(),
			}
		}
		return nil
	})

	if err != nil {
		return &fiber.Error{
			Code: fiber.StatusInternalServerError,
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
		Code: fiber.StatusOK,
		Message: "Student registered successfully",
		Data: response,
	})
	
}


func (s *StudentController) LoginStudent(c *fiber.Ctx) error {
	var req models.StudentsRequest
	if err := c.BodyParser(&req); err != nil {
		return &fiber.Error{
			Code: fiber.StatusBadRequest,
			Message: err.Error(),
		}
	}


	var student models.Students
	err := configs.Database().Transaction(func(tx *gorm.DB) error {
		result := tx.Where("email = ?", req.Email).Preload("Classes").First(&student)
		if result.Error != nil {
			return &fiber.Error{
				Code: fiber.StatusNotFound,
				Message: result.Error.Error(),
			}
		}
		return nil
	})

	if err != nil {
		return &fiber.Error{
			Code: fiber.StatusNotFound,
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
		Code: fiber.StatusOK,
		Message: "Student logged in successfully",
		Data: response,
	})
}


func (s *StudentController) GetSessionProfileStudent(c *fiber.Ctx) error {
	var response models.Students
	userID, ok := c.Locals("userID").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(res.ResponseCode{
			Code:  fiber.StatusUnauthorized,
			Message: "User ID not found in context",
		})
	}

	err := configs.Database().Transaction(func(tx *gorm.DB) error {
		result := tx.Where("id = ?", userID).First(&response)
		if result.Error != nil {
			return &fiber.Error{
				Code: fiber.StatusNotFound,
				Message: result.Error.Error(),
			}
		}
		return nil
	})

	if err != nil {
		return &fiber.Error{
			Code: fiber.StatusNotFound,
			Message: err.Error(),
		}
	}

	classSummaries := make([]res.ClassSummaryResponse, len(response.Classes))
	for i, class := range response.Classes {
		classSummaries[i] = res.ClassSummaryResponse{
			ID:   class.ID,
			Name: class.Name,
			Code: class.Code,
		}
	}

	data := res.StudentResponse{
		ID:        response.ID,
		Name:      response.Name,
		Username:  response.Username,
		Email:     response.Email,
		CreatedAt: response.CreatedAt,
		Classes:   classSummaries,
	}


	if response.UpdatedAt.IsZero() {
		data.UpdatedAt = nil
	}

	return c.JSON(res.ResponseCode{
		Code: fiber.StatusOK,
		Message: "Profile fetched successfully",
		Data: data,
	})
	
}


func (s *StudentController) UpdateStudent(c *fiber.Ctx) error {
	var req map[string]interface{}
	var id = c.Params("id")

	if err := c.BodyParser(&req); err != nil {
		return &fiber.Error{
			Code: fiber.ErrBadRequest.Code,
			Message: err.Error(),
		}
	}

	allowfields := map[string]bool{
		"name": true,
		"username": true,
		"password": true,
		"email": true,
		"classes": true,
	}

	for key := range req {
		if !allowfields[key] {
			delete(req, key)
		}
	}

	var student models.Students
	err := configs.Database().Transaction(func(tx *gorm.DB) error {
		result := tx.Where("id = ?", id).First(&student)
		if result.Error != nil {
			return &fiber.Error{
				Code: fiber.StatusNotFound,
				Message: result.Error.Error(),
			}
		}
		return nil
	})

	if err != nil {
		return &fiber.Error{
			Code: fiber.StatusNotFound,
			Message: err.Error(),
		}
	}

	if name, ok := req["name"].(string); ok && name != "" {
		student.Name = name
	}

	if username, ok := req["username"].(string); ok && username != "" {
		student.Username = username
	}

	if password, ok :=  req["password"].(string); ok && password != "" {
		if len(password) <= 5 || len(password) >= 12 {
			return &fiber.Error{
				Code: fiber.StatusBadRequest,
				Message: "Password must be between 5 and 12 characters",
			}
		}

		hashletter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
		hashnumber := regexp.MustCompile(`\d`).MatchString(password)
		hashspecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)

		if !hashletter || !hashnumber || !hashspecial {
			return &fiber.Error{
				Code: fiber.StatusBadRequest,
				Message: "Password must contain at least one letter, one number, and one special character",
			}
		}

		encode, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return &fiber.Error{
				Code: fiber.StatusInternalServerError,
				Message: err.Error(),
			}
		}
		student.Password = string(encode)
	}


	if email, ok := req["email"].(string); ok && email != "" {
		if !isValidEmail(email) {
			return &fiber.Error{
				Code: fiber.StatusBadRequest,
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

	req["updated_at"] = time.Now()


	err = configs.Database().Transaction(func(tx *gorm.DB) error {
		result := tx.Save(&student)
		if result.Error != nil {
			return &fiber.Error{
				Code: fiber.StatusInternalServerError,
				Message: result.Error.Error(),
			}
		}
		return nil
	})


	if err != nil {
		return &fiber.Error{
			Code: fiber.StatusInternalServerError,
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

	if updatedAt, ok := req["updated_at"].(time.Time); ok {
		response.UpdatedAt = &updatedAt
	}

	return c.JSON(res.ResponseCode{
		Code: fiber.StatusOK,
		Message: "Student updated successfully",
		Data: response,
	})

}


func (s *StudentController) DeleteStudent(c *fiber.Ctx) error {
	userID := c.Params("id")

	err := configs.Database().Transaction(func(tx *gorm.DB) error {
		var student models.Students

		result := tx.Where("id = ?", userID).First(&student)
		if result.Error != nil {
			return &fiber.Error{
				Code: fiber.StatusNotFound,
				Message: result.Error.Error(),
			}
		}

		if err := tx.Delete(&student).Error; err != nil {
			return &fiber.Error{
				Code: fiber.StatusInternalServerError,
				Message: err.Error(),
			}
		}

		return nil
		
	})

	if err != nil {
		return &fiber.Error{
			Code: fiber.StatusBadRequest,
			Message: err.Error(),
		}
	}

	return c.JSON(res.ResponseCode{
		Code: fiber.StatusOK,
		Message: "Student deleted successfully",
	})

}
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


type TeacherController struct {}

func NewTeacherController() *TeacherController {
	return &TeacherController{}
}

func (t *TeacherController) RegisterTeacher(c *fiber.Ctx) error {
	var req models.Teachers
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
	teacher := models.Teachers {
		Name: req.Name,
		Username: req.Username,
		Email: req.Email,
		Password: string(encode),
		Classes: req.Classes,
		CreatedAt: time.Now(),
	}

	err = configs.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&teacher).Error; err != nil {
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

	response := res.AdminResponse{
		ID: teacher.ID,
		Name: teacher.Name,
		Username: teacher.Username,
		Email: teacher.Email,
		CreatedAt: teacher.CreatedAt,
	}

	if teacher.UpdatedAt.IsZero() {
		response.UpdatedAt = nil
	}

	return c.JSON(res.ResponseCode{
		Code: fiber.StatusOK,
		Message: "Teacher registered successfully",
		Data: response,
	})
	
}


func (t *TeacherController) LoginTeacher(c *fiber.Ctx) error {
	var req models.TeachersRequest
	if err := c.BodyParser(&req); err != nil {
		return &fiber.Error{
			Code: fiber.StatusBadRequest,
			Message: err.Error(),
		}
	}

	if !isValidEmail(req.Email) {
		return &fiber.Error{
			Code: fiber.StatusBadRequest,
			Message: "Invalid email format",
		}
	}

	var teacher models.Teachers
	err := configs.Database().Transaction(func(tx *gorm.DB) error {
		result := tx.Where("email = ?", req.Email).First(&teacher)
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
			Code: fiber.StatusNotFound,
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
		"role" : "teacher",
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
		ID: teacher.ID,
		Name: teacher.Name,
		Username: teacher.Username,
		Email: teacher.Email,
		CreatedAt: teacher.CreatedAt,
		Token: tokenString,
	}

	if teacher.UpdatedAt.IsZero() {
		response.UpdatedAt = nil
	}

	return c.JSON(res.ResponseCode{
		Code: fiber.StatusOK,
		Message: "Teacher logged in successfully",
		Data: response,
	})
}


func (t *TeacherController) GetSessionProfileTeacher(c *fiber.Ctx) error {
	var response models.Teachers
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
				Code: fiber.StatusInternalServerError,
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

	data := res.TeacherResponse{
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


func (t *TeacherController) UpdateTeacher(c *fiber.Ctx) error {
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
		"createdClasses": true,
	}

	for key := range req {
		if !allowfields[key] {
			delete(req, key)
		}
	}

	var teacher models.Teachers
	err := configs.Database().Transaction(func(tx *gorm.DB) error {
		result := tx.Where("id = ?", id).First(&teacher)
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
		teacher.Name = name
	}

	if username, ok := req["username"].(string); ok && username != "" {
		teacher.Username = username
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
		 
		teacher.Password =  string(encode)
	}


	if email, ok := req["email"].(string); ok && email != "" {
		if !isValidEmail(email) {
			return &fiber.Error{
				Code: fiber.StatusBadRequest,
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

	req["updated_at"] = time.Now()

	err = configs.Database().Transaction(func(tx *gorm.DB) error {
		result := tx.Save(&teacher)
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
		Code: fiber.StatusOK,
		Message: "Teacher updated successfully",
		Data: response,
	})

}


func (t *TeacherController) DeleteTeacher(c *fiber.Ctx) error {
	userID := c.Params("id")
	err := configs.Database().Transaction(func(tx *gorm.DB) error {
		var teacher models.Teachers

		result := tx.Where("id = ?", userID).First(&teacher)
		if result.Error != nil {
			return &fiber.Error{
				Code: fiber.StatusNotFound,
				Message: result.Error.Error(),
			}
		}

		if err := tx.Delete(&teacher).Error; err != nil {
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
		Message: "Teacher deleted successfully",
	})

}
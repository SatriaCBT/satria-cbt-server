package controllers

import (
	"encoding/base64"
	"regexp"
	"satriacbtserver/configs"
	"satriacbtserver/models"
	"satriacbtserver/res"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
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

	hashletter := regexp.MustCompile(`[a-zA-Z]`).MatchString(req.Password)
	hashnumber := regexp.MustCompile(`\d`).MatchString(req.Password)
	hashspecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(req.Password)

	if !hashletter || !hashnumber || !hashspecial {
		return &fiber.Error{
			Code: fiber.StatusBadRequest,
			Message: "Password must contain at least one letter, one number, and one special character",
		}
	}

	encode := base64.StdEncoding.EncodeToString([]byte(req.Password))
	student := models.Students{
		Name: req.Name,
		Username: req.Username,
		Email: req.Email,
		Password: string(encode),
		Classes: req.Classes,
		CreatedAt: time.Now(),
	}

	err := configs.Database().Transaction(func(tx *gorm.DB) error {
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

	response := res.StudentResponse{
		ID: student.ID,
		Name: student.Name,
		Username: student.Username,
		Email: student.Email,
		Classes: student.Classes,
		CreatedAt: student.CreatedAt,
	}

	if !student.UpdatedAt.IsZero() {
		response.UpdatedAt = &student.UpdatedAt
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

	encode := base64.StdEncoding.EncodeToString([]byte(req.Password))
	password := string(encode)

	var student models.Students
	err := configs.Database().Transaction(func(tx *gorm.DB) error {
		result := tx.Where("email = ? AND password = ?", req.Email, password).First(&student)
		return &fiber.Error{
			Code: fiber.StatusInternalServerError,
			Message: result.Error.Error(),
		}
	})

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusUnauthorized).JSON(res.ResponseCode{
				Code:  fiber.StatusUnauthorized,
				Message: "Invalid username or password",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(res.ResponseCode{
			Code:  fiber.StatusInternalServerError,
			Message: "Failed to login user",
		})
	}

	token :=  jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"id": student.ID,
		"email": student.Email,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err :=  token.SignedString([]byte("secret"))

	if err != nil {
		return &fiber.Error{
			Code: fiber.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	response := res.StudentLoginResponse{
		ID: student.ID,
		Name: student.Name,
		Username: student.Username,
		Email: student.Email,
		CreatedAt: student.CreatedAt,
		Token: tokenString,
	}

	if !student.UpdatedAt.IsZero() {
		response.UpdatedAt = &student.UpdatedAt
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
		return &fiber.Error{
			Code: fiber.ErrUnauthorized.Code,
			Message: result.Error.Error(),
		}
	})

	if err != nil {
		return &fiber.Error{
			Code: fiber.StatusNotFound,
			Message: err.Error(),
		}
	}

	data := res.StudentResponse{
		ID: response.ID,
		Name: response.Name,
		Username: response.Username,
		Email: response.Email,
		Classes: response.Classes,
		CreatedAt: response.CreatedAt,
	}

	if !response.UpdatedAt.IsZero() {
		data.UpdatedAt = &response.UpdatedAt
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
		return &fiber.Error{
			Code: fiber.ErrUnauthorized.Code,
			Message: result.Error.Error(),
		}
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
		student.Password =  password
	}


	if email, ok := req["email"].(string); ok && email != "" {
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


	err = configs.Database().Transaction(func(tx *gorm.DB) error {
		result := tx.Save(&student)
		return &fiber.Error{
			Code: fiber.StatusInternalServerError,
			Message: result.Error.Error(),
		}
	})

	if err != nil {
		return &fiber.Error{
			Code: fiber.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	response := res.StudentResponse{
		ID: student.ID,
		Name: student.Name,
		Username: student.Username,
		Email: student.Email,
		Classes: student.Classes,
		CreatedAt: student.CreatedAt,
	}

	if !student.UpdatedAt.IsZero() {
		response.UpdatedAt = &student.UpdatedAt
	}

	return c.JSON(res.ResponseCode{
		Code: fiber.StatusOK,
		Message: "Student updated successfully",
		Data: response,
	})

}


func (s *StudentController) DeleteStudent(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(res.ResponseCode{
			Code:  fiber.StatusUnauthorized,
			Message: "User ID not found in context",
		})
	}

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
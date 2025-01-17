package controllers

import (
	"encoding/base64"
	_ "fmt"
	"os"
	"regexp"
	"satriacbtserver/configs"
	"satriacbtserver/models"
	"satriacbtserver/res"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type AdminController struct {}

func NewAdminController() *AdminController {
	return &AdminController{}
}


func isValidEmail(email string) bool {
    re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    return re.MatchString(email)
}



func (a *AdminController) RegisterAdmin(c *fiber.Ctx) error {
	var req models.Admins
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

	hashletter := regexp.MustCompile(`[A-Za-z]`).MatchString(req.Password)
	hashdigits := regexp.MustCompile(`\d`).MatchString(req.Password)
	hashSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(req.Password)

	if !hashletter || !hashdigits || !hashSpecial {
		return &fiber.Error{
			Code: fiber.StatusBadRequest,
			Message: "Password must contain at least one letter, one number, and one special character",
		}
	}

	encode := base64.StdEncoding.EncodeToString([]byte(req.Password))
	admin := models.Admins{
		Name:     req.Name,
		Username: req.Username,
		Email:    req.Email,
		Password: encode,
		CreatedAt: time.Now(),
	}

	err := configs.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&admin).Error; err != nil {
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
		ID:        admin.ID,
		Name:      admin.Name,
		Username:  admin.Username,
		Email:     admin.Email,
		CreatedAt: admin.CreatedAt,
	}

	if admin.UpdatedAt.IsZero() {
		response.UpdatedAt = nil
	}

	return c.JSON(res.ResponseCode{
		Code: fiber.StatusOK,
		Message: "Admin created successfully",
		Data: response,
	})
}


func (a *AdminController) LoginAdmin(c *fiber.Ctx) error {
	var req models.AdminsRequest
	if err := c.BodyParser(&req); err != nil {
		return &fiber.Error{
			Code:    fiber.StatusBadRequest,
			Message: "Invalid request body",
		}
	}

	if req.Email == "" || req.Password == "" {
		return &fiber.Error{
			Code:    fiber.StatusBadRequest,
			Message: "Email and password are required",
		}
	}

	if !isValidEmail(req.Email) {
		return &fiber.Error{
			Code: fiber.StatusBadRequest,
			Message: "Invalid email format",
		}
	}

	encodedPassword := base64.StdEncoding.EncodeToString([]byte(req.Password))
	password := string(encodedPassword)

	var admin models.Admins
	err := configs.Database().Transaction(func(tx *gorm.DB) error {
		result := tx.Where("email = ? AND password = ?", req.Email, password).First(&admin)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				return &fiber.Error{
					Code:    fiber.StatusUnauthorized,
					Message: "Invalid email or password",
				}
			}
			return result.Error
		}
		return nil
	})

	if err != nil {
		if fiberErr, ok := err.(*fiber.Error); ok {
			return fiberErr
		}
		return c.Status(fiber.StatusInternalServerError).JSON(res.ResponseCode{
			Code:    fiber.StatusInternalServerError,
			Message: "Failed to login user",
		})
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    admin.ID,
		"email": admin.Email,
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

	response := res.AdminLoginResponse{
		ID:        admin.ID,
		Name:      admin.Name,
		Username:  admin.Username,
		Email:     admin.Email,
		CreatedAt: admin.CreatedAt,
		Token:     tokenString,
	}

	if !admin.UpdatedAt.IsZero() {
		response.UpdatedAt = &admin.UpdatedAt
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Admin logged in successfully",
		Data:    response,
	})
}



func (a *AdminController) GetSessionProfileAdmin(c *fiber.Ctx) error {
	var response models.Admins
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

	data := res.AdminResponse{
		ID:        response.ID,
		Name:      response.Name,
		Username:  response.Username,
		Email:     response.Email,
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


func (a *AdminController) UpdateAdmin(c *fiber.Ctx) error {
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
	}

	for key := range req {
		if !allowfields[key] {
			delete(req, key)
		}
	}

	var admin models.Admins
	err := configs.Database().Transaction(func(tx *gorm.DB) error {
		result := tx.Where("id = ?", id).First(&admin)
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
		admin.Name = name
	}

	if username, ok := req["username"].(string); ok && username != "" {
		admin.Username = username
	}

	if password, ok :=  req["password"].(string); ok && password != "" {
		admin.Password =  base64.StdEncoding.EncodeToString([]byte(password))
	}


	if email, ok := req["email"].(string); ok && email != "" {
		if !isValidEmail(email) {
			return &fiber.Error{
				Code: fiber.StatusBadRequest,
				Message: "Invalid email format",
			}
		}
		admin.Email = email
	}

	req["updated_at"] = time.Now()
	


	err = configs.Database().Transaction(func(tx *gorm.DB) error {
		result := tx.Save(&admin)
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

	response := res.AdminResponse{
		ID:        admin.ID,
		Name:      admin.Name,
		Username:  admin.Username,
		Email:     admin.Email,
		CreatedAt: admin.CreatedAt,
	}

	if updatedAt, ok := req["updated_at"].(time.Time); ok {
		response.UpdatedAt = &updatedAt
	}

	return c.JSON(res.ResponseCode{
		Code: fiber.StatusOK,
		Message: "Admin updated successfully",
		Data: response,
	})

}


func (a *AdminController) DeleteAdmin(c *fiber.Ctx) error {
	userID := c.Params("id")

	err := configs.Database().Transaction(func(tx *gorm.DB) error {
		var totalAdmins int64
		var admin models.Admins

		if err := tx.Model(&models.Admins{}).Count(&totalAdmins).Error; err != nil {
			return &fiber.Error{
				Code: fiber.StatusInternalServerError,
				Message: err.Error(),
			}
		}

		if totalAdmins <= 1 {
			return &fiber.Error{
				Code: fiber.StatusBadRequest,
				Message: "Cannot delete the last admin",
			}
		}

		result := tx.Where("id = ?", userID).First(&admin)
		if result.Error != nil {
			return &fiber.Error{
				Code: fiber.StatusNotFound,
				Message: result.Error.Error(),
			}
		}

		if err := tx.Delete(&admin).Error; err != nil {
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
		Message: "Admin deleted successfully",
	})

}
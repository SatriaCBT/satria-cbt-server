package controllers

import (
	"encoding/base64"
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

func NewAdminController(adminController *AdminController) *AdminController {
	return &AdminController{}
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

	return c.JSON(res.ResponseCode{
		Code: fiber.StatusOK,
		Message: "Admin created successfully",
		Data: response,
	})
}


func (a *AdminController) LoginAdmin(c *fiber.Ctx) error {
	var req models.AdminsRequest
	var tokenSecret = os.Getenv("ADMIN_TOKEN")
	if err := c.BodyParser(&req); err != nil {
		return &fiber.Error{
			Code: fiber.ErrBadRequest.Code,
			Message: err.Error(),
		}
	}

	encode := base64.StdEncoding.EncodeToString([]byte(req.Password))
	password := string(encode)

	var admin models.Admins
	err := configs.Database().Transaction(func(tx *gorm.DB) error {
		result := tx.Where("email = ? AND password = ?", admin.Email, password).First(&admin)
			return &fiber.Error{
				Code: fiber.ErrUnauthorized.Code,
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

	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"admin_id": admin.ID,
		"email": admin.Email,
		"exp": jwt.TimeFunc().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(res.ResponseCode{
			Code:  fiber.StatusInternalServerError,
			Message: "Failed to login user",
		})
	}

	response := res.AdminLoginResponse{
		ID:        admin.ID,
		Name:      admin.Name,
		Username:  admin.Username,
		Email:     admin.Email,
		CreatedAt: admin.CreatedAt,
		UpdatedAt: admin.UpdatedAt,
		Token:     tokenString,
	}

	return c.JSON(res.ResponseCode{
		Code: fiber.StatusOK,
		Message: "Admin logged in successfully",
		Data: response,
	})
}


func (a *AdminController) GetSessionProfileAdmin(c *fiber.Ctx) error {
	var response models.Admins
	adminID, ok := c.Locals("adminID").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(res.ResponseCode{
			Code:  fiber.StatusUnauthorized,
			Message: "User ID not found in context",
		})
	}

	err := configs.Database().Transaction(func(tx *gorm.DB) error {
		result := tx.Where("id = ?", adminID).First(&response)
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

	data := res.AdminResponse{
		ID:        response.ID,
		Name:      response.Name,
		Username:  response.Username,
		Email:     response.Email,
		CreatedAt: response.CreatedAt,
		UpdatedAt: response.UpdatedAt,
	}

	return c.JSON(res.ResponseCode{
		Code: fiber.StatusOK,
		Message: "Profile fetched successfully",
		Data: data,
	})
	
}
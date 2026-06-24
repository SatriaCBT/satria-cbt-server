package controllers

import (
	"os"
	"regexp"
	"github.com/Satria-CBT/satria-cbt-server/models"
	"github.com/Satria-CBT/satria-cbt-server/res"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AdminController struct {
	db *gorm.DB
}

func NewAdminController(db *gorm.DB) *AdminController {
	return &AdminController{db: db}
}

func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

func validatePassword(password string) error {
	if len(password) <= 5 || len(password) >= 12 {
		return &fiber.Error{
			Code:    fiber.StatusBadRequest,
			Message: "Password must be between 5 and 12 characters",
		}
	}

	hashletter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
	hashnumber := regexp.MustCompile(`\d`).MatchString(password)
	hashspecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)

	if !hashletter || !hashnumber || !hashspecial {
		return &fiber.Error{
			Code:    fiber.StatusBadRequest,
			Message: "Password must contain at least one letter, one number, and one special character",
		}
	}
	return nil
}

func hashPassword(password string) (string, error) {
	encode, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: err.Error(),
		}
	}
	return string(encode), nil
}

func getUserID(c *fiber.Ctx) (uint, error) {
	claims, ok := c.Locals("userID").(jwt.MapClaims)
	if !ok {
		return 0, &fiber.Error{
			Code:    fiber.StatusUnauthorized,
			Message: "Unauthorized: invalid token claims",
		}
	}
	id, ok := claims["id"].(float64)
	if !ok {
		return 0, &fiber.Error{
			Code:    fiber.StatusUnauthorized,
			Message: "Unauthorized: invalid user ID",
		}
	}
	return uint(id), nil
}

func (a *AdminController) RegisterAdmin(c *fiber.Ctx) error {
	var req models.Admins
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

	admin := models.Admins{
		Name:      req.Name,
		Username:  req.Username,
		Email:     req.Email,
		Password:  encoded,
		CreatedAt: time.Now(),
	}

	if err := a.db.Create(&admin).Error; err != nil {
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
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
		Code:    fiber.StatusOK,
		Message: "Admin created successfully",
		Data:    response,
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
			Code:    fiber.StatusBadRequest,
			Message: "Invalid email format",
		}
	}

	var admin models.Admins
	if err := a.db.Where("email = ?", req.Email).First(&admin).Error; err != nil {
		return &fiber.Error{
			Code:    fiber.StatusNotFound,
			Message: err.Error(),
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)); err != nil {
		return &fiber.Error{
			Code:    fiber.StatusUnauthorized,
			Message: "Invalid password",
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    admin.ID,
		"email": admin.Email,
		"role":  "admin",
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

	if admin.UpdatedAt.IsZero() {
		response.UpdatedAt = nil
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Admin logged in successfully",
		Data:    response,
	})
}

func (a *AdminController) GetSessionProfileAdmin(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	var response models.Admins
	if err := a.db.Where("id = ?", userID).First(&response).Error; err != nil {
		return &fiber.Error{
			Code:    fiber.StatusNotFound,
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

	if response.UpdatedAt.IsZero() {
		data.UpdatedAt = nil
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Profile fetched successfully",
		Data:    data,
	})
}

func (a *AdminController) UpdateAdmin(c *fiber.Ctx) error {
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
	}

	for key := range req {
		if !allowfields[key] {
			delete(req, key)
		}
	}

	var admin models.Admins
	if err := a.db.Where("id = ?", id).First(&admin).Error; err != nil {
		return &fiber.Error{
			Code:    fiber.StatusNotFound,
			Message: err.Error(),
		}
	}

	if name, ok := req["name"].(string); ok && name != "" {
		admin.Name = name
	}

	if username, ok := req["username"].(string); ok && username != "" {
		admin.Username = username
	}

	if password, ok := req["password"].(string); ok && password != "" {
		if err := validatePassword(password); err != nil {
			return err
		}
		encoded, err := hashPassword(password)
		if err != nil {
			return err
		}
		admin.Password = encoded
	}

	if email, ok := req["email"].(string); ok && email != "" {
		if !isValidEmail(email) {
			return &fiber.Error{
				Code:    fiber.StatusBadRequest,
				Message: "Invalid email format",
			}
		}
		admin.Email = email
	}

	admin.UpdatedAt = time.Now()

	if err := a.db.Save(&admin).Error; err != nil {
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
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
	} else {
		response.UpdatedAt = &admin.UpdatedAt
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Admin updated successfully",
		Data:    response,
	})
}

func (a *AdminController) DeleteAdmin(c *fiber.Ctx) error {
	userID := c.Params("id")

	var totalAdmins int64
	if err := a.db.Model(&models.Admins{}).Count(&totalAdmins).Error; err != nil {
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	if totalAdmins <= 1 {
		return &fiber.Error{
			Code:    fiber.StatusBadRequest,
			Message: "Cannot delete the last admin",
		}
	}

	var admin models.Admins
	if err := a.db.Where("id = ?", userID).First(&admin).Error; err != nil {
		return &fiber.Error{
			Code:    fiber.StatusNotFound,
			Message: err.Error(),
		}
	}

	if err := a.db.Delete(&admin).Error; err != nil {
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	return c.JSON(res.ResponseCode{
		Code:    fiber.StatusOK,
		Message: "Admin deleted successfully",
	})
}

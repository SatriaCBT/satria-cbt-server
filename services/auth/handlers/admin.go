package handlers

import (
	"os"

	authmodels "github.com/Satria-CBT/satria-cbt-server/services/auth/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AdminHandler struct {
	db *gorm.DB
}

func NewAdminHandler(db *gorm.DB) *AdminHandler {
	return &AdminHandler{db: db}
}

func (h *AdminHandler) Login(c *fiber.Ctx) error {
	var req authmodels.AdminRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}

	var admin authmodels.Admin
	if err := h.db.Where("email = ?", req.Email).First(&admin).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Admin not found")
	}

	if bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)) != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid password")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    admin.ID,
		"email": admin.Email,
		"role":  "admin",
		"exp":   jwt.TimeFunc().Add(86400).Unix(),
	})
	secretKey := os.Getenv("ADMIN_TOKEN")
	tokenStr, _ := token.SignedString([]byte(secretKey))

	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Login successful",
		"data": fiber.Map{
			"id":    admin.ID,
			"name":  admin.Name,
			"email": admin.Email,
			"token": tokenStr,
		},
	})
}

func (h *AdminHandler) Register(c *fiber.Ctx) error {
	var req authmodels.Admin
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}
	encoded, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	req.Password = string(encoded)
	if err := h.db.Create(&req).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"code":    201,
		"message": "Admin created",
		"data":    req,
	})
}

func (h *AdminHandler) Profile(c *fiber.Ctx) error {
	claims := c.Locals("userID").(jwt.MapClaims)
	var admin authmodels.Admin
	if err := h.db.First(&admin, claims["id"]).Error; err != nil {
		return fiber.ErrNotFound
	}
	return c.JSON(fiber.Map{"code": 200, "message": "OK", "data": admin})
}

func (h *AdminHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	var req authmodels.Admin
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}
	var admin authmodels.Admin
	if err := h.db.First(&admin, id).Error; err != nil {
		return fiber.ErrNotFound
	}
	if req.Name != "" {
		admin.Name = req.Name
	}
	if req.Email != "" {
		admin.Email = req.Email
	}
	h.db.Save(&admin)
	return c.JSON(fiber.Map{"code": 200, "message": "Updated", "data": admin})
}

func (h *AdminHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	h.db.Delete(&authmodels.Admin{}, id)
	return c.JSON(fiber.Map{"code": 200, "message": "Deleted"})
}

package handlers

import (
	"os"

	authmodels "github.com/Satria-CBT/satria-cbt-server/services/auth/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type TeacherHandler struct {
	db *gorm.DB
}

func NewTeacherHandler(db *gorm.DB) *TeacherHandler {
	return &TeacherHandler{db: db}
}

func (h *TeacherHandler) Login(c *fiber.Ctx) error {
	var req authmodels.TeacherRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}
	var teacher authmodels.Teacher
	if err := h.db.Where("email = ?", req.Email).First(&teacher).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Teacher not found")
	}
	if bcrypt.CompareHashAndPassword([]byte(teacher.Password), []byte(req.Password)) != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid password")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    teacher.ID,
		"email": teacher.Email,
		"role":  "teacher",
		"exp":   jwt.TimeFunc().Add(86400).Unix(),
	})
	secretKey := os.Getenv("ADMIN_TOKEN")
	tokenStr, _ := token.SignedString([]byte(secretKey))
	return c.JSON(fiber.Map{
		"code": 200, "message": "Login successful",
		"data": fiber.Map{"id": teacher.ID, "name": teacher.Name, "email": teacher.Email, "token": tokenStr},
	})
}

func (h *TeacherHandler) Register(c *fiber.Ctx) error {
	var req authmodels.Teacher
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}
	encoded, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	req.Password = string(encoded)
	if err := h.db.Create(&req).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"code": 201, "message": "Teacher created"})
}

func (h *TeacherHandler) Profile(c *fiber.Ctx) error {
	claims := c.Locals("userID").(jwt.MapClaims)
	var teacher authmodels.Teacher
	if err := h.db.First(&teacher, claims["id"]).Error; err != nil {
		return fiber.ErrNotFound
	}
	return c.JSON(fiber.Map{"code": 200, "data": teacher})
}

func (h *TeacherHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	var req authmodels.Teacher
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}
	var teacher authmodels.Teacher
	if err := h.db.First(&teacher, id).Error; err != nil {
		return fiber.ErrNotFound
	}
	if req.Name != "" {
		teacher.Name = req.Name
	}
	if req.Email != "" {
		teacher.Email = req.Email
	}
	h.db.Save(&teacher)
	return c.JSON(fiber.Map{"code": 200, "data": teacher})
}

func (h *TeacherHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	h.db.Delete(&authmodels.Teacher{}, id)
	return c.JSON(fiber.Map{"code": 200, "message": "Deleted"})
}

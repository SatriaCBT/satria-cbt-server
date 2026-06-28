package handlers

import (
	"os"

	authmodels "github.com/Satria-CBT/satria-cbt-server/services/auth/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type StudentHandler struct {
	db *gorm.DB
}

func NewStudentHandler(db *gorm.DB) *StudentHandler {
	return &StudentHandler{db: db}
}

func (h *StudentHandler) Login(c *fiber.Ctx) error {
	var req authmodels.StudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}
	var student authmodels.Student
	if err := h.db.Where("email = ?", req.Email).First(&student).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Student not found")
	}
	if bcrypt.CompareHashAndPassword([]byte(student.Password), []byte(req.Password)) != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid password")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    student.ID,
		"email": student.Email,
		"role":  "student",
		"exp":   jwt.TimeFunc().Add(86400).Unix(),
	})
	secretKey := os.Getenv("ADMIN_TOKEN")
	tokenStr, _ := token.SignedString([]byte(secretKey))
	return c.JSON(fiber.Map{
		"code": 200, "message": "Login successful",
		"data": fiber.Map{"id": student.ID, "name": student.Name, "email": student.Email, "token": tokenStr},
	})
}

func (h *StudentHandler) Register(c *fiber.Ctx) error {
	var req authmodels.Student
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}
	encoded, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	req.Password = string(encoded)
	if err := h.db.Create(&req).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"code": 201, "message": "Student created"})
}

func (h *StudentHandler) Profile(c *fiber.Ctx) error {
	claims := c.Locals("userID").(jwt.MapClaims)
	var student authmodels.Student
	if err := h.db.First(&student, claims["id"]).Error; err != nil {
		return fiber.ErrNotFound
	}
	return c.JSON(fiber.Map{"code": 200, "data": student})
}

func (h *StudentHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	var req authmodels.Student
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}
	var student authmodels.Student
	if err := h.db.First(&student, id).Error; err != nil {
		return fiber.ErrNotFound
	}
	if req.Name != "" {
		student.Name = req.Name
	}
	if req.Email != "" {
		student.Email = req.Email
	}
	h.db.Save(&student)
	return c.JSON(fiber.Map{"code": 200, "data": student})
}

func (h *StudentHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	h.db.Delete(&authmodels.Student{}, id)
	return c.JSON(fiber.Map{"code": 200, "message": "Deleted"})
}

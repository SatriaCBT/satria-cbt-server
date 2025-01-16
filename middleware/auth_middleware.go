package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gofiber/fiber/v2"
)

type Claims struct {
	ID string `json:"id"`
	jwt.RegisteredClaims
}

func AuthenticateToken(secretKey string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Access token not found",
			})
		}

		tokenString := strings.Split(authHeader, " ")[1]
		if tokenString == "" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Access token not found",
			})
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})
		if err != nil || !token.Valid {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "Invalid token",
			})
		}

		c.Locals("userID", claims)
		return c.Next()
	}
}

func GetAuthenticatedUser(c *fiber.Ctx) error {
	user := c.Locals("userID").(*Claims)
	return c.JSON(fiber.Map{
		"success": true,
		"userID":  user.ID,
		"expires": user.ExpiresAt,
	})
}

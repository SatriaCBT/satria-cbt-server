package auth

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

func AuthenticateToken(role []string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Access token not found",
			})
		}

		bearerParts := strings.Split(authHeader, " ")
		if len(bearerParts) != 2 || bearerParts[0] != "Bearer" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Invalid authorization format",
			})
		}

		token, err := jwt.Parse(bearerParts[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			secret := os.Getenv("ADMIN_TOKEN")
			if secret == "" {
				return nil, fmt.Errorf("ADMIN_TOKEN not set")
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "Invalid or expired token",
			})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "Invalid token claims",
			})
		}

		if len(role) > 0 {
			userRole, ok := claims["role"].(string)
			if !ok {
				return c.Status(http.StatusForbidden).JSON(fiber.Map{
					"success": false,
					"message": "Invalid role claim",
				})
			}

			allowed := false
			for _, r := range role {
				if userRole == r {
					allowed = true
					break
				}
			}
			if !allowed {
				return c.Status(http.StatusForbidden).JSON(fiber.Map{
					"success": false,
					"message": "Insufficient permissions",
				})
			}
		}

		c.Locals("userID", claims)
		return c.Next()
	}
}

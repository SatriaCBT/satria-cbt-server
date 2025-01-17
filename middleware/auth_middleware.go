package middleware

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

		tokenString := bearerParts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			
			secretKey := os.Getenv("ADMIN_TOKEN")
			if secretKey == "" {
				fmt.Println("Warning: ADMIN_TOKEN is empty")
			}
			return []byte(secretKey), nil
		})

		if err != nil {
			fmt.Printf("Token parsing error: %v\n", err)
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": fmt.Sprintf("Token validation failed: %v", err),
			})
		}

		if !token.Valid {
			fmt.Printf("Token invalid: %+v\n", token)
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "Token is invalid",
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

			isAllowed := false
			for _, r := range role {
				if userRole == r {
					isAllowed = true
					break
				}
			}

			if !isAllowed {
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

func GetAuthenticatedUser(c *fiber.Ctx) error {
	claims := c.Locals("userID").(jwt.MapClaims)
	return c.JSON(fiber.Map{
		"success": true,
		"id":      claims["id"],
		"email":   claims["email"],
		"role":    claims["role"],
		"exp":     claims["exp"],
	})
}
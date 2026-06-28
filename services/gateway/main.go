package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
)

var jwtSecret []byte

func main() {
	godotenv.Load()
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("default-secret-change-in-production")
	}

	app := fiber.New()
	app.Use(cors.New())

	authURL := getEnv("AUTH_SERVICE_URL", "http://localhost:3001")
	examURL := getEnv("EXAM_SERVICE_URL", "http://localhost:3002")

	// Public auth routes (no token required)
	app.Post("/api/admin/login", proxy.Forward(authURL+"/api/admin/login"))
	app.Post("/api/teacher/login", proxy.Forward(authURL+"/api/teacher/login"))
	app.Post("/api/student/login", proxy.Forward(authURL+"/api/student/login"))
	app.Post("/api/admin/register", proxy.Forward(authURL+"/api/admin/register"))

	// Protected routes
	api := app.Group("/api", authMiddleware)

	// Auth routes (plural in gateway, singular in auth service)
	api.All("/admin*", proxy.Forward(authURL+"/api/admin"))
	api.All("/teacher*", proxy.Forward(authURL+"/api/teacher"))
	api.All("/student*", proxy.Forward(authURL+"/api/student"))

	// Exam routes
	api.All("/subjects*", proxy.Forward(examURL+"/api/subjects"))
	api.All("/questions*", proxy.Forward(examURL+"/api/questions"))
	api.All("/exams*", proxy.Forward(examURL+"/api/exams"))
	api.All("/attempts*", proxy.Forward(examURL+"/api/attempts"))
	api.All("/classes*", proxy.Forward(examURL+"/api/classes"))
	api.All("/dashboard*", proxy.Forward(examURL+"/api/dashboard"))

	// WebSocket proxy (no auth middleware — WS handshake can't carry Bearer header easily)
	app.All("/ws*", proxy.Forward(examURL))

	log.Println("Gateway running on :3000")
	log.Fatal(app.Listen(":" + getPort("3000")))
}

func getPort(def string) string {
	if p := os.Getenv("GATEWAY_PORT"); p != "" {
		return p
	}
	return def
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

type Claims struct {
	ID   uint   `json:"id"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func authMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "Missing token")
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenStr == authHeader {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid token format")
	}

	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
	}

	c.Request().Header.Set("X-User-ID", fmt.Sprintf("%d", claims.ID))
	c.Request().Header.Set("X-User-Role", claims.Role)

	return c.Next()
}

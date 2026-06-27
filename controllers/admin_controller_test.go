package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"io"
	"log"
	"net/http/httptest"
	"os"
	"github.com/Satria-CBT/satria-cbt-server/configs"
	"github.com/Satria-CBT/satria-cbt-server/models"
	"github.com/Satria-CBT/satria-cbt-server/res"
	"testing"
)

func TestMain(m *testing.M) {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	code := m.Run()
	os.Exit(code)
}

func setupTest() *fiber.App {
	app := fiber.New()
	db := configs.Database()
	adminController := NewAdminController(db)

	app.Post("/register", adminController.RegisterAdmin)
	app.Post("/login", adminController.LoginAdmin)
	app.Get("/profile", adminController.GetSessionProfileAdmin)
	app.Put("/update/:id", adminController.UpdateAdmin)
	app.Delete("/delete/:id", adminController.DeleteAdmin)

	return app
}

func TestRegisterAdmin(t *testing.T) {
	configs.Database()
	app := setupTest()

	tests := []struct {
		name           string
		payload        models.Admins
		expectedStatus int
	}{
		{
			name: "Valid Registration",
			payload: models.Admins{
				Name:     "hfohf",
				Username: "oefh",
				Email:    "oeqfh@example.com",
				Password: "Pass123!",
			},
			expectedStatus: fiber.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payloadBytes, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(payloadBytes))
			req.Header.Set("Content-Type", "application/json")

			resp, _ := app.Test(req)
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					log.Println("Error closing response body:", err)
				}
			}(resp.Body)
			responseBodyBytes, _ := io.ReadAll(resp.Body)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			var response res.AdminResponse
			err := json.Unmarshal(responseBodyBytes, &response)
			if err != nil {
				t.Fatalf("Error unmarshalling response body: %v", err)
			}
		})
	}
}

func TestAdminController_LoginAdmin(t *testing.T) {
	configs.Database()
	app := setupTest()
	tests := []struct {
		name           string
		payload        models.AdminsRequest
		expectedStatus int
	}{
		// TODO: Add test cases.
		{
			name: "Valid Login",
			payload: models.AdminsRequest{
				Email:    "oeqfh@example.com",
				Password: "Pass123!",
			},
			expectedStatus: fiber.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payloadBytes, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(payloadBytes))
			req.Header.Set("Content-Type", "application/json")
			test, _ := app.Test(req)
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					log.Println("Error closing response body:", err)
				}
			}(test.Body)

			responseBodyBytes, _ := io.ReadAll(test.Body)
			assert.Equal(t, tt.expectedStatus, test.StatusCode)
			var response res.AdminLoginResponse
			err := json.Unmarshal(responseBodyBytes, &response)
			if err != nil {
				t.Fatalf("Error unmarshalling response body: %v", err)
			}
		})
	}
}

func TestAdminController_GetSessionProfileAdmin(t *testing.T) {
	err := os.Setenv("ADMIN_TOKEN", "admin")
	if err != nil {
		t.Fatalf("Failed to set environment variable: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("ADMIN_TOKEN"); err != nil {
			t.Logf("Failed to unset environment variable: %v", err)
		}
	}()

	db := configs.Database()
	app := setupTest()

	tests := []struct {
		name           string
		payload        string
		expectedStatus int
	}{
		{
			name:           "Valid Get Session Profile",
			payload:        "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImxhZmtuYWxma25AZ21haWwuY29tIiwiZXhwIjoxNzM4MTkxMTMxLCJpZCI6MTEsInJvbGUiOiJhZG1pbiJ9.RvHJkyiJpiA4y-cpjL4fEGR8XaPf5Wka-Voxdev-wtA",
			expectedStatus: fiber.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{}
			c := app.AcquireCtx(ctx)
			defer app.ReleaseCtx(c)

			if tt.payload != "" {
				token, err := jwt.Parse(tt.payload, func(token *jwt.Token) (interface{}, error) {
					return []byte("admin"), nil
				})
				if err != nil {
					t.Fatalf("Failed to parse JWT: %v", err)
				}

				claims, ok := token.Claims.(jwt.MapClaims)
				if !ok || !token.Valid {
					t.Fatal("Invalid token claims")
				}

				c.Locals("userID", claims)
			}

			a := NewAdminController(db)
			err := a.GetSessionProfileAdmin(c)
			if err != nil {
				t.Fatalf("Error getting session profile admin: %v", err)
			}

			if c.Response().StatusCode() != tt.expectedStatus {
				t.Errorf("Expected status %v but got %v", tt.expectedStatus, c.Response().StatusCode())
			}

			var response struct {
				Code    int               `json:"code"`
				Message string            `json:"message"`
				Data    res.AdminResponse `json:"data"`
			}

			err = json.Unmarshal(c.Response().Body(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			assert.Equal(t, fiber.StatusOK, response.Code)
			assert.Equal(t, "Profile fetched successfully", response.Message)
			assert.NotEmpty(t, response.Data.ID)
			assert.NotEmpty(t, response.Data.Email)
			assert.NotEmpty(t, response.Data.Name)
			assert.NotEmpty(t, response.Data.Username)
		})
	}
}

func TestAdminController_UpdateAdmin(t *testing.T) {
	configs.Database()
	app := setupTest()
	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
	}{
		// TODO: Add test cases.
		{
			name: "Valid Update",
			payload: map[string]interface{}{
				"name": "budiman",
			},
			expectedStatus: fiber.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowfields := map[string]bool{
				"name":     true,
				"username": true,
				"password": true,
				"email":    true,
			}

			for key := range tt.payload {
				if !allowfields[key] {
					t.Fatalf("Invalid field name: %s", key)
				}
			}

			id := "11"

			payloadBytes, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("PUT", "/update/"+id, bytes.NewBuffer(payloadBytes))
			req.Header.Set("Content-Type", "application/json")

			resp, _ := app.Test(req)
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					log.Println("Error closing response body:", err)
				}
			}(resp.Body)
			responseBodyBytes, _ := io.ReadAll(resp.Body)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			var response res.AdminResponse
			err := json.Unmarshal(responseBodyBytes, &response)
			if err != nil {
				t.Fatalf("Error unmarshalling response body: %v", err)
			}

		})
	}
}

func TestAdminController_DeleteAdmin(t *testing.T) {
	configs.Database()
	app := setupTest()

	tests := []struct {
		name           string
		payload        uint
		expectedStatus int
	}{
		// TODO: Add test cases.
		{
			name:           "Valid Delete",
			payload:        11,
			expectedStatus: fiber.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/delete/"+fmt.Sprint(tt.payload), nil)
			resp, _ := app.Test(req)
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %v but got %v", tt.expectedStatus, resp.StatusCode)
			}
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

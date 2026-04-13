package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"taskflow/internal/db"
	"taskflow/internal/utils"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func setupTestRouter() *gin.Engine {
	// Initialize connection logic mimicking main.go
	godotenv.Load("../../.env")
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" || dsn == "postgres_url" {
		dsn = "postgres://postgres:postgres@localhost:5432/taskflow?sslmode=disable"
	}
	// Docker compose uses db hostname, but local go test needs localhost.
	dsn = strings.Replace(dsn, "@db:", "@localhost:", 1)
	_ = os.Setenv("DATABASE_URL", dsn)

	db.Connect()
	utils.InitLogger()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	auth := r.Group("/auth")
	{
		auth.POST("/register", Register)
		auth.POST("/login", Login)
	}
	return r
}

func cleanupTestUser(email string) {
	if db.DB != nil {
		db.DB.Exec(context.Background(), "DELETE FROM users WHERE email=$1", email)
	}
}

func TestAuthFlow(t *testing.T) {
	r := setupTestRouter()
	testEmail := "integration_test_user@example.com"
	testPassword := "supersecret123"

	// 1. Cleanup before starting (just in case)
	cleanupTestUser(testEmail)

	// 2. Test Registration
	t.Run("Register Success", func(t *testing.T) {
		body := map[string]string{
			"name":     "Integration Test",
			"email":    testEmail,
			"password": testPassword,
		}
		jsonValue, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonValue))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("Expected status 201, got %v: %s", w.Code, w.Body.String())
		}
	})

	// 3. Test Registration Duplicate
	t.Run("Register Duplicate", func(t *testing.T) {
		body := map[string]string{
			"name":     "Integration Test",
			"email":    testEmail,
			"password": testPassword,
		}
		jsonValue, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonValue))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("Expected status 400 for duplicate, got %v: %s", w.Code, w.Body.String())
		}
	})

	// 4. Test Login Success
	t.Run("Login Success", func(t *testing.T) {
		body := map[string]string{
			"email":    testEmail,
			"password": testPassword,
		}
		jsonValue, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonValue))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %v: %s", w.Code, w.Body.String())
		}

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		if _, exists := response["token"]; !exists {
			t.Fatalf("Expected JWT token in response")
		}
	})

	// 5. Cleanup after tests
	t.Cleanup(func() {
		cleanupTestUser(testEmail)
	})
}

package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/BragdonD/betalink-auth/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Mock data
var validUser = middleware.UserData{
	UserID:    "12345",
	FirstName: "John",
	LastName:  "Doe",
}

func TestAuthRequired(t *testing.T) {
	// Mock the authentication server
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "Bearer valid-token" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(validUser)
		} else {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		}
	}))
	defer authServer.Close()

	tests := []struct {
		name         string
		authHeader   string
		expectedCode int
		expectedBody string
	}{
		{
			name:         "Valid token",
			authHeader:   "Bearer valid-token",
			expectedCode: http.StatusOK,
			expectedBody: "",
		},
		{
			name:         "Missing token",
			authHeader:   "",
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"error":"Authorization header is required"}`,
		},
		{
			name:         "Invalid token format",
			authHeader:   "Basic invalid-token",
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"error":"Invalid Authorization header format"}`,
		},
		{
			name:         "Unauthorized token",
			authHeader:   "Bearer invalid-token",
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"error":"Auth server error: Unauthorized"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up Gin
			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.Use(middleware.AuthRequired(authServer.URL))
			r.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			// Create a test request
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			// Serve the request
			r.ServeHTTP(w, req)

			// Assertions
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}
		})
	}
}

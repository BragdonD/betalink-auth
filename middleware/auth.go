package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// UserData represents the user information retrieved from the auth server
type UserData struct {
	UserID    string
	FirstName string
	LastName  string
}

// AuthRequired is a gin middleware that checks if the user is authenticated.
// If the user is not authenticated, it will return a 401 Unauthorized status.
// Otherwise, it will store the user information in the context and call
// the next handler.
func AuthRequired(authServerURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// Validate the format of the Authorization header
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			c.Abort()
			return
		}

		req, err := http.NewRequest("GET", authServerURL, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		// Forward the Authorization header to the auth server
		req.Header.Set("Authorization", authHeader)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			message := "Unauthorized"
			if resp != nil {
				body, _ := io.ReadAll(resp.Body)
				message = fmt.Sprintf("Auth server error: %s", bytes.TrimSpace(body))
				resp.Body.Close()
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": message})
			c.Abort()
			return
		}

		var user UserData
		if err := json.NewDecoder(resp.Body).Decode(&user); err == nil {
			c.Set("user", user) // Store user info in the context
		}
		resp.Body.Close()

		c.Next()
	}
}

package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// UserData represents the user information retrieved from the auth server
type UserData struct {
	UserID    string `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// AuthResponse represents the response from the auth server
type AuthResponse struct {
	Data    UserData `json:"data"`
	Success bool     `json:"success"`
	Error   string   `json:"error"`
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
			if err != nil {
				message = fmt.Sprintf("Could not connect to auth server: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"data":    "",
					"error":   message,
				})
			} else {
				var respBody map[string]interface{}
				if decodeErr := json.NewDecoder(resp.Body).Decode(&respBody); decodeErr != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"success": false,
						"data":    "",
						"error":   "Failed to parse response from auth server",
					})
					return
				}
				c.JSON(http.StatusUnauthorized, respBody)
			}
			c.Abort()
			return
		}

		var authResp AuthResponse
		if err := json.NewDecoder(resp.Body).Decode(&authResp); err == nil {
			if !authResp.Success {
				c.JSON(http.StatusUnauthorized, gin.H{"error": authResp.Error})
				c.Abort()
				return
			}
			c.Set("user", authResp.Data) // Store user info in the context
		}
		resp.Body.Close()

		c.Next()
	}
}

package betalinkauth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a password using bcrypt algorithm
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("could not hash password: %w", err)
	}
	return string(hash), nil
}

// ComparePassword compares a password with a hash
func ComparePassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// GenerateJWT generates a JWT token containing the provided data
func GenerateJWT(data map[string]interface{}, secret string) (string, error) {
	claims := jwt.MapClaims{}

	// Add claims from data to the token
	for key, value := range data {
		claims[key] = value
	}

	// Create a new token object
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the provided secret
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// GenerateAccessToken generates an access token with user-specific data
func GenerateAccessToken(userID string, roles []string, secret string) (string, error) {
	// Define claims
	claims := map[string]interface{}{
		"user_id": userID,
		"roles":   roles,
		"exp":     time.Now().Add(time.Hour).Unix(), // Token expires in 1 hour
		"iat":     time.Now().Unix(),
		"iss":     "betalink-auth",
		"aud":     "betalink",
	}

	// Generate the JWT using the helper function
	return GenerateJWT(claims, secret)
}

// ValidateAccessToken validates an access token
func ValidateAccessToken(token, secret string) (jwt.MapClaims, error) {
	// Parse the token
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not parse token: %w", err)
	}

	// Check if the token is valid
	if !parsedToken.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	// Extract the claims
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("could not extract claims from token")
	}

	return claims, nil
}

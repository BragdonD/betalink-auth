package betalinkauth_test

import (
	"testing"
	"time"

	betalinkauth "github.com/BragdonD/betalink-auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	password := "mysecretpassword"
	hash, err := betalinkauth.HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
}

func TestComparePassword(t *testing.T) {
	password := "mysecretpassword"
	hash, err := betalinkauth.HashPassword(password)
	assert.NoError(t, err)

	err = betalinkauth.ComparePassword(password, hash)
	assert.NoError(t, err)

	err = betalinkauth.ComparePassword("wrongpassword", hash)
	assert.Error(t, err)
}

func TestGenerateJWT(t *testing.T) {
	data := map[string]interface{}{
		"user_id": "12345",
		"roles":   []string{"admin", "user"},
	}
	secret := "mysecret"

	token, err := betalinkauth.GenerateJWT(data, secret)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestGenerateAccessToken(t *testing.T) {
	userID := "12345"
	roles := []string{"admin", "user"}
	secret := "mysecret"

	token, err := betalinkauth.GenerateAccessToken(userID, roles, secret)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Parse the token to verify claims
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Equal(t, userID, claims["user_id"])
	assert.ElementsMatch(t, roles, claims["roles"])
	assert.Equal(t, "betalink-auth", claims["iss"])
	assert.Equal(t, "betalink", claims["aud"])
	assert.WithinDuration(t, time.Now().Add(time.Hour), time.Unix(int64(claims["exp"].(float64)), 0), time.Minute)
	assert.WithinDuration(t, time.Now(), time.Unix(int64(claims["iat"].(float64)), 0), time.Minute)
}

func TestValidateAccessToken(t *testing.T) {
	userID := "12345"
	roles := []string{"admin", "user"}
	secret := "mysecret"

	token, err := betalinkauth.GenerateAccessToken(userID, roles, secret)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := betalinkauth.ValidateAccessToken(token, secret)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims["user_id"])
	assert.ElementsMatch(t, roles, claims["roles"])
	assert.Equal(t, "betalink-auth", claims["iss"])
	assert.Equal(t, "betalink", claims["aud"])
	assert.WithinDuration(t, time.Now().Add(time.Hour), time.Unix(int64(claims["exp"].(float64)), 0), time.Minute)
	assert.WithinDuration(t, time.Now(), time.Unix(int64(claims["iat"].(float64)), 0), time.Minute)
}

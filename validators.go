package betalinkauth

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

const (
	// emailRegex is the regex pattern for email validation
	emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	// specialChars is the list of special characters that are allowed in a password
	specialChars = `!@#$%^&*()_+{}|:"<>?`
)

// ValidateEmail validates an email address based on the following rules:
// - Must contain an @ symbol
// - Must contain a period
// - Must have at least 2 characters after the period
func ValidateEmail(email string) (bool, error) {
	emailRegex, err := regexp.Compile(emailRegex)
	if err != nil {
		return false, fmt.Errorf("error compiling email regex: %w", err)
	}
	return emailRegex.MatchString(email), nil
}

// ValidatePassword validates a password based on the following rules:
// - Must be at least 8 characters long
// - Must contain at least one lowercase letter
// - Must contain at least one uppercase letter
// - Must contain at least one digit
// - Must contain at least one special character
func ValidatePassword(password string) (bool, error) {
	if len(password) < 8 {
		return false, fmt.Errorf("password must be at least 8 characters long")
	}

	var hasLower, hasUpper, hasDigit, hasSpecial bool

	for _, char := range password {
		switch {
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsDigit(char):
			hasDigit = true
		case strings.ContainsRune(specialChars, char):
			hasSpecial = true
		}
	}

	if !hasLower {
		return false, fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasUpper {
		return false, fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasDigit {
		return false, fmt.Errorf("password must contain at least one digit")
	}
	if !hasSpecial {
		return false, fmt.Errorf("password must contain at least one special character")
	}

	return true, nil
}

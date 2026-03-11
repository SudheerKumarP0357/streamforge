package auth

import (
	"errors"
	"regexp"
)

// simple regex for basic email format validation
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail checks if an email matches a basic format
func ValidateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// ValidateRegistration checks email and password constraints
func ValidateRegistration(email, password string) error {
	if email == "" {
		return errors.New("email is required")
	}
	
	if !ValidateEmail(email) {
		return errors.New("invalid email format")
	}

	if password == "" {
		return errors.New("password is required")
	}

	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	return nil
}

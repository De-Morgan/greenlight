package data

import (
	"errors"
	"net/mail"
	"strings"
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"
	"morgan.greenlight.nex/internal/validator"
)

var AnonymousUser = &User{}

type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"version"`
}

type password struct {
	plaintext *string
	hash      []byte
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

// The Set() method calculates the bcrypt hash of a plaintext password, and stores both
// the hash and the plaintext versions in the struct.
func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}
	p.plaintext = &plaintextPassword
	p.hash = hash
	return nil
}

// The Matches() method checks whether the provided plaintext password matches the
// hashed password stored in the struct, returning true if it matches and false
// otherwise.
func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email == "", "email", "must not be empty")
	_, err := mail.ParseAddress(email)
	v.Check(err != nil, "email", "must be a valid email address")
}
func validatePassword(password string) bool {
	// Check length between 8 and 72
	if len(password) < 8 || len(password) > 72 {
		return false
	}
	// Check for at least one lowercase letter, one uppercase letter, and one special character
	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	// Return true if all conditions are met
	return hasUpper && hasLower && hasNumber && hasSpecial
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password == "", "password", "must be provided")
	v.Check(!validatePassword(password), "password", "is invalid")
}

func ValidateUserName(v *validator.Validator, name string) {
	v.Check(strings.TrimSpace(name) == "", "name", "must be provided")
}

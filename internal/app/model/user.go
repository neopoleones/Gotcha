package model

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User represents account in gotcha service.
// Call ClearSensitive and BeforeCreate to calculate hash of user. Call Validate
// before saving entity.
type User struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email,omitempty"`
	Password  string    `json:"password,omitempty"`
	Hash      string    `json:"-"` // shadow field of Password
	CreatedAt time.Time `json:"created_at"`
}

// ClearSensitive clears sensitive fields like password and...
func (u *User) ClearSensitive() {
	u.Password = "" // omitempty will hide password field in json
}

// Validate checks important fields of User and returns error if entity is malformed.
// **Constrains**
// Username: required, size(6, 32), printable ASCII
// Email: required, size(6, 64), email format, printable ASCII
// Password: required if Hash field is empty (BeforeCreate not called), size(8, 32), printable ASCII
func (u *User) Validate() error {
	usernameField := validation.Field(&u.Username, validation.Required, validation.Length(6, 32), is.PrintableASCII)
	emailField := validation.Field(&u.Email, validation.Required, validation.Length(6, 64), is.Email, is.PrintableASCII)
	passwordField := validation.Field(&u.Password, validation.By(requiredIf(u.Hash == "")), validation.Length(8, 32), is.PrintableASCII)

	return validation.ValidateStruct(u, usernameField, emailField, passwordField)
}

func (u *User) BeforeCreate() error {
	// Using 10 rounds for bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Hash = string(hashedPassword)
	return nil
}

func (u *User) IsCorrectPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.Hash), []byte(password)) == nil
}

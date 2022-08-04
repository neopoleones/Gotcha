package model

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID   uuid.UUID  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
	Hash     string `json:"-"`                  // shadow field of Password
	CreatedAt time.Time `json:"created_at"`

}

func (u *User) ClearSensitive() {
	u.Password = "" // omitempty will hide password field in json
}

func (u *User) Validate() error {
	return nil
}

func (u *User) BeforeCreate() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Hash = string(hashedPassword)
	return nil
}

func (u *User) isCorrectPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.Hash), []byte(password)) == nil
}
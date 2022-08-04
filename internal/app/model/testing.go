package model

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestUser returns user instance filled with dummy values (without hash)
func TestUser(t *testing.T) *User {
	t.Helper()

	return &User{
		ID:        uuid.New(),
		Username:  "username",
		Email:     "username@gmail.com",
		Password:  "ExamplePassword",
		CreatedAt: time.Now(),
	}
}

package model

import (
	"testing"
)

// TestUser returns user instance filled with dummy values (without hash)
func TestUser(t *testing.T) *User {
	t.Helper()

	return &User{
		Username: "username",
		Email:    "username@gmail.com",
		Password: "ExamplePassword",
	}
}

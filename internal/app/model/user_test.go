package model_test

import (
	"fmt"
	"strings"
	"testing"

	"Gotcha/internal/app/model"
	"github.com/stretchr/testify/assert"
)

func TestUser_BeforeCreate(t *testing.T) {
	user := model.TestUser(t)

	assert.NoError(t, user.BeforeCreate(), "Failed to create hash")
	assert.NotEmpty(t, user.Hash, "Hash is still nil")
	fmt.Printf("Created hash: %s\n", user.Hash)
}

func TestUser_IsCorrectPassword(t *testing.T) {
	user := model.TestUser(t)

	assert.NoError(t, user.BeforeCreate(), "Failed to create hash")
	assert.True(t, user.IsCorrectPassword(user.Password), "hash(pwd) != pwd_hash")
	assert.False(t, user.IsCorrectPassword("incorrect"), "hash(pwd2) == pwd_hash")
}

func TestUser_ClearSensitive(t *testing.T) {
	user := model.TestUser(t)
	user.ClearSensitive()

	assert.Empty(t, user.Password, "pwd not empty after ClearSensitive call")
}

func TestUser_Validate(t *testing.T) {

	testCases := []struct {
		isValid       bool
		failMessage   string
		testName      string
		userGenerator func() *model.User
	}{
		{
			isValid:     true,
			testName:    "Valid test",
			failMessage: "Valid user is recognised as malformed!",
			userGenerator: func() *model.User {
				return model.TestUser(t)
			},
		},
		{
			isValid:     false,
			testName:    "Malformed email (empty)",
			failMessage: "Validate doesn't see empty email",
			userGenerator: func() *model.User {
				u := model.TestUser(t)
				u.Email = ""
				return u
			},
		},
		{
			isValid:     false,
			testName:    "Malformed email (short)",
			failMessage: "Validate doesn't see short(<6) email",
			userGenerator: func() *model.User {
				u := model.TestUser(t)
				u.Email = "a@a.a"
				return u
			},
		},
		{
			isValid:     false,
			testName:    "Malformed email (long)",
			failMessage: "Validate doesn't see long(>64) email",
			userGenerator: func() *model.User {
				u := model.TestUser(t)
				u.Email = strings.Repeat("A", 64) + "@gmail.com"
				return u
			},
		},
		{
			isValid:     false,
			testName:    "Malformed email (incorrect format)",
			failMessage: "Validate doesn't see malformed email (incorrect format)",
			userGenerator: func() *model.User {
				u := model.TestUser(t)
				u.Email = "hacker_input' or 1=1-- -"
				return u
			},
		},
		{
			isValid:     false,
			testName:    "Malformed password (empty)",
			failMessage: "Validate doesn't see empty password",
			userGenerator: func() *model.User {
				u := model.TestUser(t)
				u.Password = ""
				return u
			},
		},
		{
			isValid:     false,
			testName:    "Malformed password (short)",
			failMessage: "Validate doesn't see short(<8) password",
			userGenerator: func() *model.User {
				u := model.TestUser(t)
				u.Password = "pass"
				return u
			},
		},
		{
			isValid:     false,
			testName:    "Malformed password (long)",
			failMessage: "Validate doesn't see long(>64) password",
			userGenerator: func() *model.User {
				u := model.TestUser(t)
				u.Email = strings.Repeat("A", 128)
				return u
			},
		},
		{
			isValid:     true,
			testName:    "Empty password when hash is specified",
			failMessage: "Hash exists, but entity still needs a password",
			userGenerator: func() *model.User {
				u := model.TestUser(t)
				u.Hash = "kinda_hash"
				u.Password = ""
				return u
			},
		},

		{
			isValid:     false,
			testName:    "Malformed username (empty)",
			failMessage: "Validate doesn't see empty username",
			userGenerator: func() *model.User {
				u := model.TestUser(t)
				u.Username = ""
				return u
			},
		},
		{
			isValid:     false,
			testName:    "Malformed username (short)",
			failMessage: "Validate doesn't see short(<6) username",
			userGenerator: func() *model.User {
				u := model.TestUser(t)
				u.Username = "abc"
				return u
			},
		},
		{
			isValid:     false,
			testName:    "Malformed username (long)",
			failMessage: "Validate doesn't see long(>32) username",
			userGenerator: func() *model.User {
				u := model.TestUser(t)
				u.Username = strings.Repeat("A", 48)
				return u
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			user := testCase.userGenerator()
			if testCase.isValid {
				assert.NoError(t, user.Validate(), testCase.failMessage)
			} else {
				assert.Error(t, user.Validate(), testCase.failMessage)
			}
		})
	}
}

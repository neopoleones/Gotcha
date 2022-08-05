package teststore

import (
	"testing"

	"Gotcha/internal/app/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUserRepository_NewUser(t *testing.T) {
	repository := New().User()

	testUser := model.TestUser(t)
	assert.NoError(t, repository.SaveUser(testUser), "Failed to save valid user")
	assert.NotEqual(t, testUser.ID, uuid.Nil, "User must have an uuid after save")
	assert.False(t, testUser.CreatedAt.IsZero(), "User must have a creation time")

	assert.Error(t, repository.SaveUser(testUser), "Duplicate user was created successfully!")
}

func TestUserRepository_FindUserBySobriquet(t *testing.T) {
	repository := New().User()

	testUser := model.TestUser(t)
	_ = repository.SaveUser(testUser) // Suppose that TestUserRepository_NewUser was passed

	sameUser1, err := repository.FindUserBySobriquet(testUser.Username)
	assert.NoError(t, err, "Failed to get user by username")

	sameUser2, err := repository.FindUserBySobriquet(testUser.Email)
	assert.NoError(t, err, "Failed to get user by email")
	assert.Equal(t, sameUser1, sameUser2, "Repository returned different users")

	_, err = repository.FindUserBySobriquet("qwerty123")
	assert.Error(t, err, "Repository returned user of unknown nickname")
}

func TestUserRepository_FindUserByID(t *testing.T) {
	repository := New().User()

	testUser := model.TestUser(t)
	_ = repository.SaveUser(testUser) // Suppose that TestUserRepository_NewUser was passed

	foundUser, err := repository.FindUserByID(testUser.ID)
	assert.NoError(t, err, "Failed to get user by id")
	testUser.ClearSensitive()
	assert.Equal(t, testUser, foundUser)

	_, err = repository.FindUserByID(uuid.Nil)
	assert.Error(t, err, "Repository returned user for nil UUID")
}

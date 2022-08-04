package postgres_test

import (
	"testing"

	"Gotcha/internal/app/model"
	"Gotcha/internal/app/storage/postgres"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUserRepository_NewUser(t *testing.T) {
	db, sanitize := postgres.TestDB(t, databaseConnectionString)
	repository := postgres.NewStore(db).User()
	defer sanitize("Users")

	testUser := model.TestUser(t)

	assert.NoError(t, repository.SaveUser(testUser), "Failed to save valid user")
	assert.NotEqual(t, testUser.ID, uuid.Nil, "User must have an uuid after save")
	assert.False(t, testUser.CreatedAt.IsZero(), "User must have a creation time")

	assert.Error(t, repository.SaveUser(testUser), "Duplicate user was created successfully!")
}

func TestUserRepository_FindUserByEmail(t *testing.T) {
	db, sanitize := postgres.TestDB(t, databaseConnectionString)
	repository := postgres.NewStore(db).User()
	defer sanitize("Users")

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

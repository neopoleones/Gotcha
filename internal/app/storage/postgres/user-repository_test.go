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

func TestUserRepository_FindUserBySobriquet(t *testing.T) {
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

func TestUserRepository_FindUserByID(t *testing.T) {
	db, sanitize := postgres.TestDB(t, databaseConnectionString)
	repository := postgres.NewStore(db).User()
	defer sanitize("Users")

	testUser := model.TestUser(t)
	_ = repository.SaveUser(testUser)

	foundUser, err := repository.FindUserByID(testUser.ID)
	assert.NoError(t, err, "Failed to get user by id")
	testUser.ClearSensitive()
	assert.Equal(t, testUser, foundUser)

	_, err = repository.FindUserByID(uuid.Nil)
	assert.Error(t, err, "Repository returned user for nil UUID")
}

func TestUserRepository_GetAllUsers(t *testing.T) {
	db, sanitize := postgres.TestDB(t, databaseConnectionString)
	repository := postgres.NewStore(db).User()
	defer sanitize("Users")

	viewer := model.TestUser(t)
	viewer.Email += "v"
	viewer.Username += "v"
	_ = repository.SaveUser(viewer)

	testUser := model.TestUser(t)
	_ = repository.SaveUser(testUser)

	// First case
	allUsers, err := repository.GetAllUsers(viewer)
	assert.NoError(t, err, "Got error while collecting users")
	assert.Equal(t, len(allUsers), 1)
	assert.Equal(t, allUsers[0].ID, testUser.ID)

	// Second case
	secondUser := model.TestUser(t)
	secondUser.Username += "s"
	secondUser.Email += "s"
	_ = repository.SaveUser(secondUser)

	allUsers, err = repository.GetAllUsers(viewer)
	assert.NoError(t, err, "Got error while collecting users (second testcase)")
	assert.Equal(t, len(allUsers), 2)
}

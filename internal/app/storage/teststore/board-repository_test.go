package teststore_test

import (
	"testing"

	"Gotcha/internal/app/model"
	"Gotcha/internal/app/storage/postgres"
	"Gotcha/internal/app/storage/teststore"
	"github.com/stretchr/testify/assert"
)

func TestBoardRepository_NewRootBoard(t *testing.T) {
	store := teststore.New()
	userRepo := store.User()
	boardRepo := store.Board()

	testUser := model.TestUser(t)
	_ = userRepo.SaveUser(testUser)

	board, err := boardRepo.NewRootBoard(testUser, "Hello world!")
	assert.NoError(t, err, "Failed to create root board")
	assert.False(t, board.CreatedAt.IsZero(), "Board not created: time not added")
	assert.Len(t, board.U2BRelations, 1, "Board not created: Owner relation not added")
}

func TestBoardRepository_Relations(t *testing.T) {
	store := teststore.New()
	userRepo := store.User()
	boardRepo := store.Board()

	testUser := model.TestUser(t)
	_ = userRepo.SaveUser(testUser)

	testBoard, _ := boardRepo.NewRootBoard(testUser, "Example root board")
	relationID, err := boardRepo.CreateRelation(testBoard, testUser, postgres.DescriptionAllGranted, model.PrivilegeReadOnly)
	assert.NoError(t, err, "Failed to create relation")

	relationPrivilege, err := boardRepo.GetPrivilegeFromRelation(relationID)
	assert.NoError(t, err, "Failed to get relation privilege")
	assert.Equal(t, relationPrivilege, model.PrivilegeReadOnly)
}
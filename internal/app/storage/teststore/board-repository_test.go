package teststore_test

import (
	"testing"

	"Gotcha/internal/app/model"
	"Gotcha/internal/app/storage"
	"Gotcha/internal/app/storage/postgres"
	"Gotcha/internal/app/storage/teststore"
	"github.com/google/uuid"
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
	assert.False(t, board.Base.CreatedAt.IsZero(), "Board not created: time not added")
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

	bp, err := boardRepo.GetPrivilegeFromRelation(relationID)
	assert.NoError(t, err, "Failed to get relation privilege")
	assert.Equal(t, bp.Privilege, model.PrivilegeReadOnly)
	assert.Equal(t, bp.BoardID, testBoard.Base.ID)
}

func TestBoardRepository_DeleteRootBoard(t *testing.T) {
	store := teststore.New()
	userRepo := store.User()
	boardRepo := store.Board()

	testUser := model.TestUser(t)
	_ = userRepo.SaveUser(testUser)

	testBoard, _ := boardRepo.NewRootBoard(testUser, "Example root board")
	assert.NoError(t, boardRepo.DeleteRootBoard(testBoard.Base.ID, testBoard.U2BRelations, testUser), "Failed to delete board")
	boards, _ := boardRepo.GetRootBoardsOfUser(testUser)
	assert.Equal(t, len(boards), 0, "Board still exists in database")

	// Check if we can delete a board with incorrect relations
	testBoard, _ = boardRepo.NewRootBoard(testUser, "Example root board")
	anotherBoard, _ := boardRepo.NewRootBoard(testUser, "Example root board")
	assert.ErrorIs(t,
		boardRepo.DeleteRootBoard(testBoard.Base.ID, anotherBoard.U2BRelations, testUser),
		storage.ErrSecurityError, "Deleted table with fake relations")

	// Check if we can delete a board as granted user (not owner)
	anotherUser := model.TestUser(t)
	anotherUser.Username += "another"
	anotherUser.Email += "another"

	newRelation, _ := boardRepo.CreateRelation(testBoard, anotherUser, "RW access for my friend", model.PrivilegeReadWrite)

	assert.ErrorIs(t,
		boardRepo.DeleteRootBoard(testBoard.Base.ID, []uuid.UUID{newRelation}, anotherUser),
		storage.ErrSecurityError, "Server allows you to delete a board as a non-owner")
}

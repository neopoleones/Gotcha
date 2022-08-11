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
	relationID, err := boardRepo.CreateRelation(testBoard.Base.ID, testUser.ID, postgres.DescriptionAllGranted, model.PrivilegeReadOnly)
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

	newRelation, _ := boardRepo.CreateRelation(testBoard.Base.ID, anotherUser.ID, "RW access for my friend", model.PrivilegeReadWrite)

	assert.ErrorIs(t,
		boardRepo.DeleteRootBoard(testBoard.Base.ID, []uuid.UUID{newRelation}, anotherUser),
		storage.ErrSecurityError, "Server allows you to delete a board as a non-owner")
}

func TestBoardRepository_NewNestedBoard(t *testing.T) {
	store := teststore.New()
	userRepo := store.User()
	boardRepo := store.Board()

	// Create test assets
	user := model.TestUser(t)
	_ = userRepo.SaveUser(user)
	rootBoard, _ := boardRepo.NewRootBoard(user, "Root")

	sideBoard, err := boardRepo.NewNestedBoard(rootBoard.Base.ID, "Nested", user)
	assert.NoError(t, err, "Failed to create side board (got error)")
	assert.NotNil(t, sideBoard, "Failed to create side board (it's nil)")
	assert.NotEqual(t, sideBoard.RelationID, uuid.Nil, "Relation ID is nil")

	// Security check
	secondUser := model.TestUser(t)
	secondUser.Username += "a"
	secondUser.Email += "a"

	_ = userRepo.SaveUser(secondUser)
	_, err = boardRepo.NewNestedBoard(rootBoard.Base.ID, "Nested", secondUser)
	assert.Error(t, err, "Somehow created sideboard as user, that doesn't have write permission")

	// Add rw permission
	_, err = boardRepo.CreateRelation(rootBoard.Base.ID, secondUser.ID, "test", model.PrivilegeReadWrite)
	_, err = boardRepo.NewNestedBoard(rootBoard.Base.ID, "Nested", secondUser)
	assert.NoError(t, err, "User has a permission, but it's forbidden to create sideboard")
}

func TestBoardRepository_GetRootOfNestedBoard(t *testing.T) {
	store := teststore.New()
	userRepo := store.User()
	boardRepo := store.Board()

	// Create test assets
	user := model.TestUser(t)
	_ = userRepo.SaveUser(user)
	rootBoard, _ := boardRepo.NewRootBoard(user, "Root")
	sideBoard, _ := boardRepo.NewNestedBoard(rootBoard.Base.ID, "Nested", user)
	secondSide, err := boardRepo.NewNestedBoard(sideBoard.Base.ID, "Nested", user)
	assert.NoError(t, err, "Failed to create second side")

	rootBoardFound, err := boardRepo.GetRootOfNestedBoard(secondSide.Base.ID)
	assert.NoError(t, err, "Failed to get root of nested board")
	assert.Equal(t, rootBoardFound.Base.ID, rootBoard.Base.ID)
}

func TestBoardRepository_GetNestedBoards(t *testing.T) {
	store := teststore.New()
	userRepo := store.User()
	boardRepo := store.Board()

	// Create test assets
	user := model.TestUser(t)
	anotherUser := model.TestUser(t)
	anotherUser.Username += "a"
	anotherUser.Email += "a"
	_ = userRepo.SaveUser(user)
	_ = userRepo.SaveUser(anotherUser)

	// Add two nested boards as owner
	rootBoard, _ := boardRepo.NewRootBoard(user, "Root")
	nestedBoardOne, _ := boardRepo.NewNestedBoard(rootBoard.Base.ID, "Nested #one", user)
	nestedBoardTwo, _ := boardRepo.NewNestedBoard(rootBoard.Base.ID, "Nested #two", user)

	// And one as guest user (rw privilege)
	boardRepo.CreateRelation(rootBoard.Base.ID, anotherUser.ID, "test", model.PrivilegeReadWrite)
	_, err := boardRepo.NewNestedBoard(rootBoard.Base.ID, "Nested another", anotherUser)

	boards, err := boardRepo.GetNestedBoards(rootBoard.Base.ID, user)
	assert.NoError(t, err, "Failed to get nested boards")
	assert.Equal(t, len(boards), 3)
	assert.Equal(t, boards[0].RelationID, nestedBoardOne.RelationID)
	assert.Equal(t, boards[1].RelationID, nestedBoardTwo.RelationID)
}

func TestBoardRepository_DeleteNestedBoard(t *testing.T) {
	store := teststore.New()
	userRepo := store.User()
	boardRepo := store.Board()

	// Create test assets
	user := model.TestUser(t)
	anotherUser := model.TestUser(t)
	anotherUser.Username += "a"
	anotherUser.Email += "a"
	_ = userRepo.SaveUser(user)
	_ = userRepo.SaveUser(anotherUser)

	// Add two nested boards as owner
	rootBoard, _ := boardRepo.NewRootBoard(user, "Root")
	nestedBoardOne, _ := boardRepo.NewNestedBoard(rootBoard.Base.ID, "Nested #one", user)
	nestedBoardTwo, _ := boardRepo.NewNestedBoard(rootBoard.Base.ID, "Nested #two", user)

	// Successful delete (as author)
	assert.NoError(t, boardRepo.DeleteNestedBoard(nestedBoardTwo.Base.ID, user), "Failed to delete nested board")
	boards, _ := boardRepo.GetNestedBoards(rootBoard.Base.ID, user)
	assert.Equal(t, len(boards), 1, "Board not deleted!")

	// Attempt to delete board as user without permissions
	assert.Error(t, boardRepo.DeleteNestedBoard(nestedBoardOne.Base.ID, anotherUser), "Failed to delete nested board")
}

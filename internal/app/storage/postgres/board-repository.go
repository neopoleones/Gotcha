package postgres

import (
	"Gotcha/internal/app/model"
	"github.com/google/uuid"
)

const (
	InsertBoardQuery = `
		INSERT INTO "Board"(title) VALUES($1) RETURNING id, created_at;
	`
	InsertBoardRelationQuery = `
		INSERT INTO "UserToBoard"(board_id, user_id, access_type, description)
			VALUES($1, $2, $3, $4) RETURNING id
	`
	GetBoardsOfUserQuery = `
		SELECT b.id, b.title, b.created_at, utb.id FROM "Board" b
			INNER JOIN "UserToBoard" utb ON utb.board_id = b.id
		WHERE utb.user_id = $1;
	`
	GetPermissionOfRelationQuery = `
		SELECT access_type from "UserToBoard" where id = $1;
	`
)

const (
	DescriptionAllGranted = "All privileges granted to owner"
)

type BoardRepository struct {
	store *Store
}

func (br *BoardRepository) NewRootBoard(user *model.User, title string) (*model.Board, error) {
	board := model.NewBoard(title)

	if err := board.Validate(); err != nil {
		return nil, err
	}

	// Save board
	if err := br.store.db.QueryRow(InsertBoardQuery, title).Scan(&board.ID, &board.CreatedAt); err != nil {
		return nil, err
	}

	// Create author relation
	rel, err := br.CreateRelation(board, user, DescriptionAllGranted, model.PrivilegeAuthor)
	if err != nil {
		return nil, err
	}
	board.AddRelation(rel)
	return board, nil
}

func (br *BoardRepository) GetRootBoardsOfUser(user *model.User) ([]*model.Board, error) {
	boardsMap := make(map[uuid.UUID]*model.Board)

	// Query for boards, close Row on function exit
	boardRows, err := br.store.db.Query(GetBoardsOfUserQuery, user.ID)
	if err != nil {
		return nil, err
	}
	defer boardRows.Close()

	// Iterate over all rows, add
	for boardRows.Next() {
		var relationID uuid.UUID
		board := model.NewBoard("default")

		// Just scan the row into board instance
		if err := boardRows.Scan(&board.ID, &board.Title, &board.CreatedAt, &relationID); err != nil {
			return nil, err
		}

		// Get all relations for boards
		savedBoard, found := boardsMap[board.ID]
		if !found {
			board.AddRelation(relationID)
			boardsMap[board.ID] = board
		} else {
			savedBoard.AddRelation(relationID)
		}
	}

	return mapBoardValues(boardsMap), nil
}

func (br *BoardRepository) CreateRelation(board *model.Board, user *model.User, desc string, ac model.PrivilegeType) (uuid.UUID, error) {
	var authorRelation uuid.UUID
	relationRow := br.store.db.QueryRow(InsertBoardRelationQuery, board.ID, user.ID, ac, desc)
	if err := relationRow.Scan(&authorRelation); err != nil {
		return uuid.Nil, err
	}
	return authorRelation, nil
}

func (br *BoardRepository) GetPrivilegeFromRelation(relationID uuid.UUID) (model.PrivilegeType, error) {
	var privilege model.PrivilegeType
	return privilege, br.store.db.QueryRow(GetPermissionOfRelationQuery, relationID).Scan(&privilege)
}

func mapBoardValues(boards map[uuid.UUID]*model.Board) []*model.Board {
	boardsSlice := make([]*model.Board, 0, len(boards))
	for _, val := range boards {
		boardsSlice = append(boardsSlice, val)
	}
	return boardsSlice
}

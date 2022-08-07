package postgres

import (
	"strings"

	"Gotcha/internal/app/model"
	"Gotcha/internal/app/storage"
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
		SELECT access_type, board_id, user_id FROM "UserToBoard" WHERE id = $1;
	`
	DeleteRelationsByBoardID = `
		DELETE FROM "UserToBoard" WHERE board_id = $1;
	`
	DeleteBoardsByID = `
		DELETE FROM "Board" WHERE id = $1;
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

func (br *BoardRepository) GetPrivilegeFromRelation(relationID uuid.UUID) (*model.BoardPermission, error) {
	bp := model.BoardPermission{}
	return &bp, br.store.db.QueryRow(GetPermissionOfRelationQuery, relationID).Scan(&bp.Privilege, &bp.BoardID, &bp.UserID)
}

func (br *BoardRepository) DeleteRootBoard(boardID uuid.UUID, relations []uuid.UUID, user *model.User) error {
	// Security check
	for _, relation := range relations {
		bp, err := br.GetPrivilegeFromRelation(relation)

		if boardID != bp.BoardID || bp.UserID != user.ID {
			return storage.ErrSecurityError
		}

		if err != nil {
			if strings.Contains(err.Error(), "no rows") {
				return storage.ErrNotFound
			}
			return storage.ErrSecurityError
		}
		// Then, we have permissions to delete a board
		if bp.Privilege == model.PrivilegeAuthor {
			if _, err := br.store.db.Exec(DeleteRelationsByBoardID, boardID); err != nil {
				return err
			}
			_, err := br.store.db.Exec(DeleteBoardsByID, boardID)
			return err
		}
	}

	// Not an author! Refuse request
	return storage.ErrSecurityError
}

func mapBoardValues(boards map[uuid.UUID]*model.Board) []*model.Board {
	boardsSlice := make([]*model.Board, 0, len(boards))
	for _, val := range boards {
		boardsSlice = append(boardsSlice, val)
	}
	return boardsSlice
}

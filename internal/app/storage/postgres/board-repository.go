package postgres

import (
	"database/sql"
	"errors"
	"strings"

	"Gotcha/internal/app/model"
	"Gotcha/internal/app/storage"
	"github.com/google/uuid"
)

const (
	InsertBoardQuery = `
		INSERT INTO "Board"(title) VALUES($1) RETURNING id, created_at;
	`
	GetRelationsOfBoardQuery = `
		SELECT utb.id, b.title, b.created_at FROM "Board" b
			INNER JOIN "UserToBoard" utb ON utb.board_id = b.id
		WHERE utb.board_id = $1;
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
	DeleteBoardByID = `
		DELETE FROM "Board" WHERE id = $1;
	`
	NewNestedBoardRelation = `
		INSERT INTO "BoardToBoard"(root_board_id, subboard_id) VALUES ($1, $2) returning id;
    `
	GetNestedBoardsQuery = `
		SELECT b2b.id, b2b.subboard_id, b.created_at, b.title from "BoardToBoard" b2b 
			inner join "Board" b on b.id = b2b.subboard_id
		where root_board_id = $1;
	`
	DeleteNestedBoardRelation = `
		DELETE FROM "BoardToBoard" where subboard_id = $1;
	`
	GetRootOfSideBoardQuery = `
    	SELECT root_board_id from "BoardToBoard" where subboard_id = $1;
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

	if err := board.Base.Validate(); err != nil {
		return nil, err
	}

	// Save board
	if err := br.store.db.QueryRow(InsertBoardQuery, title).Scan(&board.Base.ID, &board.Base.CreatedAt); err != nil {
		return nil, err
	}

	// Create author relation
	rel, err := br.CreateRelation(board.Base.ID, user.ID, DescriptionAllGranted, model.PrivilegeAuthor)
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
		if err := boardRows.Scan(&board.Base.ID, &board.Base.Title, &board.Base.CreatedAt, &relationID); err != nil {
			return nil, err
		}

		// Get all relations for boards
		savedBoard, found := boardsMap[board.Base.ID]
		if !found {
			board.AddRelation(relationID)
			boardsMap[board.Base.ID] = board
		} else {
			savedBoard.AddRelation(relationID)
		}
	}

	return mapBoardValues(boardsMap), nil
}

func (br *BoardRepository) CreateRelation(boardID, userID uuid.UUID, desc string, ac model.PrivilegeType) (uuid.UUID, error) {
	var authorRelation uuid.UUID
	relationRow := br.store.db.QueryRow(InsertBoardRelationQuery, boardID, userID, ac, desc)
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
			_, err := br.store.db.Exec(DeleteBoardByID, boardID)
			return err
		}
	}

	// Not an author! Refuse request
	return storage.ErrSecurityError
}

func (br *BoardRepository) GetBoardInfo(boardID uuid.UUID) (*model.Board, error) {
	board := model.NewBoard("default")
	board.Base.ID = boardID

	// Get relations
	relationsRows, err := br.store.db.Query(GetRelationsOfBoardQuery, boardID)
	if err != nil {
		return nil, err
	}
	for relationsRows.Next() {
		var relationID uuid.UUID
		if err := relationsRows.Scan(&relationID, &board.Base.Title, &board.Base.CreatedAt); err != nil {
			return nil, err
		}
		board.AddRelation(relationID)
	}
	return board, nil
}

func (br *BoardRepository) GetRootOfNestedBoard(boardID uuid.UUID) (*model.Board, error) {
	var rootID uuid.UUID

	err := br.store.db.QueryRow(GetRootOfSideBoardQuery, boardID).Scan(&rootID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return br.GetBoardInfo(boardID)
		}
		return nil, err
	}
	return br.GetRootOfNestedBoard(rootID)
}

func (br *BoardRepository) NewNestedBoard(rootBoardID uuid.UUID, title string, user *model.User) (*model.NestedBoard, error) {
	rootBoard, err := br.GetRootOfNestedBoard(rootBoardID)
	if err != nil {
		return nil, err
	}
	for _, rel := range rootBoard.U2BRelations {
		bp, err := br.GetPrivilegeFromRelation(rel)
		if err != nil {
			return nil, storage.ErrSecurityError
		}

		if (bp.Privilege == model.PrivilegeReadWrite || bp.Privilege == model.PrivilegeAuthor) && bp.UserID == user.ID {
			// Save the board
			nestedBoard := model.NestedBoard{
				Base: model.BaseBoard{
					Title: title,
				},
				RootBoard: rootBoardID,
			}

			err := br.store.db.QueryRow(InsertBoardQuery, title).Scan(&nestedBoard.Base.ID, &nestedBoard.Base.CreatedAt)
			if err != nil {
				return nil, err
			}
			err = br.store.db.QueryRow(NewNestedBoardRelation, rootBoardID, nestedBoard.Base.ID).Scan(&nestedBoard.RelationID)
			if err != nil {
				return nil, err
			}
			return &nestedBoard, nil
		}
	}
	return nil, storage.ErrSecurityError
}

func (br *BoardRepository) GetNestedBoards(rootBoardID uuid.UUID, user *model.User) ([]*model.NestedBoard, error) {
	rootBoard, err := br.GetRootOfNestedBoard(rootBoardID)
	if err != nil {
		return nil, err
	}
	for _, rel := range rootBoard.U2BRelations {
		bp, err := br.GetPrivilegeFromRelation(rel)
		if err != nil {
			return nil, storage.ErrSecurityError
		}

		// Any permission allows user to see the nested boards
		if bp.UserID == user.ID {
			boards := make([]*model.NestedBoard, 0)
			nestedBoardsRows, err := br.store.db.Query(GetNestedBoardsQuery, rootBoardID)
			if err != nil {
				return nil, err
			}

			for nestedBoardsRows.Next() {
				board := model.NestedBoard{
					Base:      model.BaseBoard{},
					RootBoard: rootBoardID,
				}
				if err := nestedBoardsRows.Scan(&board.RelationID, &board.Base.ID, &board.Base.CreatedAt, &board.Base.Title); err != nil {
					return nil, err
				}
				boards = append(boards, &board)

			}
			return boards, nil
		}
	}
	return nil, storage.ErrSecurityError
}

func (br *BoardRepository) DeleteNestedBoard(boardID uuid.UUID, user *model.User) error {
	rootBoard, err := br.GetRootOfNestedBoard(boardID)
	if err != nil {
		return err
	}
	for _, rel := range rootBoard.U2BRelations {
		bp, err := br.GetPrivilegeFromRelation(rel)
		if err != nil {
			return storage.ErrSecurityError
		}

		if (bp.Privilege == model.PrivilegeAuthor || bp.Privilege == model.PrivilegeReadWrite) && user.ID == bp.UserID {
			if _, err := br.store.db.Exec(DeleteNestedBoardRelation, boardID); err != nil {
				return err
			}
			_, err := br.store.db.Exec(DeleteBoardByID, boardID)
			return err
		}
	}
	return storage.ErrSecurityError
}

func mapBoardValues(boards map[uuid.UUID]*model.Board) []*model.Board {
	boardsSlice := make([]*model.Board, 0, len(boards))
	for _, val := range boards {
		boardsSlice = append(boardsSlice, val)
	}
	return boardsSlice
}

// **HELCODE**

package teststore

import (
	"fmt"
	"time"

	"Gotcha/internal/app/model"
	"Gotcha/internal/app/storage"
	"github.com/google/uuid"
)

type Relation struct {
	ID            uuid.UUID
	BoardID       uuid.UUID
	UserID        uuid.UUID
	Description   string
	privilegeType model.PrivilegeType
}

type NestedRelation struct {
	RelationID    uuid.UUID
	BoardID       uuid.UUID
	NestedBoardID uuid.UUID
}

type BoardRepository struct {
	storage         *Storage
	Relations       []*Relation
	NestedRelations map[uuid.UUID]*NestedRelation // root board: relation
	Boards          map[uuid.UUID]*model.Board
	NestedBoards    map[uuid.UUID]*model.NestedBoard
}

func (b *BoardRepository) NewRootBoard(user *model.User, title string) (*model.Board, error) {

	board := model.NewBoard(title)
	if err := board.Base.Validate(); err != nil {
		return nil, err
	}

	board.Base.CreatedAt = time.Now()
	board.Base.ID = uuid.New()
	b.Boards[board.Base.ID] = board

	_, err := b.CreateRelation(board.Base.ID, user.ID, "Root board", model.PrivilegeAuthor)
	if err != nil {
		return nil, err
	}

	return board, nil
}

func (b *BoardRepository) GetRootBoardsOfUser(user *model.User) ([]*model.Board, error) {
	boards := make([]*model.Board, 0, 2)
	for _, relation := range b.Relations {
		if relation.UserID == user.ID {
			boards = append(boards, b.Boards[relation.BoardID])
		}
	}
	return boards, nil
}

func (b *BoardRepository) GetPrivilegeFromRelation(relationID uuid.UUID) (*model.BoardPermission, error) {
	for _, relation := range b.Relations {
		if relation.ID == relationID {
			return &model.BoardPermission{
				BoardID:   relation.BoardID,
				UserID:    relation.UserID,
				Privilege: relation.privilegeType,
			}, nil
		}
	}
	return nil, storage.ErrNotFound
}

func (b *BoardRepository) CreateRelation(boardID, userID uuid.UUID, desc string, privilegeType model.PrivilegeType) (uuid.UUID, error) {
	rel := Relation{
		ID:            uuid.New(),
		BoardID:       boardID,
		Description:   desc,
		privilegeType: privilegeType,
		UserID:        userID,
	}
	fmt.Printf("Board in CreateRelation: %+v\n", boardID)
	b.Relations = append(b.Relations, &rel)
	b.Boards[boardID].AddRelation(rel.ID)

	return rel.ID, nil
}

func (b *BoardRepository) DeleteRootBoard(boardID uuid.UUID, relations []uuid.UUID, user *model.User) error {
	for _, givenRelation := range relations {
		currBoardPermission, err := b.GetPrivilegeFromRelation(givenRelation)

		// User attempted to bypass mitigations
		if currBoardPermission.BoardID != boardID || currBoardPermission.UserID != user.ID || err != nil {
			return storage.ErrSecurityError
		}
		if currBoardPermission.Privilege == model.PrivilegeAuthor {
			for i, rel := range b.Relations {
				if rel.BoardID == boardID {
					b.Relations = append(b.Relations[:i], b.Relations[i+1:]...)
				}
			}
			delete(b.Boards, boardID)
			return nil
		}
	}
	return storage.ErrSecurityError
}

func (b *BoardRepository) GetRootOfNestedBoard(nestedBoardID uuid.UUID) (*model.Board, error) {
	currRelation, found := b.NestedRelations[nestedBoardID]
	if !found {
		// If not found, then it's a root board
		return b.GetBoardInfo(nestedBoardID)
	}
	return b.GetRootOfNestedBoard(currRelation.BoardID)
}

func (b *BoardRepository) NewNestedBoard(rootBoardID uuid.UUID, title string, user *model.User) (*model.NestedBoard, error) {
	rootBoard, err := b.GetRootOfNestedBoard(rootBoardID)
	if err != nil {
		return nil, err
	}
	for _, rel := range rootBoard.U2BRelations {
		bp, _ := b.GetPrivilegeFromRelation(rel)
		if (bp.Privilege == model.PrivilegeReadWrite || bp.Privilege == model.PrivilegeAuthor) && bp.UserID == user.ID {
			// Create board
			nestedBoard := model.NestedBoard{
				Base: model.BaseBoard{
					Title:     title,
					ID:        uuid.New(),
					CreatedAt: time.Now(),
				},
				RootBoard: rootBoardID,
			}
			b.NestedBoards[nestedBoard.Base.ID] = &nestedBoard

			// Create relation
			rel := NestedRelation{
				RelationID:    uuid.New(),
				BoardID:       rootBoardID,
				NestedBoardID: nestedBoard.Base.ID,
			}
			b.NestedRelations[nestedBoard.Base.ID] = &rel
			nestedBoard.RelationID = rel.RelationID
			return &nestedBoard, nil
		}
	}

	return nil, storage.ErrSecurityError
}

func (b *BoardRepository) GetNestedBoards(rootBoardID uuid.UUID, user *model.User) ([]*model.NestedBoard, error) {
	rootBoard, err := b.GetRootOfNestedBoard(rootBoardID)
	if err != nil {
		return nil, err
	}
	for _, rel := range rootBoard.U2BRelations {
		bp, _ := b.GetPrivilegeFromRelation(rel)

		// Any permission allows user to see the nested boards
		if bp.UserID == user.ID {
			boards := make([]*model.NestedBoard, 0)
			for _, rel := range b.NestedRelations {
				if rel.BoardID == rootBoardID {
					boards = append(boards, b.NestedBoards[rel.NestedBoardID])
				}
			}
			return boards, nil
		}
	}
	return nil, storage.ErrSecurityError
}

func (b *BoardRepository) DeleteNestedBoard(boardID uuid.UUID, user *model.User) error {
	rootBoard, err := b.GetRootOfNestedBoard(boardID)
	if err != nil {
		return err
	}
	for _, rel := range rootBoard.U2BRelations {
		bp, _ := b.GetPrivilegeFromRelation(rel)
		if (bp.Privilege == model.PrivilegeAuthor || bp.Privilege == model.PrivilegeReadWrite) && user.ID == bp.UserID {
			delete(b.NestedRelations, boardID)
			delete(b.NestedBoards, boardID)
			return nil
		}
	}
	return storage.ErrSecurityError
}

func (b *BoardRepository) GetBoardInfo(boardID uuid.UUID) (*model.Board, error) {
	board, found := b.Boards[boardID]
	if !found {
		return nil, storage.ErrNotFound
	}
	return board, nil
}

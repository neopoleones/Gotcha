// **HELCODE**

package teststore

import (
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

type BoardRepository struct {
	storage   *Storage
	Relations []*Relation
	Boards    map[uuid.UUID]*model.Board
}

func (b *BoardRepository) NewRootBoard(user *model.User, title string) (*model.Board, error) {

	board := model.NewBoard(title)
	if err := board.Validate(); err != nil {
		return nil, err
	}

	board.Base.CreatedAt = time.Now()
	board.Base.ID = uuid.New()

	relID, _ := b.CreateRelation(board, user, "Root board", model.PrivilegeAuthor)

	board.AddRelation(relID)
	b.Boards[board.Base.ID] = board
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

func (b *BoardRepository) CreateRelation(board *model.Board, user *model.User, desc string, privilegeType model.PrivilegeType) (uuid.UUID, error) {
	rel := Relation{
		ID:            uuid.New(),
		BoardID:       board.Base.ID,
		Description:   desc,
		privilegeType: privilegeType,
		UserID:        user.ID,
	}

	b.Relations = append(b.Relations, &rel)
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

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

	board.CreatedAt = time.Now()
	board.ID = uuid.New()

	relID, _ := b.CreateRelation(board, user, "Root board", model.PrivilegeAuthor)

	board.AddRelation(relID)
	b.Boards[board.ID] = board
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

func (b *BoardRepository) GetPrivilegeFromRelation(relationID uuid.UUID) (model.PrivilegeType, error) {
	for _, relation := range b.Relations {
		if relation.ID == relationID {
			return relation.privilegeType, nil
		}
	}
	return -1, storage.ErrNotFound
}

func (b *BoardRepository) CreateRelation(board *model.Board, user *model.User, desc string, privilegeType model.PrivilegeType) (uuid.UUID, error) {
	rel := Relation{
		ID:            uuid.New(),
		BoardID:       board.ID,
		Description:   desc,
		privilegeType: privilegeType,
		UserID:        user.ID,
	}

	b.Relations = append(b.Relations, &rel)
	return rel.ID, nil
}

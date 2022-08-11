package storage

import (
	"Gotcha/internal/app/model"
	"github.com/google/uuid"
)

type UserRepository interface {
	FindUserBySobriquet(sobriquet string) (*model.User, error)
	FindUserByID(userID uuid.UUID) (*model.User, error)
	SaveUser(user *model.User) error
	GetAllUsers(user *model.User) ([]*model.User, error)
}

type BoardRepository interface {
	NewRootBoard(user *model.User, title string) (*model.Board, error)
	GetRootBoardsOfUser(user *model.User) ([]*model.Board, error)
	GetPrivilegeFromRelation(relationID uuid.UUID) (*model.BoardPermission, error)
	DeleteRootBoard(boardID uuid.UUID, relations []uuid.UUID, user *model.User) error
	GetRootOfNestedBoard(boardID uuid.UUID) (*model.Board, error)
	NewNestedBoard(rootBoardID uuid.UUID, title string, user *model.User) (*model.NestedBoard, error)
	GetNestedBoards(rootBoardID uuid.UUID, user *model.User) ([]*model.NestedBoard, error)
	DeleteNestedBoard(boardID uuid.UUID, user *model.User) error
	GetBoardInfo(boardID uuid.UUID) (*model.Board, error)
	CreateRelation(boardID, userID uuid.UUID, desc string, privilegeType model.PrivilegeType) (uuid.UUID, error)
}

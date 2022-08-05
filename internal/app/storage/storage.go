package storage

import (
	"errors"

	"Gotcha/internal/app/model"
	"github.com/google/uuid"
)

var (
	ErrNotFound = errors.New("record not found")
)

type UserRepository interface {
	FindUserBySobriquet(sobriquet string) (*model.User, error)
	FindUserByID(userID uuid.UUID) (*model.User, error)
	SaveUser(user *model.User) error
}

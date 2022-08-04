package storage

import (
	"errors"

	"Gotcha/internal/app/model"
)

var (
	ErrNotFound = errors.New("record not found")
)

type UserRepository interface {
	FindUserBySobriquet(sobriquet string) (*model.User, error)
	SaveUser(user *model.User) error
}

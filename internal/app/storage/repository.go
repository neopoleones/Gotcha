package storage

import (
	"Gotcha/internal/app/model"
	"github.com/google/uuid"
)

type UserRepository interface {
	FindUserBySobriquet(sobriquet string) (*model.User, error)
	FindUserByID(userID uuid.UUID) (*model.User, error)
	SaveUser(user *model.User) error
}

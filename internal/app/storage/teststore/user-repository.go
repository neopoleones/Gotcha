package teststore

import (
	"time"

	"Gotcha/internal/app/model"
	"Gotcha/internal/app/storage"
	"github.com/google/uuid"
)

type UserRepository struct {
	storage *Storage
	users   map[uuid.UUID]*model.User
}

// FindUserBySobriquet is very slow, but still usable for tests
func (u UserRepository) FindUserBySobriquet(sobriquet string) (*model.User, error) {

	for _, user := range u.users {
		if sobriquet == user.Email || sobriquet == user.Username {
			return user, nil
		}
	}
	return nil, storage.ErrNotFound
}

func (u UserRepository) FindUserByID(userID uuid.UUID) (*model.User, error) {
	user, found := u.users[userID]

	if !found {
		return nil, storage.ErrNotFound
	}
	return user, nil
}

func (u UserRepository) SaveUser(user *model.User) error {
	// Some sanity checks before saving
	if err := user.Validate(); err != nil {
		return err
	}
	if err := user.BeforeCreate(); err != nil {
		return err
	}
	if _, err := u.FindUserBySobriquet(user.Username); err == nil {
		return storage.ErrEntityDuplicate
	}
	if _, err := u.FindUserBySobriquet(user.Email); err == nil {
		return storage.ErrEntityDuplicate
	}

	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	u.users[user.ID] = user
	return nil
}

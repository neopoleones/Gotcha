package teststore

import (
	"Gotcha/internal/app/model"
	"Gotcha/internal/app/storage"
	"github.com/google/uuid"
)

type Storage struct {
	// Repositories
	userRepository *UserRepository
}

// New ...
func New() *Storage {
	return &Storage{}
}

func (storage *Storage) User() storage.UserRepository {
	if storage.userRepository == nil {
		storage.userRepository = &UserRepository{
			storage,
			make(map[uuid.UUID]*model.User),
		}
	}
	return storage.userRepository
}

func (storage *Storage) Close() {
	// ... implementation requirement
}

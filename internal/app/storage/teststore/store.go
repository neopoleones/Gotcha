package teststore

import (
	"Gotcha/internal/app/model"
	"Gotcha/internal/app/storage"
	"github.com/google/uuid"
)

type Storage struct {
	// Repositories
	userRepository  *UserRepository
	boardRepository *BoardRepository
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

func (storage *Storage) Board() storage.BoardRepository {
	if storage.boardRepository == nil {
		storage.boardRepository = &BoardRepository{
			storage,
			make([]*Relation, 0),
			make(map[uuid.UUID]*model.Board),
		}
	}
	return storage.boardRepository
}

func (storage *Storage) Close() {
	// ... implementation requirement
}

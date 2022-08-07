package postgres

import (
	"database/sql"

	"Gotcha/internal/app/storage"
)

// Store is an SQL(postgresql tested) implementation of gotcha storage
type Store struct {
	db              *sql.DB
	userRepository  *UserRepository
	boardRepository *BoardRepository
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

// User returns UserRepository for related operations. Implementation requirement
func (store *Store) User() storage.UserRepository {
	if store.userRepository == nil {
		store.userRepository = &UserRepository{store: store}
	}
	return store.userRepository
}

func (store *Store) Board() storage.BoardRepository {
	if store.boardRepository == nil {
		store.boardRepository = &BoardRepository{store: store}
	}
	return store.boardRepository
}

func (store *Store) Close() {
	// TODO: Add hooks
	// ...
	_ = store.db.Close()
}

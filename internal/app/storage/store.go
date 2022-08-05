package storage

import (
	"errors"
)

type Storage interface {
	User() UserRepository
	Close()
}

var (
	ErrNotFound        = errors.New("entity not found")
	ErrEntityDuplicate = errors.New("entity duplicate")
)

package storage

import (
	"errors"
)

type Storage interface {
	Board() BoardRepository
	User() UserRepository
	Close()
}

var (
	ErrNotFound        = errors.New("entity not found")
	ErrEntityDuplicate = errors.New("entity duplicate")
	ErrSecurityError   = errors.New("not permitted")
)

const (
	SessionsStoreRedis = "redis"
)

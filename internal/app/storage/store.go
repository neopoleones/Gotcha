package storage

type Storage interface {
	User() UserRepository
	Close()
}

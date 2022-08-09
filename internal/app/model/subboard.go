package model

import (
	"time"

	"github.com/google/uuid"
)

type BaseBoard struct {
	Title     string    `json:"title"`
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type SubBoard struct {
	Base      BaseBoard
	RootBoard uuid.UUID
}

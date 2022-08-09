package model

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/google/uuid"
)

type PrivilegeType int

const (
	_ PrivilegeType = iota
	PrivilegeAuthor
	PrivilegeReadOnly
	PrivilegeReadWrite
)

type Board struct {
	Base         BaseBoard
	U2BRelations []uuid.UUID `json:"relations"`
}

type BoardPermission struct {
	BoardID   uuid.UUID
	UserID    uuid.UUID
	Privilege PrivilegeType
}

func NewBoard(title string) *Board {
	return &Board{
		U2BRelations: make([]uuid.UUID, 0, 4),
		Base:         BaseBoard{Title: title},
	}
}

func (b *Board) AddRelation(uuid uuid.UUID) {
	b.U2BRelations = append(b.U2BRelations, uuid)
}

func (b *Board) Validate() error {
	return validation.Validate(b.Base.Title, validation.Length(1, 255))
}

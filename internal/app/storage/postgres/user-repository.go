package postgres

import (
	"database/sql"
	"strings"

	"Gotcha/internal/app/model"
	"Gotcha/internal/app/storage"
	"github.com/google/uuid"
)

const (
	saveUserQuery = `
		INSERT INTO "Users"(username, email, hash)
			VALUES($1, $2, $3) RETURNING id, created_at;
	`
	findUserByQuery = `
		SELECT id, username, email, hash, created_at FROM "Users" where username = $1 or email = $1;
	`
	findUserByIDQuery = `
		SELECT id, username, email, hash, created_at FROM "Users" where id = $1;
	`
	getAllUsers = `
		SELECT id, username, created_at FROM "Users";
	`
)

// UserRepository interface implementation (depends on SQL database)
type UserRepository struct {
	store *Store
}

// FindUserBySobriquet performs a simple search query by email and username.
// Returns error if user not found
func (repo *UserRepository) FindUserBySobriquet(sobriquet string) (*model.User, error) {
	u := model.User{}
	userRow := repo.store.db.QueryRow(findUserByQuery, sobriquet)
	if err := userRow.Scan(&u.ID, &u.Username, &u.Email, &u.Hash, &u.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

// SaveUser performs validation check, gets hash of password and then saves the user
func (repo *UserRepository) SaveUser(user *model.User) error {
	if err := user.Validate(); err != nil {
		return err
	}
	if err := user.BeforeCreate(); err != nil {
		return err
	}

	resultRow := repo.store.db.QueryRow(saveUserQuery, user.Username, user.Email, user.Hash)
	err := resultRow.Scan(&user.ID, &user.CreatedAt)
	if err != nil && strings.Contains(err.Error(), "duplicate") {
		return storage.ErrEntityDuplicate
	}
	return err
}

func (repo *UserRepository) FindUserByID(userID uuid.UUID) (*model.User, error) {
	u := model.User{}
	userRow := repo.store.db.QueryRow(findUserByIDQuery, userID)
	if err := userRow.Scan(&u.ID, &u.Username, &u.Email, &u.Hash, &u.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (repo *UserRepository) GetAllUsers(currUser *model.User) ([]*model.User, error) {
	users := make([]*model.User, 0)
	usrRows, err := repo.store.db.Query(getAllUsers)
	if err != nil {
		return nil, err
	}

	for usrRows.Next() {
		user := model.User{}
		err := usrRows.Scan(&user.ID, &user.Username, &user.CreatedAt)
		if err != nil {
			return nil, err
		}

		// Don't include current user
		if user.ID == currUser.ID {
			continue
		}
		users = append(users, &user)
	}
	return users, nil
}

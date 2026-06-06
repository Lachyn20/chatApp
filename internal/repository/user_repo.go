package repository

import (
	"chatApp/internal/model"
	"context"
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var ErrEmailAlreadyExists = errors.New("email already exists")

type UserRepository interface {
	Register(ctx context.Context, user model.User) (int, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
}

type userRepo struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepo{db: db}  // db data bilen create function 
}

func (r *userRepo) Register(ctx context.Context, user model.User) (int, error) {
	var id int
	err := r.db.QueryRowContext(ctx, "SELECT id FROM users WHERE email = $1", user.Email).Scan(&id)
	if err == nil {
		return 0, ErrEmailAlreadyExists
	}
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	query := `
		INSERT INTO users (username, password, email)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	err = r.db.QueryRowContext(ctx, query, user.Username, string(hashedPassword), user.Email).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, username, password, email
		FROM users
		WHERE email = $1
	`
	
	var user model.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Username, &user.Password, &user.Email)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
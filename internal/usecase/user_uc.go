package usecase

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"chatApp/internal/model"
	"chatApp/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type userUsecase struct {
	repo             repository.UserRepository
	jwtSecret        string
	tokenExpiryHours int
}

type UserUsecase interface {
	Register(ctx context.Context, user model.User) (string, error)
	Login(ctx context.Context, email, password string) (string, error)
}

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrEmailAlreadyExists  = errors.New("email already exists")
	ErrMissingCredentials  = errors.New("username, password and email are required")
	ErrMissingLoginDetails = errors.New("email and password are required")
)

func NewUserUsecase(repo repository.UserRepository, jwtSecret string, tokenExpiryHours int) UserUsecase {
	return &userUsecase{
		repo:             repo,
		jwtSecret:        jwtSecret,
		tokenExpiryHours: tokenExpiryHours,
	}
}

func (u *userUsecase) Register(ctx context.Context, user model.User) (string, error) {
	if strings.TrimSpace(user.Username) == "" || strings.TrimSpace(user.Password) == "" || strings.TrimSpace(user.Email) == "" {
		return "", ErrMissingCredentials
	}

	userID, err := u.repo.Register(ctx, user)
	if err != nil {
		if errors.Is(err, repository.ErrEmailAlreadyExists) {
			return "", ErrEmailAlreadyExists
		}
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  userID,
		"username": user.Username,
		"email":    user.Email,
		"exp":      time.Now().Add(time.Duration(u.tokenExpiryHours) * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(u.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (u *userUsecase) Login(ctx context.Context, email, password string) (string, error) {
	if strings.TrimSpace(email) == "" || strings.TrimSpace(password) == "" {
		return "", ErrMissingLoginDetails
	}

	user, err := u.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid password")
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
		"exp":      time.Now().Add(time.Duration(u.tokenExpiryHours) * time.Hour).Unix(),
	})
	
	tokenString, err := token.SignedString([]byte(u.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil

}
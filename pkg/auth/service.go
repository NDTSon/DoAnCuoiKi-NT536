package auth

import (
	"context"
	"database/sql"
	"errors"
	"github.com/livekit/livekit-server/pkg/storage"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type Service struct {
	users *storage.UserRepository
}

func NewService(users *storage.UserRepository) *Service {
	return &Service{users: users}
}

func (s *Service) Register(ctx context.Context, email, password, displayName string) (*storage.User, error) {
	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	user := &storage.User{
		Email:        email,
		PasswordHash: hash,
		DisplayName:  sql.NullString{String: displayName, Valid: displayName != ""},
	}
	if err := s.users.CreateUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*storage.User, error) {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if err := CheckPassword(user.PasswordHash, password); err != nil {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

package service

import (
	"context"
	"errors"
	"time"

	"go_postgres/internal/models"
	"go_postgres/internal/repository"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
)

type CreateUserRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type UpdateUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Password  string `json:"password,omitempty"`
}

type UserResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserService interface {
	CreateUser(ctx context.Context, req CreateUserRequest) (*UserResponse, error)
	GetUser(ctx context.Context, id uint) (*UserResponse, error)
	ListUsers(ctx context.Context, page, pageSize int) ([]*UserResponse, int64, error)
	UpdateUser(ctx context.Context, id uint, req UpdateUserRequest) (*UserResponse, error)
	DeleteUser(ctx context.Context, id uint) error
	AuthenticateUser(ctx context.Context, email, password string) (*UserResponse, error)
}

type DefaultUserService struct {
	repo   repository.UserRepository
	logger *zap.Logger
}

func NewUserService(repo repository.UserRepository, logger *zap.Logger) UserService {
	return &DefaultUserService{
		repo:   repo,
		logger: logger,
	}
}

func (s *DefaultUserService) CreateUser(ctx context.Context, req CreateUserRequest) (*UserResponse, error) {
	_, err := s.repo.GetByEmail(ctx, req.Email)
	if err == nil {
		return nil, ErrUserAlreadyExists
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("failed to hash password", zap.Error(err))
		return nil, err
	}

	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		IsActive:     true,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return nil, ErrUserAlreadyExists
		}
		return nil, err
	}

	return s.mapUserToResponse(user), nil
}

func (s *DefaultUserService) GetUser(ctx context.Context, id uint) (*UserResponse, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return s.mapUserToResponse(user), nil
}

func (s *DefaultUserService) ListUsers(ctx context.Context, page, pageSize int) ([]*UserResponse, int64, error) {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	users, count, err := s.repo.List(ctx, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	var userResponse []*UserResponse
	for _, user := range users {
		userResponse = append(userResponse, s.mapUserToResponse(user))
	}

	return userResponse, count, nil
}

func (s *DefaultUserService) UpdateUser(ctx context.Context, id uint, req UpdateUserRequest) (*UserResponse, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Update fields
	user.FirstName = req.FirstName
	user.LastName = req.LastName

	// Update password if provided
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			s.logger.Error("Failed to hash password", zap.Error(err))
			return nil, err
		}
		user.PasswordHash = string(hashedPassword)
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return s.mapUserToResponse(user), nil
}

func (s *DefaultUserService) DeleteUser(ctx context.Context, id uint) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	return nil
}

func (s *DefaultUserService) AuthenticateUser(ctx context.Context, email, password string) (*UserResponse, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.mapUserToResponse(user), nil
}

func (s *DefaultUserService) mapUserToResponse(user *models.User) *UserResponse {
	return &UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

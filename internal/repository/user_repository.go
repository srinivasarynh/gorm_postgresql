package repository

import (
	"context"
	"errors"

	"go_postgres/internal/models"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrNotFound = errors.New("record not found")
	ErrConflict = errors.New("record already exists")
	ErrDatabase = errors.New("database error")
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uint) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	List(ctx context.Context, offset, limit int) ([]*models.User, int64, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uint) error
}

type GormUserRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewUserRepository(db *gorm.DB, logger *zap.Logger) UserRepository {
	return &GormUserRepository{
		db:     db,
		logger: logger,
	}
}

func (r *GormUserRepository) Create(ctx context.Context, user *models.User) error {
	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		if r.isUniqueConstraintError(result.Error) {
			return ErrConflict
		}
		r.logger.Error("failed to create user", zap.Error(result.Error))
		return ErrDatabase
	}
	return nil
}

func (r *GormUserRepository) GetByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}

		r.logger.Error("failed to get user by ID", zap.Error(result.Error))
		return nil, ErrDatabase
	}

	return &user, nil
}

func (r *GormUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		r.logger.Error("Failed to get user by email", zap.Error(result.Error))
		return nil, ErrDatabase
	}
	return &user, nil
}

func (r *GormUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).Where("username = ?", username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		r.logger.Error("Failed to get user by username", zap.Error(result.Error))
		return nil, ErrDatabase
	}
	return &user, nil
}

func (r *GormUserRepository) List(ctx context.Context, offset, limit int) ([]*models.User, int64, error) {
	var users []*models.User
	var count int64

	// Count total records
	if err := r.db.WithContext(ctx).Model(&models.User{}).Count(&count).Error; err != nil {
		r.logger.Error("Failed to count users", zap.Error(err))
		return nil, 0, ErrDatabase
	}

	// Get paginated records
	result := r.db.WithContext(ctx).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&users)

	if result.Error != nil {
		r.logger.Error("Failed to list users", zap.Error(result.Error))
		return nil, 0, ErrDatabase
	}

	return users, count, nil
}

func (r *GormUserRepository) Update(ctx context.Context, user *models.User) error {
	result := r.db.WithContext(ctx).Save(user)
	if result.Error != nil {
		if r.isUniqueConstraintError(result.Error) {
			return ErrConflict
		}
		r.logger.Error("Failed to update user", zap.Error(result.Error))
		return ErrDatabase
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *GormUserRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.User{}, id)
	if result.Error != nil {
		r.logger.Error("Failed to delete user", zap.Error(result.Error))
		return ErrDatabase
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *GormUserRepository) isUniqueConstraintError(err error) bool {
	// Postgres unique constraint error code
	return err != nil && err.Error() == "ERROR: duplicate key value violates unique constraint"
}

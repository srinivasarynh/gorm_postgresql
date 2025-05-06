package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in our system
type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Username     string         `gorm:"size:50;uniqueIndex;not null" json:"username"`
	Email        string         `gorm:"size:100;uniqueIndex;not null" json:"email"`
	PasswordHash string         `gorm:"size:100;not null" json:"-"` // Never expose in JSON
	FirstName    string         `gorm:"size:50" json:"first_name"`
	LastName     string         `gorm:"size:50" json:"last_name"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"` // Support for soft delete
}

// TableName specifies the table name for the User model
func (User) TableName() string {
	return "app_users"
}

// BeforeCreate is a GORM hook that runs before creating a record
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// You can implement any pre-save logic here
	// For example, hashing the password (although this should be done at the service level)
	return nil
}

// BeforeUpdate is a GORM hook that runs before updating a record
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	// You can implement any pre-update logic here
	return nil
}

package core

import (
	"time"

	"github.com/google/uuid"
)

// UserSign maps to core.user_sign.
type UserSign struct {
	UserID       uuid.UUID  `gorm:"column:user_id;type:uuid;primaryKey" json:"userId"`
	Username     string     `gorm:"column:username;type:varchar(24);not null" json:"username"`
	PasswordHash string     `gorm:"column:password_hash;type:text;not null" json:"-"`
	UpdatedAt    *time.Time `gorm:"column:updated_at;type:timestamp" json:"updatedAt,omitempty"`
}

func (UserSign) TableName() string {
	return "core.user_sign"
}

package sso

import (
	"time"

	"github.com/google/uuid"
)

// AccessToken maps to sso.access_token (DB 백업·조회용).
type AccessToken struct {
	UserID    uuid.UUID `gorm:"column:user_id;type:uuid;not null;index" json:"userId"`
	Token     string    `gorm:"column:token;type:text;primaryKey;not null" json:"-"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;not null;autoCreateTime" json:"createdAt"`
	ExpiresAt time.Time `gorm:"column:expires_at;type:timestamp;not null" json:"expiresAt"`
}

func (AccessToken) TableName() string {
	return "sso.access_token"
}

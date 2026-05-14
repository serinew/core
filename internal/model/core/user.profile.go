package core

import (
	"time"

	"github.com/google/uuid"
)

// UserProfile maps to core.user_profile.
type UserProfile struct {
	UserID          uuid.UUID  `gorm:"column:user_id;type:uuid;primaryKey" json:"userId"`
	UserName        *string    `gorm:"column:user_name;type:varchar(256)" json:"userName,omitempty"`
	NickName        *string    `gorm:"column:nick_name;type:varchar(256)" json:"nickName,omitempty"`
	ProfileImageURL *string    `gorm:"column:profile_image_url;type:text" json:"profileImageUrl,omitempty"`
	UpdatedAt       *time.Time `gorm:"column:updated_at;type:timestamp" json:"updatedAt,omitempty"`
}

func (UserProfile) TableName() string {
	return "core.user_profile"
}

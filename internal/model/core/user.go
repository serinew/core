package core

import (
	"time"

	"github.com/google/uuid"
)

// User maps to core.user (스크린샷 스키마 기준).
type User struct {
	ID        uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name      string    `gorm:"column:name;type:varchar(256);not null" json:"name"`
	Email     string    `gorm:"column:email;type:varchar(64);not null" json:"email"`
	Phone     string    `gorm:"column:phone;type:varchar(24);not null" json:"phone"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;not null;autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;not null;autoUpdateTime" json:"updatedAt"`
}

func (User) TableName() string {
	return "core.user"
}

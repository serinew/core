package core

import (
	"github.com/google/uuid"
)

// Admin maps to core.admin.
type Admin struct {
	UserId    uuid.UUID `gorm:"column:user_id;type:uuid;primaryKey" json:"userId"`
	Authority string    `gorm:"column:authority;type:varchar(12);not null" json:"authority"`
}

func (Admin) TableName() string {
	return "core.admin"
}

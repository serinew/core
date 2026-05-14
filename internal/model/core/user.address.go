package core

import (
	"time"

	"github.com/google/uuid"
)

// UserAddress maps to core.user_address.
type UserAddress struct {
	UserID    uuid.UUID `gorm:"column:user_id;type:uuid;primaryKey" json:"userId"`
	Addr1     string    `gorm:"column:addr1;type:varchar(256);not null" json:"addr1"`
	Addr2     *string   `gorm:"column:addr2;type:varchar(256)" json:"addr2,omitempty"`
	ZipCode   int16     `gorm:"column:zip_code;type:smallint;not null" json:"zipCode"`
	Lat       *float64  `gorm:"column:lat;type:double precision" json:"lat,omitempty"`
	Lng       *float64  `gorm:"column:lng;type:double precision" json:"lng,omitempty"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;not null;autoCreateTime" json:"createdAt"`
}

func (UserAddress) TableName() string {
	return "core.user_address"
}

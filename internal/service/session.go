package service

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/serinew/core/internal/model/core"
	"github.com/serinew/core/internal/repository"
	"github.com/serinew/core/internal/types"
)

// LoadSignLoginData는 사용자·주소·프로필을 로그인 응답(SignLoginData) 형태로 조회합니다.
func LoadSignLoginData(store *repository.Store, userID uuid.UUID) (*types.SignLoginData, error) {
	u, err := repository.Repo[core.User](store, "core.user")().SelectByID(userID)
	if err != nil {
		return nil, err
	}
	var addresses []core.UserAddress
	if err := repository.Repo[core.UserAddress](store, "core.user_address")().
		FindWhere(&addresses, "user_id = ?", userID); err != nil {
		return nil, err
	}
	var profile core.UserProfile
	err = repository.Repo[core.UserProfile](store, "core.user_profile")().
		Query().Where("user_id = ?", userID).Take(&profile).Error
	out := &types.SignLoginData{
		User:    *u,
		Address: addresses,
		Profile: nil,
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return out, nil
		}
		return nil, err
	}
	out.Profile = &profile
	return out, nil
}

package service

import (
	"github.com/google/uuid"

	"github.com/serinew/core/internal/model/core"
	"github.com/serinew/core/internal/repository"
	"github.com/serinew/core/internal/types"
)

var userListSearchColumns = []string{"name", "email", "phone"}

// UserService는 core.user 조회 등 사용자 도메인 유스케이스를 다룹니다.
type UserService struct {
	store *repository.Store
	users func() *repository.RepoConfig[core.User]
}

// NewUserService는 저장소 바인딩으로 UserService를 만듭니다.
func NewUserService(store *repository.Store) *UserService {
	return &UserService{
		store: store,
		users: repository.Repo[core.User](store, "core.user"),
	}
}

// GetByID는 PK id로 사용자 한 건을 읽습니다.
func (s *UserService) GetByID(id uuid.UUID) (*core.User, error) {
	return s.users().SelectByID(id)
}

// GetSignLoginByID는 로그인 성공 응답과 같은 형태(user·address[]·profile)로 조회합니다.
func (s *UserService) GetSignLoginByID(id uuid.UUID) (*types.SignLoginData, error) {
	return LoadSignLoginData(s.store, id)
}

// List는 페이지·검색·정렬 조건으로 사용자 목록과 필터에 맞는 전체 건수(count)를 반환합니다.
// opts가 nil이면 Page=1, PageSize=20, Search 없음과 동등합니다 (repository.Pagination).
func (s *UserService) List(opts *repository.ReadOpts) ([]core.User, int64, error) {
	return s.users().FindPage(opts, userListSearchColumns)
}

// ListSignLogin은 List와 같은 필터·페이지 규칙으로 조회하고, 각 행을 SignLoginData(주소·프로필 포함)로 만듭니다.
// 페이지에 포함된 사용자 id 집합에 대해 주소·프로필을 IN 쿼리로 한 번씩 불러오므로 N+1은 없습니다.
func (s *UserService) ListSignLogin(opts *repository.ReadOpts) ([]types.SignLoginData, int64, error) {
	rows, total, err := s.List(opts)
	if err != nil {
		return nil, 0, err
	}
	if len(rows) == 0 {
		return []types.SignLoginData{}, total, nil
	}
	ids := make([]uuid.UUID, len(rows))
	for i := range rows {
		ids[i] = rows[i].ID
	}

	var addresses []core.UserAddress
	addrRepo := repository.Repo[core.UserAddress](s.store, "core.user_address")()
	if err := addrRepo.Query().Where("user_id IN ?", ids).Find(&addresses).Error; err != nil {
		return nil, 0, err
	}

	var profiles []core.UserProfile
	profRepo := repository.Repo[core.UserProfile](s.store, "core.user_profile")()
	if err := profRepo.Query().Where("user_id IN ?", ids).Find(&profiles).Error; err != nil {
		return nil, 0, err
	}

	byAddr := make(map[uuid.UUID][]core.UserAddress, len(rows))
	for _, a := range addresses {
		byAddr[a.UserID] = append(byAddr[a.UserID], a)
	}
	byProf := make(map[uuid.UUID]*core.UserProfile, len(profiles))
	for i := range profiles {
		p := &profiles[i]
		byProf[p.UserID] = p
	}

	byAdmin, err := loadSignLoginAdminsByUserIDs(s.store, ids)
	if err != nil {
		return nil, 0, err
	}

	out := make([]types.SignLoginData, len(rows))
	for i := range rows {
		u := rows[i]
		addr := byAddr[u.ID]
		if addr == nil {
			addr = []core.UserAddress{}
		}
		out[i] = types.SignLoginData{
			User:    u,
			Address: addr,
			Profile: byProf[u.ID],
			Admin:   byAdmin[u.ID],
		}
	}
	return out, total, nil
}

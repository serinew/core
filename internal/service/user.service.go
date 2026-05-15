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

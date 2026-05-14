package service

import (
	"github.com/google/uuid"

	"github.com/serinew/core/internal/model/core"
	"github.com/serinew/core/internal/repository"
)

var userListSearchColumns = []string{"name", "email", "phone"}

// UserService는 core.user 조회 등 사용자 도메인 유스케이스를 다룹니다.
type UserService struct {
	users func() *repository.RepoConfig[core.User]
}

// NewUserService는 저장소 바인딩으로 UserService를 만듭니다.
func NewUserService(store *repository.Store) *UserService {
	return &UserService{
		users: repository.Repo[core.User](store, "core.user"),
	}
}

// GetByID는 PK id로 사용자 한 건을 읽습니다.
func (s *UserService) GetByID(id uuid.UUID) (*core.User, error) {
	return s.users().SelectByID(id)
}

// List는 페이지·검색·정렬 조건으로 사용자 목록과 전체 건수를 반환합니다.
// opts가 nil이면 Page=1, PageSize=20, Search 없음과 동등합니다 (repository.Pagination).
func (s *UserService) List(opts *repository.ReadOpts) ([]core.User, int64, error) {
	return s.users().FindPage(opts, userListSearchColumns)
}

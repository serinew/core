package users

import (
	"math"

	"github.com/gin-gonic/gin"

	userroute "github.com/serinew/core/internal/app/v1/users/userId"
	"github.com/serinew/core/internal/config"
	"github.com/serinew/core/internal/middleware"
	"github.com/serinew/core/internal/repository"
	"github.com/serinew/core/internal/service"
	"github.com/serinew/core/internal/types"
	. "github.com/serinew/core/internal/util/http"
)

// UsersRoutes는 /v1/users 하위 전부에 accessToken 쿠키 인증을 적용합니다 (Express router.use와 유사).
func UsersRoutes(r *gin.RouterGroup, store *repository.Store, cfg config.Config) {
	tokens := service.NewTokenService(store, cfg.TokenSecret)
	r.Use(middleware.RequireAccessTokenCookie(tokens))

	svc := service.NewUserService(store)
	h := &userHandlers{svc: svc}

	r.GET("/", h.List)

	userroute.Routes(r, svc)
}

type userHandlers struct {
	svc *service.UserService
}

// List는 페이지·검색 조건으로 사용자 목록을 반환합니다 (각 행에 address·profile 포함).
//
// @Summary 사용자 목록
// @Tags users
// @Security CookieAuth
// @Produce json
// @Param page query int false "페이지 (기본 1)"
// @Param pageSize query int false "페이지 크기 (기본 20, 최대 100)"
// @Param search query string false "검색어"
// @Success 200 {object} types.SuccessEnvelopeSignLoginList
// @Failure 401 {object} types.ErrorDoc
// @Failure 500 {object} types.ErrorDoc
// @Router /users [get]
func (h *userHandlers) List(c *gin.Context) {
	opts := repository.ReadOptsFromListQuery(types.FetchListQuery(c))
	rows, countTotal, err := h.svc.ListSignLogin(opts)
	if err != nil {
		Http.InternalError(c, &ErrOpts{Message: err.Error()})
		return
	}
	var cn int
	if countTotal > int64(math.MaxInt) {
		cn = math.MaxInt
	} else {
		cn = int(countTotal)
	}
	Http.OK(c, &SuccOpts{
		Data:  rows,
		Count: &cn,
	})
}

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

	r.GET("/", func(c *gin.Context) {
		opts := repository.ReadOptsFromListQuery(types.FetchListQuery(c))
		rows, countTotal, err := svc.List(opts)
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
	})

	userroute.Routes(r, svc)
}

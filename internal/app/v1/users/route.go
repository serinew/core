package users

import (
	"github.com/gin-gonic/gin"

	userroute "github.com/serinew/core/internal/app/v1/users/userId"
	"github.com/serinew/core/internal/repository"
	"github.com/serinew/core/internal/service"
	"github.com/serinew/core/internal/types"
	. "github.com/serinew/core/internal/util/http"
)

func UsersRoutes(r *gin.RouterGroup, store *repository.Store) {
	svc := service.NewUserService(store)

	r.GET("/", func(c *gin.Context) {
		opts := repository.ReadOptsFromListQuery(types.FetchListQuery(c))
		rows, total, err := svc.List(opts)
		if err != nil {
			Http.InternalError(c, &ErrOpts{Message: err.Error()})
			return
		}
		Http.OK(c, &SuccOpts{Data: gin.H{"items": rows, "total": total}})
	})

	userroute.Routes(r, svc)
}

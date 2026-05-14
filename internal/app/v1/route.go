package v1

import (
	"github.com/gin-gonic/gin"

	"github.com/serinew/core/internal/app/v1/sign"
	"github.com/serinew/core/internal/app/v1/users"
	"github.com/serinew/core/internal/config"
	"github.com/serinew/core/internal/repository"
	. "github.com/serinew/core/internal/util/http"
)

func V1Routes(r *gin.RouterGroup, store *repository.Store, cfg config.Config) {
	r.GET("/", func(c *gin.Context) {
		Http.OK(c, &SuccOpts{Data: "v1 route"})
	})
	sign.SignRoutes(r.Group("/sign"), store, cfg)
	users.UsersRoutes(r.Group("/users"), store, cfg)
}

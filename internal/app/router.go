package app

import (
	"github.com/gin-gonic/gin"

	v1 "github.com/serinew/core/internal/app/v1"
	"github.com/serinew/core/internal/config"
	"github.com/serinew/core/internal/repository"
	"github.com/serinew/core/internal/service"
	. "github.com/serinew/core/internal/util/http"
)

func AppRoutes(r *gin.Engine, store *repository.Store, cfg config.Config, tokens *service.TokenService) {
	r.Use(v1.SwaggerBarePathRedirect())

	r.GET("/", func(c *gin.Context) {
		Http.OK(c, &SuccOpts{Message: "Core Api service server"})
	})

	v1.V1Routes(r.Group("/v1"), store, cfg, tokens)
}

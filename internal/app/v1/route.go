package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/serinew/core/docs"
	"github.com/serinew/core/internal/app/v1/sign"
	"github.com/serinew/core/internal/app/v1/users"
	"github.com/serinew/core/internal/config"
	"github.com/serinew/core/internal/middleware"
	"github.com/serinew/core/internal/repository"
	"github.com/serinew/core/internal/service"
	_ "github.com/serinew/core/internal/types"
	. "github.com/serinew/core/internal/util/http"
)

// OpenAPI 스펙(basePath /v1)과 동일한 문서 UI 접두입니다.
const swaggerDocsPathPrefix = "/v1/docs"

// SwaggerBarePathRedirect는 `/docs/*any` 만 등록했을 때 매칭되지 않는 `GET /v1/docs`·`/v1/docs/` 를 UI로 보냅니다.
// 엔진 전역 미들웨어여야 합니다(미매칭 요청에는 라우트 그룹 미들웨어가 실행되지 않음).
func SwaggerBarePathRedirect() gin.HandlerFunc {
	prefix := swaggerDocsPathPrefix
	target := prefix + "/index.html"
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}
		p := c.Request.URL.Path
		if p == prefix || p == prefix+"/" {
			c.Redirect(http.StatusFound, target)
			c.Abort()
			return
		}
		c.Next()
	}
}

func V1Routes(r *gin.RouterGroup, store *repository.Store, cfg config.Config, tokens *service.TokenService) {
	mountSwagger(r, store, tokens)
	r.GET("/", getRoot)

	sign.SignRoutes(r.Group("/sign"), store, cfg)
	users.UsersRoutes(r.Group("/users"), store, cfg)
}

// mountSwagger는 로그인(accessToken)·core.admin 행이 있는 사용자만 `/v1/docs/*` 를 볼 수 있습니다.
func mountSwagger(r *gin.RouterGroup, store *repository.Store, tokens *service.TokenService) {
	doc := r.Group("/docs")
	doc.Use(middleware.RequireAccessTokenCookie(tokens))
	doc.Use(middleware.RequireSwaggerDocsAdmin(store))
	doc.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
		ginSwagger.URL(swaggerDocsPathPrefix+"/doc.json"),
		ginSwagger.DefaultModelsExpandDepth(3),
	))
}

// getRoot는 `/v1` 그룹 상태 확인용입니다.
//
// @Summary     V1 상태
// @Description `/v1` 라우트 그룹 확인
// @Tags        meta
// @Produce     json
// @Success     200  {object}  types.SuccessEnvelopeStr
// @Router      / [get]
func getRoot(c *gin.Context) {
	Http.OK(c, &SuccOpts{Data: "v1 route"})
}

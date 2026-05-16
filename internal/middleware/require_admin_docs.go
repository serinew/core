package middleware

import (
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/serinew/core/internal/model/core"
	"github.com/serinew/core/internal/repository"
	httputil "github.com/serinew/core/internal/util/http"
)

// RequireSwaggerDocsAdmin은 [RequireAccessTokenCookie] 다음에 두고,
// 현재 사용자 UUID가 core.admin 에 있을 때만 통과합니다 (Swagger UI 전용).
func RequireSwaggerDocsAdmin(store *repository.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		data, ok := SignLoginDataFromContext(c)
		if !ok || data == nil {
			httputil.Http.Unauthorized(c, &httputil.ErrOpts{Message: "session required"})
			c.Abort()
			return
		}
		var row core.Admin
		err := repository.Repo[core.Admin](store, "core.admin")().
			Query().Where("user_id = ?", data.ID).Take(&row).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				httputil.Http.Forbidden(c, &httputil.ErrOpts{Message: "swagger docs require admin"})
				c.Abort()
				return
			}
			httputil.Http.InternalError(c, &httputil.ErrOpts{Message: err.Error()})
			c.Abort()
			return
		}
		c.Next()
	}
}

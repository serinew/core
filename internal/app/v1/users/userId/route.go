package userId

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/serinew/core/internal/service"
	. "github.com/serinew/core/internal/util/http"
)

// Routes mounts GET /:userId on the caller's group (e.g. /v1/users + /:userId).
func Routes(r *gin.RouterGroup, svc *service.UserService) {
	r.GET("/:userId", func(c *gin.Context) {
		raw := c.Param("userId")
		id, err := uuid.Parse(raw)
		if err != nil {
			Http.BadRequest(c, &ErrOpts{Message: "invalid user id"})
			return
		}
		data, err := svc.GetSignLoginByID(id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				Http.NotFound(c, &ErrOpts{Message: "user not found"})
				return
			}
			Http.InternalError(c, &ErrOpts{Message: err.Error()})
			return
		}
		Http.OK(c, &SuccOpts{Data: data})
	})
}

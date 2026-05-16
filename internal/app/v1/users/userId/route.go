package userId

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/serinew/core/internal/service"
	_ "github.com/serinew/core/internal/types"
	. "github.com/serinew/core/internal/util/http"
)

// Routes mounts GET /:userId on the caller's group (e.g. /v1/users + /:userId).
func Routes(r *gin.RouterGroup, svc *service.UserService) {
	h := &handler{svc: svc}
	r.GET("/:userId", h.GetByID)
}

type handler struct {
	svc *service.UserService
}

// GetByID는 사용자 단건을 주소·프로필과 함께 반환합니다.
//
// @Summary 사용자 단건
// @Tags users
// @Security CookieAuth
// @Produce json
// @Param userId path string true "사용자 UUID"
// @Success 200 {object} types.SuccessEnvelopeSignLogin
// @Failure 400 {object} types.ErrorDoc
// @Failure 401 {object} types.ErrorDoc
// @Failure 404 {object} types.ErrorDoc
// @Failure 500 {object} types.ErrorDoc
// @Router /users/{userId} [get]
func (h *handler) GetByID(c *gin.Context) {
	raw := c.Param("userId")
	id, err := uuid.Parse(raw)
	if err != nil {
		Http.BadRequest(c, &ErrOpts{Message: "invalid user id"})
		return
	}
	data, err := h.svc.GetSignLoginByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			Http.NotFound(c, &ErrOpts{Message: "user not found"})
			return
		}
		Http.InternalError(c, &ErrOpts{Message: err.Error()})
		return
	}
	Http.OK(c, &SuccOpts{Data: data})
}

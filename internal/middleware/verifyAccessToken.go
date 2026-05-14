package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/serinew/core/internal/service"
	"github.com/serinew/core/internal/types"
	httputil "github.com/serinew/core/internal/util/http"
)

const ctxSignLoginData = "signLoginData"

// SignLoginDataFromContext returns session payload set by RequireAccessTokenCookie.
func SignLoginDataFromContext(c *gin.Context) (*types.SignLoginData, bool) {
	v, ok := c.Get(ctxSignLoginData)
	if !ok {
		return nil, false
	}
	data, ok := v.(*types.SignLoginData)
	return data, ok
}

// RequireAccessTokenCookie는 Express의 router.use(auth)와 같이, 이후 핸들러 전에
// accessToken 쿠키를 검증하고 *types.SignLoginData를 컨텍스트에 넣습니다.
func RequireAccessTokenCookie(tokens *service.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, err := c.Cookie(service.AccessTokenCookieName)
		if err != nil && !errors.Is(err, http.ErrNoCookie) {
			httputil.Http.InternalError(c, &httputil.ErrOpts{Message: "cookie read failed"})
			c.Abort()
			return
		}
		raw = strings.TrimSpace(raw)
		if raw == "" {
			httputil.Http.Unauthorized(c, &httputil.ErrOpts{Message: "accessToken cookie missing"})
			c.Abort()
			return
		}
		data, err := tokens.AuthenticateFromToken(c.Request.Context(), raw)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrTokenMissing),
				errors.Is(err, service.ErrTokenInvalid):
				httputil.Http.Unauthorized(c, &httputil.ErrOpts{Message: err.Error()})
			case errors.Is(err, service.ErrTokenMisconfigured):
				httputil.Http.InternalError(c, &httputil.ErrOpts{Message: err.Error()})
			default:
				httputil.Http.InternalError(c, &httputil.ErrOpts{Message: err.Error()})
			}
			c.Abort()
			return
		}
		c.Set(ctxSignLoginData, data)
		c.Next()
	}
}

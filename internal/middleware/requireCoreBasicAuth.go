package middleware

import (
	"crypto/subtle"
	"encoding/base64"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/serinew/core/internal/config"
	httputil "github.com/serinew/core/internal/util/http"
)

// RequireBasicCoreSecret은 CORE_SECRET_KEY가 설정된 경우에만 적용됩니다.
// 클라이언트는 Authorization: Basic <base64> 를 보내야 하며, 디코드 결과는
// 전체가 키이거나 "임의사용자:키" 형식(비밀번호가 키)이어야 합니다.
func RequireBasicCoreSecret(cfg config.Config) gin.HandlerFunc {
	secret := strings.TrimSpace(cfg.CoreSecretKey)
	if secret == "" {
		return func(c *gin.Context) { c.Next() }
	}
	secretBytes := []byte(secret)

	return func(c *gin.Context) {
		scheme, creds, ok := strings.Cut(strings.TrimSpace(c.GetHeader("Authorization")), " ")
		if !ok || scheme == "" || creds == "" {
			httputil.Http.Unauthorized(c, &httputil.ErrOpts{Message: "Authorization header required"})
			c.Abort()
			return
		}
		if strings.ToLower(scheme) != "basic" {
			httputil.Http.Unauthorized(c, &httputil.ErrOpts{Message: "Authorization must be Basic"})
			c.Abort()
			return
		}
		raw := strings.TrimSpace(creds)
		decoded, err := base64.StdEncoding.DecodeString(raw)
		if err != nil {
			httputil.Http.Unauthorized(c, &httputil.ErrOpts{Message: "invalid Basic credentials"})
			c.Abort()
			return
		}
		var candidate string
		if i := strings.IndexByte(string(decoded), ':'); i >= 0 {
			candidate = string(decoded[i+1:])
		} else {
			candidate = string(decoded)
		}
		candidateBytes := []byte(candidate)
		if len(candidateBytes) != len(secretBytes) ||
			subtle.ConstantTimeCompare(candidateBytes, secretBytes) != 1 {
			httputil.Http.Unauthorized(c, &httputil.ErrOpts{Message: "invalid credentials"})
			c.Abort()
			return
		}
		c.Next()
	}
}

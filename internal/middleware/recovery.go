package middleware

import (
	"log"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	httputil "github.com/serinew/core/internal/util/http"
)

// RegisterRecovery replaces the default panic handler with JSON 500 responses.
// Stack traces are logged to the process logger, not sent to the client.
func RegisterRecovery(r *gin.Engine) {
	r.Use(gin.CustomRecovery(func(c *gin.Context, recovered any) {
		if recovered != nil {
			log.Printf("[panic] %v\n%s", recovered, debug.Stack())
		}
		if c.Writer.Written() {
			return
		}
		httputil.Http.InternalError(c, &httputil.ErrOpts{
			Message: "Internal server error.",
		})
		c.Abort()
	}))
}

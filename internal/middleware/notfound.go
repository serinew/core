package middleware

import (
	"github.com/gin-gonic/gin"
	httputil "github.com/serinew/core/internal/util/http"
)

// RegisterNotFound wires Gin's NoRoute handler (unmatched paths → 404 JSON).
func RegisterNotFound(r *gin.Engine) {
	r.NoRoute(func(c *gin.Context) {
		httputil.Http.NotFound(c, &httputil.ErrOpts{
			Message: "The requested route was not found.",
		})
	})
}

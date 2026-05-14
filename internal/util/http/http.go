package util

import (
	"github.com/gin-gonic/gin"
	"github.com/serinew/core/internal/types"
)

// Re-export DTO / option types so handlers can depend on util only.
type (
	SuccessDoc   = types.SuccessDoc
	ErrorDoc     = types.ErrorDoc
	RedirectDoc  = types.RedirectDoc
	SuccOpts     = types.SuccOpts
	ErrOpts      = types.ErrOpts
	RedirectOpts = types.RedirectOpts
)

// HTTP is the namespaced façade for Gin JSON responses (TypeScript Http class style).
type HTTP struct{}

// Http exposes Ok, BadRequest, etc. without importing internal/types in handlers.
var Http HTTP

func (HTTP) Success(c *gin.Context, opts *SuccOpts) { types.Success(c, opts) }
func (HTTP) Error(c *gin.Context, opts *ErrOpts)    { types.Error(c, opts) }
func (HTTP) Redirect(c *gin.Context, opts *RedirectOpts) {
	types.Redirect(c, opts)
}

func (HTTP) OK(c *gin.Context, opts *SuccOpts)             { types.OK(c, opts) }
func (HTTP) Created(c *gin.Context, opts *SuccOpts)        { types.Created(c, opts) }
func (HTTP) Accepted(c *gin.Context, opts *SuccOpts)       { types.Accepted(c, opts) }
func (HTTP) NoContent(c *gin.Context, opts *SuccOpts)      { types.NoContent(c, opts) }
func (HTTP) PartialContent(c *gin.Context, opts *SuccOpts) { types.PartialContent(c, opts) }
func (HTTP) MovedPermanently(c *gin.Context, opts *RedirectOpts) {
	types.MovedPermanently(c, opts)
}
func (HTTP) Found(c *gin.Context, opts *RedirectOpts)       { types.Found(c, opts) }
func (HTTP) SeeOther(c *gin.Context, opts *RedirectOpts)    { types.SeeOther(c, opts) }
func (HTTP) NotModified(c *gin.Context, opts *RedirectOpts) { types.NotModified(c, opts) }
func (HTTP) TemporaryRedirect(c *gin.Context, opts *RedirectOpts) {
	types.TemporaryRedirect(c, opts)
}
func (HTTP) PermanentRedirect(c *gin.Context, opts *RedirectOpts) {
	types.PermanentRedirect(c, opts)
}

func (HTTP) BadRequest(c *gin.Context, opts *ErrOpts) { types.BadRequest(c, opts) }
func (HTTP) Unauthorized(c *gin.Context, opts *ErrOpts) {
	types.Unauthorized(c, opts)
}
func (HTTP) PaymentRequired(c *gin.Context, opts *ErrOpts) {
	types.PaymentRequired(c, opts)
}
func (HTTP) Forbidden(c *gin.Context, opts *ErrOpts) { types.Forbidden(c, opts) }
func (HTTP) NotFound(c *gin.Context, opts *ErrOpts)  { types.NotFound(c, opts) }
func (HTTP) MethodNotAllowed(c *gin.Context, opts *ErrOpts) {
	types.MethodNotAllowed(c, opts)
}
func (HTTP) NotAcceptable(c *gin.Context, opts *ErrOpts) {
	types.NotAcceptable(c, opts)
}
func (HTTP) Conflict(c *gin.Context, opts *ErrOpts) { types.Conflict(c, opts) }
func (HTTP) Gone(c *gin.Context, opts *ErrOpts)     { types.Gone(c, opts) }
func (HTTP) PreconditionFailed(c *gin.Context, opts *ErrOpts) {
	types.PreconditionFailed(c, opts)
}
func (HTTP) PayloadTooLarge(c *gin.Context, opts *ErrOpts) {
	types.PayloadTooLarge(c, opts)
}
func (HTTP) URITooLong(c *gin.Context, opts *ErrOpts) { types.URITooLong(c, opts) }
func (HTTP) UnsupportedMediaType(c *gin.Context, opts *ErrOpts) {
	types.UnsupportedMediaType(c, opts)
}
func (HTTP) UnprocessableEntity(c *gin.Context, opts *ErrOpts) {
	types.UnprocessableEntity(c, opts)
}
func (HTTP) TooManyRequests(c *gin.Context, opts *ErrOpts) {
	types.TooManyRequests(c, opts)
}
func (HTTP) InternalError(c *gin.Context, opts *ErrOpts) {
	types.InternalError(c, opts)
}
func (HTTP) NotImplemented(c *gin.Context, opts *ErrOpts) {
	types.NotImplemented(c, opts)
}
func (HTTP) BadGateway(c *gin.Context, opts *ErrOpts) { types.BadGateway(c, opts) }
func (HTTP) ServiceUnavailable(c *gin.Context, opts *ErrOpts) {
	types.ServiceUnavailable(c, opts)
}
func (HTTP) GatewayTimeout(c *gin.Context, opts *ErrOpts) {
	types.GatewayTimeout(c, opts)
}

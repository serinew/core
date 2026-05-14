package types

import (
	stdhttp "net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type SuccessDoc struct {
	Status     string `json:"status"`
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	Count      *int   `json:"count,omitempty"`
	Data       any    `json:"data,omitempty"`
}

type ErrorDoc struct {
	Status     string `json:"status"`
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

type RedirectDoc struct {
	Status     string `json:"status"`
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	Data       any    `json:"data,omitempty"`
}

type SuccOpts struct {
	Data       any
	Message    string
	StatusCode int // 0 = default for that helper
	Count      *int
}

type ErrOpts struct {
	Message    string
	StatusCode int // 0 = helper default or 400
}

type RedirectOpts struct {
	Data       any
	Message    string
	StatusCode int
}

func buildSuccess(opts *SuccOpts, defCode int, defMsg string) SuccessDoc {
	var o SuccOpts
	if opts != nil {
		o = *opts
	}
	code := o.StatusCode
	if code == 0 {
		code = defCode
	}
	msg := strings.TrimSpace(o.Message)
	if msg == "" {
		msg = defMsg
	}
	return SuccessDoc{
		Status:     "success",
		StatusCode: code,
		Message:    msg,
		Count:      o.Count,
		Data:       o.Data,
	}
}

func Success(c *gin.Context, opts *SuccOpts) {
	doc := buildSuccess(opts, stdhttp.StatusOK, "Success")
	c.JSON(doc.StatusCode, doc)
}

func Error(c *gin.Context, opts *ErrOpts) {
	var o ErrOpts
	if opts != nil {
		o = *opts
	}
	code := o.StatusCode
	if code == 0 {
		code = stdhttp.StatusBadRequest
	}
	msg := strings.TrimSpace(o.Message)
	if msg == "" {
		msg = "Error"
	}
	c.JSON(code, ErrorDoc{
		Status:     "error",
		StatusCode: code,
		Message:    msg,
	})
}

func Redirect(c *gin.Context, opts *RedirectOpts) {
	var o RedirectOpts
	if opts != nil {
		o = *opts
	}
	code := o.StatusCode
	if code == 0 {
		code = stdhttp.StatusFound
	}
	msg := strings.TrimSpace(o.Message)
	if msg == "" {
		msg = "Redirect"
	}
	c.JSON(code, RedirectDoc{
		Status:     "redirect",
		StatusCode: code,
		Message:    msg,
		Data:       o.Data,
	})
}

// OK shorthand 200 / "OK"
func OK(c *gin.Context, opts *SuccOpts) {
	doc := buildSuccess(opts, stdhttp.StatusOK, "OK")
	c.JSON(doc.StatusCode, doc)
}

func Created(c *gin.Context, opts *SuccOpts) {
	doc := buildSuccess(opts, stdhttp.StatusCreated, "Created")
	c.JSON(doc.StatusCode, doc)
}

func Accepted(c *gin.Context, opts *SuccOpts) {
	doc := buildSuccess(opts, stdhttp.StatusAccepted, "Accepted")
	c.JSON(doc.StatusCode, doc)
}

func NoContent(c *gin.Context, opts *SuccOpts) {
	doc := buildSuccess(withEmpty(opts), stdhttp.StatusNoContent, "No Content")
	c.JSON(doc.StatusCode, doc)
}

func PartialContent(c *gin.Context, opts *SuccOpts) {
	doc := buildSuccess(opts, stdhttp.StatusPartialContent, "Partial Content")
	c.JSON(doc.StatusCode, doc)
}

func withEmpty(o *SuccOpts) *SuccOpts {
	if o == nil {
		return &SuccOpts{Data: nil}
	}
	copy := *o
	copy.Data = nil
	return &copy
}

func redirPrep(opts *RedirectOpts, code int, msg string) *RedirectOpts {
	if opts == nil {
		opts = &RedirectOpts{}
	}
	if opts.StatusCode == 0 {
		opts.StatusCode = code
	}
	if strings.TrimSpace(opts.Message) == "" {
		opts.Message = msg
	}
	return opts
}

func MovedPermanently(c *gin.Context, opts *RedirectOpts) {
	Redirect(c, redirPrep(opts, stdhttp.StatusMovedPermanently, "Moved Permanently"))
}

func Found(c *gin.Context, opts *RedirectOpts) {
	Redirect(c, redirPrep(opts, stdhttp.StatusFound, "Found"))
}

func SeeOther(c *gin.Context, opts *RedirectOpts) {
	Redirect(c, redirPrep(opts, stdhttp.StatusSeeOther, "See Other"))
}

func NotModified(c *gin.Context, opts *RedirectOpts) {
	Redirect(c, redirPrep(opts, stdhttp.StatusNotModified, "Not Modified"))
}

func TemporaryRedirect(c *gin.Context, opts *RedirectOpts) {
	Redirect(c, redirPrep(opts, stdhttp.StatusTemporaryRedirect, "Temporary Redirect"))
}

func PermanentRedirect(c *gin.Context, opts *RedirectOpts) {
	Redirect(c, redirPrep(opts, stdhttp.StatusPermanentRedirect, "Permanent Redirect"))
}

func errPrep(opts *ErrOpts, code int, msg string) *ErrOpts {
	if opts == nil {
		opts = &ErrOpts{}
	}
	if opts.StatusCode == 0 {
		opts.StatusCode = code
	}
	if strings.TrimSpace(opts.Message) == "" {
		opts.Message = msg
	}
	return opts
}

func BadRequest(c *gin.Context, opts *ErrOpts) {
	Error(c, errPrep(opts, stdhttp.StatusBadRequest, "Bad Request"))
}

func Unauthorized(c *gin.Context, opts *ErrOpts) {
	Error(c, errPrep(opts, stdhttp.StatusUnauthorized, "Unauthorized"))
}

func PaymentRequired(c *gin.Context, opts *ErrOpts) {
	Error(c, errPrep(opts, stdhttp.StatusPaymentRequired, "Payment Required"))
}

func Forbidden(c *gin.Context, opts *ErrOpts) {
	Error(c, errPrep(opts, stdhttp.StatusForbidden, "Forbidden"))
}

func NotFound(c *gin.Context, opts *ErrOpts) {
	Error(c, errPrep(opts, stdhttp.StatusNotFound, "Not Found"))
}

func MethodNotAllowed(c *gin.Context, opts *ErrOpts) {
	Error(c, errPrep(opts, stdhttp.StatusMethodNotAllowed, "Method Not Allowed"))
}

func NotAcceptable(c *gin.Context, opts *ErrOpts) {
	Error(c, errPrep(opts, stdhttp.StatusNotAcceptable, "Not Acceptable"))
}

func Conflict(c *gin.Context, opts *ErrOpts) {
	Error(c, errPrep(opts, stdhttp.StatusConflict, "Conflict"))
}

func Gone(c *gin.Context, opts *ErrOpts) {
	Error(c, errPrep(opts, stdhttp.StatusGone, "Gone"))
}

func PreconditionFailed(c *gin.Context, opts *ErrOpts) {
	Error(c, errPrep(opts, stdhttp.StatusPreconditionFailed, "Precondition Failed"))
}

func PayloadTooLarge(c *gin.Context, opts *ErrOpts) {
	Error(c, errPrep(opts, stdhttp.StatusRequestEntityTooLarge, "Payload Too Large"))
}

func URITooLong(c *gin.Context, opts *ErrOpts) {
	Error(c, errPrep(opts, stdhttp.StatusRequestURITooLong, "URI Too Long"))
}

func UnsupportedMediaType(c *gin.Context, opts *ErrOpts) {
	Error(c, errPrep(opts, stdhttp.StatusUnsupportedMediaType, "Unsupported Media Type"))
}

func UnprocessableEntity(c *gin.Context, opts *ErrOpts) {
	Error(c, errPrep(opts, stdhttp.StatusUnprocessableEntity, "Unprocessable Entity"))
}

func TooManyRequests(c *gin.Context, opts *ErrOpts) {
	Error(c, errPrep(opts, stdhttp.StatusTooManyRequests, "Too Many Requests"))
}

func InternalError(c *gin.Context, opts *ErrOpts) {
	Error(c, errPrep(opts, stdhttp.StatusInternalServerError, "Internal Server Error"))
}

func NotImplemented(c *gin.Context, opts *ErrOpts) {
	Error(c, errPrep(opts, stdhttp.StatusNotImplemented, "Not Implemented"))
}

func BadGateway(c *gin.Context, opts *ErrOpts) {
	Error(c, errPrep(opts, stdhttp.StatusBadGateway, "Bad Gateway"))
}

func ServiceUnavailable(c *gin.Context, opts *ErrOpts) {
	Error(c, errPrep(opts, stdhttp.StatusServiceUnavailable, "Service Unavailable"))
}

func GatewayTimeout(c *gin.Context, opts *ErrOpts) {
	Error(c, errPrep(opts, stdhttp.StatusGatewayTimeout, "Gateway Timeout"))
}

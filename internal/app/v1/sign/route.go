package sign

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/serinew/core/internal/config"
	"github.com/serinew/core/internal/repository"
	"github.com/serinew/core/internal/service"
	"github.com/serinew/core/internal/types"
	. "github.com/serinew/core/internal/util/http"
)

// SignRoutes: POST …/sign/* + GET /v1/sign (쿠키 accessToken).
func SignRoutes(r *gin.RouterGroup, store *repository.Store, cfg config.Config) {
	h := &signHandlers{
		svc:    service.NewSignService(store),
		tokens: service.NewTokenService(store, cfg.TokenSecret),
		cfg:    cfg,
	}
	r.POST("/register", h.Register)
	r.POST("/login", h.Login)
	r.POST("/password", h.ChangePassword)
	r.GET("", h.Session)
	r.GET("/", h.Session)
}

type signHandlers struct {
	svc    *service.SignService
	tokens *service.TokenService
	cfg    config.Config
}

// Register는 회원 가입을 처리합니다.
//
// @Summary 회원 가입
// @Tags sign
// @Accept json
// @Produce json
// @Param body body types.SignUpReq true "요청 바디"
// @Success 201 {object} types.SuccessEnvelopeUser
// @Failure 400 {object} types.ErrorDoc
// @Failure 409 {object} types.ErrorDoc
// @Router /sign/register [post]
func (h *signHandlers) Register(c *gin.Context) {
	var body types.SignUpReq
	if !types.BindJSON(c, &body) {
		return
	}
	u, err := h.svc.SignUp(service.SignUpInput{
		Name:     body.Name,
		Email:    body.Email,
		Phone:    body.Phone,
		Username: body.Username,
		Password: body.Password,
		Addr1:    body.Addr1,
		Addr2:    body.Addr2,
		ZipCode:  body.ZipCode,
		Lat:      body.Lat,
		Lng:      body.Lng,
	})
	if err != nil {
		writeSignErr(c, err)
		return
	}
	Http.Created(c, &SuccOpts{Data: u})
}

// Login은 로그인 후 accessToken HttpOnly 쿠키를 설정합니다.
//
// @Summary 로그인
// @Tags sign
// @Accept json
// @Produce json
// @Param body body types.SignLoginReq true "요청 바디"
// @Success 200 {object} types.SuccessEnvelopeSignLogin
// @Failure 401 {object} types.ErrorDoc
// @Failure 500 {object} types.ErrorDoc
// @Router /sign/login [post]
func (h *signHandlers) Login(c *gin.Context) {
	var body types.SignLoginReq
	if !types.BindJSON(c, &body) {
		return
	}
	data, err := h.svc.Login(body.Username, body.Password)
	if err != nil {
		writeSignErr(c, err)
		return
	}
	prevTok := readLoginAccessTokenCookie(c)
	rawTok, _, err := h.tokens.IssueOrRenewAccessToken(c.Request.Context(), prevTok, data.ID)
	if err != nil {
		writeSignErr(c, err)
		return
	}
	c.SetCookie(
		service.AccessTokenCookieName,
		rawTok,
		service.AccessTokenCookieMaxAgeSeconds,
		"/",
		"",
		h.cfg.CookieSecure,
		true,
	)
	if sess, aerr := h.tokens.AuthenticateFromToken(c.Request.Context(), rawTok); aerr == nil && sess != nil {
		data = sess
	}
	Http.OK(c, &SuccOpts{Data: data})
}

// readLoginAccessTokenCookie는 로그인 재발급·연장 판별용입니다. 없으면 빈 문자열.
func readLoginAccessTokenCookie(c *gin.Context) string {
	raw, err := c.Cookie(service.AccessTokenCookieName)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(raw)
}

// ChangePassword는 비밀번호 변경을 처리합니다.
//
// @Summary 비밀번호 변경
// @Tags sign
// @Accept json
// @Produce json
// @Param body body types.SignChangePasswordReq true "요청 바디"
// @Success 200 {object} types.SuccessEnvelopeMsg
// @Failure 400 {object} types.ErrorDoc
// @Failure 401 {object} types.ErrorDoc
// @Failure 404 {object} types.ErrorDoc
// @Router /sign/password [post]
func (h *signHandlers) ChangePassword(c *gin.Context) {
	var body types.SignChangePasswordReq
	if !types.BindJSON(c, &body) {
		return
	}
	id, err := uuid.Parse(body.UserID)
	if err != nil {
		Http.BadRequest(c, &ErrOpts{Message: "invalid userId"})
		return
	}
	err = h.svc.ChangePassword(id, body.OldPassword, body.NewPassword)
	if err != nil {
		writeSignErr(c, err)
		return
	}
	Http.OK(c, &SuccOpts{Message: "password updated"})
}

// Session은 accessToken 쿠키로 현재 세션 페이로드를 반환합니다.
//
// @Summary 현재 세션
// @Tags sign
// @Produce json
// @Success 200 {object} types.SuccessEnvelopeSignLogin
// @Failure 401 {object} types.ErrorDoc
// @Failure 500 {object} types.ErrorDoc
// @Router /sign [get]
func (h *signHandlers) Session(c *gin.Context) {
	raw, err := c.Cookie(service.AccessTokenCookieName)
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		Http.InternalError(c, &ErrOpts{Message: err.Error()})
		return
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		Http.Unauthorized(c, &ErrOpts{Message: "accessToken cookie missing"})
		return
	}
	data, err := h.tokens.AuthenticateFromToken(c.Request.Context(), raw)
	if err != nil {
		writeSignErr(c, err)
		return
	}
	Http.OK(c, &SuccOpts{Data: data})
}

func writeSignErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrSignDuplicateUsername),
		errors.Is(err, service.ErrSignDuplicateEmail):
		Http.Conflict(c, &ErrOpts{Message: err.Error()})
	case errors.Is(err, service.ErrSignInvalidLogin),
		errors.Is(err, service.ErrTokenMissing),
		errors.Is(err, service.ErrTokenInvalid):
		Http.Unauthorized(c, &ErrOpts{Message: err.Error()})
	case errors.Is(err, service.ErrTokenMisconfigured):
		Http.InternalError(c, &ErrOpts{Message: err.Error()})
	case errors.Is(err, service.ErrSignWeakPassword),
		errors.Is(err, service.ErrSignWrongPassword):
		Http.BadRequest(c, &ErrOpts{Message: err.Error()})
	case errors.Is(err, gorm.ErrRecordNotFound):
		Http.NotFound(c, &ErrOpts{Message: "resource not found"})
	default:
		if strings.HasPrefix(err.Error(), "sign: ") {
			Http.BadRequest(c, &ErrOpts{Message: err.Error()})
			return
		}
		Http.InternalError(c, &ErrOpts{Message: err.Error()})
	}
}

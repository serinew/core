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
	svc := service.NewSignService(store)
	tokens := service.NewTokenService(store, cfg.TokenSecret)

	r.POST("/register", func(c *gin.Context) {
		var body types.SignUpReq
		if !types.BindJSON(c, &body) {
			return
		}
		u, err := svc.SignUp(service.SignUpInput{
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
	})

	r.POST("/login", func(c *gin.Context) {
		var body types.SignLoginReq
		if !types.BindJSON(c, &body) {
			return
		}
		data, err := svc.Login(body.Username, body.Password)
		if err != nil {
			writeSignErr(c, err)
			return
		}
		rawTok, _, err := tokens.IssueAccessToken(c.Request.Context(), data.ID)
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
			cfg.CookieSecure,
			true,
		)
		Http.OK(c, &SuccOpts{Data: data})
	})

	r.POST("/password", func(c *gin.Context) {
		var body types.SignChangePasswordReq
		if !types.BindJSON(c, &body) {
			return
		}
		id, err := uuid.Parse(body.UserID)
		if err != nil {
			Http.BadRequest(c, &ErrOpts{Message: "invalid userId"})
			return
		}
		err = svc.ChangePassword(id, body.OldPassword, body.NewPassword)
		if err != nil {
			writeSignErr(c, err)
			return
		}
		Http.OK(c, &SuccOpts{Message: "password updated"})
	})

	sessionGET := func(c *gin.Context) {
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
		data, err := tokens.AuthenticateFromToken(c.Request.Context(), raw)
		if err != nil {
			writeSignErr(c, err)
			return
		}
		Http.OK(c, &SuccOpts{Data: data})
	}
	r.GET("", sessionGET)
	r.GET("/", sessionGET)
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

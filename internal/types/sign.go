package types

import "github.com/serinew/core/internal/model/core"

// Sign API 요청 바디 (JSON camelCase).

// SignUpReq corresponds to POST …/sign/register.
type SignUpReq struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`

	Addr1   string   `json:"addr1" binding:"required,max=256"`
	Addr2   *string  `json:"addr2,omitempty"`
	ZipCode int16    `json:"zipCode" binding:"required"`
	Lat     *float64 `json:"lat,omitempty"`
	Lng     *float64 `json:"lng,omitempty"`
}

// SignLoginReq corresponds to POST …/sign/login.
type SignLoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// SignChangePasswordReq corresponds to POST …/sign/password.
type SignChangePasswordReq struct {
	UserID      string `json:"userId" binding:"required"`
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

// SignLoginAdmin는 관리자 행이 있을 때만 응답에 포함됩니다 (omitempty).
type SignLoginAdmin struct {
	Authority string `json:"authority"`
}

// SignLoginData는 로그인 성공 시 data 객체(JSON)입니다. 사용자 필드는 최상위로 펼칩니다 (json inline).
type SignLoginData struct {
	core.User `json:",inline"`
	Address   []core.UserAddress `json:"address"`
	Profile   *core.UserProfile  `json:"profile"`
	Admin     *SignLoginAdmin    `json:"admin,omitempty"`
}

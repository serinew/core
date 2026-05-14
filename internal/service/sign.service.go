package service

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/serinew/core/internal/model/core"
	"github.com/serinew/core/internal/repository"
	"github.com/serinew/core/internal/types"
)

const bcryptCost = bcrypt.DefaultCost
const minPasswordLen = 8

var (
	ErrSignDuplicateUsername = errors.New("sign: duplicate username")
	ErrSignDuplicateEmail    = errors.New("sign: duplicate email")
	ErrSignInvalidLogin      = errors.New("sign: invalid username or password")
	ErrSignWeakPassword      = errors.New("sign: password too short")
	ErrSignWrongPassword     = errors.New("sign: current password mismatch")
)

// SignUpInput는 회원가입 시 core.user / core.user_sign / 주소·프로필 행 초기값에 넣을 값입니다.
type SignUpInput struct {
	Name     string
	Email    string
	Phone    string
	Username string
	Password string

	Addr1   string
	Addr2   *string
	ZipCode int16
	Lat     *float64
	Lng     *float64
}

// SignService는 로그인·가입·비밀번호 변경을 담당합니다.
type SignService struct {
	store     *repository.Store
	users     func() *repository.RepoConfig[core.User]
	signs     func() *repository.RepoConfig[core.UserSign]
	addresses func() *repository.RepoConfig[core.UserAddress]
	profiles  func() *repository.RepoConfig[core.UserProfile]
}

func NewSignService(store *repository.Store) *SignService {
	return &SignService{
		store:     store,
		users:     repository.Repo[core.User](store, "core.user"),
		signs:     repository.Repo[core.UserSign](store, "core.user_sign"),
		addresses: repository.Repo[core.UserAddress](store, "core.user_address"),
		profiles:  repository.Repo[core.UserProfile](store, "core.user_profile"),
	}
}

// SignUp은 사용자·로그인·주소(core.user_address)·빈 프로필(core.user_profile)을 트랜잭션으로 생성합니다.
func (s *SignService) SignUp(in SignUpInput) (*core.User, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Email = strings.TrimSpace(in.Email)
	in.Phone = strings.TrimSpace(in.Phone)
	in.Username = strings.TrimSpace(in.Username)
	in.Addr1 = strings.TrimSpace(in.Addr1)
	if in.Addr2 != nil {
		v := strings.TrimSpace(*in.Addr2)
		if v == "" {
			in.Addr2 = nil
		} else {
			in.Addr2 = &v
		}
	}
	if err := validateSignUpFields(in); err != nil {
		return nil, err
	}
	if err := s.ensureUsernameAvailable(in.Username); err != nil {
		return nil, err
	}
	if err := s.ensureEmailAvailable(in.Email); err != nil {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcryptCost)
	if err != nil {
		return nil, err
	}

	var created *core.User
	err = s.store.DB.Transaction(func(tx *gorm.DB) error {
		u := core.User{
			Name:  in.Name,
			Email: in.Email,
			Phone: in.Phone,
		}
		if err := tx.Model(&core.User{}).Create(&u).Error; err != nil {
			return err
		}
		sign := core.UserSign{
			UserID:       u.ID,
			Username:     in.Username,
			PasswordHash: string(hash),
		}
		if err := tx.Model(&core.UserSign{}).Create(&sign).Error; err != nil {
			return err
		}
		addr := core.UserAddress{
			UserID:  u.ID,
			Addr1:   in.Addr1,
			Addr2:   in.Addr2,
			ZipCode: in.ZipCode,
			Lat:     in.Lat,
			Lng:     in.Lng,
		}
		if err := tx.Model(&core.UserAddress{}).Create(&addr).Error; err != nil {
			return err
		}
		profile := core.UserProfile{UserID: u.ID}
		if err := tx.Model(&core.UserProfile{}).Create(&profile).Error; err != nil {
			return err
		}
		created = &u
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

// Login은 username·password를 검증한 뒤 user·주소 목록·프로필을 묶어 반환합니다.
func (s *SignService) Login(username, password string) (*types.SignLoginData, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return nil, ErrSignInvalidLogin
	}
	sign, err := s.findSignByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSignInvalidLogin
		}
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(sign.PasswordHash), []byte(password)); err != nil {
		return nil, ErrSignInvalidLogin
	}
	return LoadSignLoginData(s.store, sign.UserID)
}

// ChangePassword는 로그인한 사용자의 비밀번호를 교체합니다.
func (s *SignService) ChangePassword(userID uuid.UUID, oldPassword, newPassword string) error {
	if err := validateNewPassword(newPassword); err != nil {
		return err
	}
	var sign core.UserSign
	if err := s.signs().Query().Where("user_id = ?", userID).Take(&sign).Error; err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(sign.PasswordHash), []byte(oldPassword)); err != nil {
		return ErrSignWrongPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcryptCost)
	if err != nil {
		return err
	}
	now := time.Now()
	return s.signs().Query().Where("user_id = ?", userID).Updates(map[string]any{
		"password_hash": string(hash),
		"updated_at":    &now,
	}).Error
}

func validateSignUpFields(in SignUpInput) error {
	if in.Name == "" || in.Email == "" || in.Phone == "" || in.Username == "" || in.Addr1 == "" {
		return errors.New("sign: required field missing")
	}
	if len(in.Name) > 256 || len(in.Email) > 64 || len(in.Phone) > 24 || len(in.Username) > 24 || len(in.Addr1) > 256 {
		return errors.New("sign: field length exceeds schema")
	}
	if in.Addr2 != nil && len(*in.Addr2) > 256 {
		return errors.New("sign: field length exceeds schema")
	}
	return validateNewPassword(in.Password)
}

func validateNewPassword(pw string) error {
	if len(pw) < minPasswordLen {
		return ErrSignWeakPassword
	}
	return nil
}

func (s *SignService) ensureUsernameAvailable(username string) error {
	var sign core.UserSign
	err := s.signs().Query().Where("username = ?", username).Take(&sign).Error
	if err == nil {
		return ErrSignDuplicateUsername
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}

func (s *SignService) ensureEmailAvailable(email string) error {
	var u core.User
	err := s.users().Query().Where("email = ?", email).Take(&u).Error
	if err == nil {
		return ErrSignDuplicateEmail
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}

func (s *SignService) findSignByUsername(username string) (*core.UserSign, error) {
	var sign core.UserSign
	if err := s.signs().Query().Where("username = ?", username).Take(&sign).Error; err != nil {
		return nil, err
	}
	return &sign, nil
}

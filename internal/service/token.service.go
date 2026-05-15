package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/serinew/core/internal/model/sso"
	"github.com/serinew/core/internal/repository"
	"github.com/serinew/core/internal/types"
)

const (
	// AccessTokenCookieName 로그인 쿠키 이름 (accessToken=value).
	AccessTokenCookieName = "accessToken"
	// AccessTokenCookieMaxAgeSeconds Set-Cookie Max-Age (48h).
	AccessTokenCookieMaxAgeSeconds = 48 * 3600
	redisTokenKeyPrefix   = "users:token:"
	accessTokenAuthWindow = 24 * time.Hour
	// tokenCacheTTL Redis 세션 행 최대 보존(키 TTL 상한); 실제 TTL은 만료까지의 남은 시간으로 잡음.
	tokenCacheTTL = 48 * time.Hour
	// MinTokenSecretLength 미만이면 발급 실패 (.env TOKEN_SECRET).
	MinTokenSecretLength = 16
)

var (
	ErrTokenMissing       = errors.New("token: access token cookie missing")
	ErrTokenInvalid       = errors.New("token: invalid or expired")
	ErrTokenMisconfigured = errors.New("token: TOKEN_SECRET unset or too short")
)

// tokenRedisEnvelope — Redis 하나의 키에만 저장. API에는 data(SignLoginData)만 노출하면 됩니다.
type tokenRedisEnvelope struct {
	ExpiresAt time.Time           `json:"expiresAt"`
	Data      types.SignLoginData `json:"data"`
}

// TokenService는 브라우저 accessToken(SSO 행·Redis) 검증 및 발급을 담당합니다.
type TokenService struct {
	store    *repository.Store
	secret   string
	ssoRepos func() *repository.RepoConfig[sso.AccessToken]
}

func NewTokenService(store *repository.Store, tokenSecret string) *TokenService {
	return &TokenService{
		store:    store,
		secret:   tokenSecret,
		ssoRepos: repository.Repo[sso.AccessToken](store, "sso.access_token"),
	}
}

// IssueAccessToken은 DB·Redis에 토큰을 기록하고, Set-Cookie에 넣을 raw 값을 반환합니다.
func (t *TokenService) IssueAccessToken(ctx context.Context, userID uuid.UUID) (token string, authExpires time.Time, err error) {
	if len(t.secret) < MinTokenSecretLength {
		return "", time.Time{}, ErrTokenMisconfigured
	}
	token, err = deriveOpaqueAccessToken(t.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	now := time.Now().UTC()
	authExpires = now.Add(accessTokenAuthWindow)

	row := &sso.AccessToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: authExpires,
	}
	if err := t.ssoRepos().Create(row); err != nil {
		return "", time.Time{}, err
	}

	data, loadErr := LoadSignLoginData(t.store, userID)
	if loadErr != nil {
		return "", time.Time{}, loadErr
	}
	if err := t.writeRedis(ctx, token, data, authExpires); err != nil {
		log.Printf("token: redis cache write failed (%v), proceeding with DB-only auth", err)
	}
	return token, authExpires, nil
}

func (t *TokenService) writeRedis(ctx context.Context, token string, data *types.SignLoginData, authExpires time.Time) error {
	if t.store == nil || t.store.Redis == nil {
		return nil
	}
	env := tokenRedisEnvelope{
		ExpiresAt: authExpires,
		Data:      *data,
	}
	b, err := json.Marshal(env)
	if err != nil {
		return err
	}
	return t.store.Redis.Set(ctx, redisTokenKey(token), b, redisSessionTTL(authExpires)).Err()
}

func (t *TokenService) AuthenticateFromToken(ctx context.Context, rawToken string) (*types.SignLoginData, error) {
	rawToken = sanitizeToken(rawToken)
	if rawToken == "" {
		return nil, ErrTokenMissing
	}
	cached, ok, rErr := t.tryRedis(ctx, rawToken)
	if rErr != nil {
		log.Printf("token: redis read failed (%v), falling back to PostgreSQL", rErr)
	}
	if ok && cached != nil {
		return cached, nil
	}
	return t.tryDB(ctx, rawToken)
}

func sanitizeToken(s string) string {
	// 쿠키값 앞뒤 공백 제거 (일부 클라이언트용)
	i, j := 0, len(s)
	for i < j && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	for j > i && (s[j-1] == ' ' || s[j-1] == '\t') {
		j--
	}
	return s[i:j]
}

func (t *TokenService) tryRedis(ctx context.Context, token string) (*types.SignLoginData, bool, error) {
	if t.store.Redis == nil {
		return nil, false, nil
	}
	b, err := t.store.Redis.Get(ctx, redisTokenKey(token)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false, nil
		}
		return nil, false, err
	}
	exp, session, decErr := decodeRedisTokenBlob(b)
	if decErr != nil {
		return nil, false, fmt.Errorf("token redis json: %w", decErr)
	}
	if time.Now().UTC().After(exp) {
		t.purgeRedisToken(ctx, token)
		return nil, false, ErrTokenInvalid
	}
	return session, true, nil
}

// decodeRedisTokenBlob: v2 { expiresAt, data }, v1(레거시) flat+inline.
func decodeRedisTokenBlob(b []byte) (exp time.Time, data *types.SignLoginData, err error) {
	var env tokenRedisEnvelope
	if uerr := json.Unmarshal(b, &env); uerr == nil && env.Data.ID != uuid.Nil {
		d := env.Data
		return env.ExpiresAt, &d, nil
	}
	var leg struct {
		types.SignLoginData `json:",inline"`
		ExpiresAt           time.Time `json:"expiresAt"`
	}
	if err := json.Unmarshal(b, &leg); err != nil {
		return time.Time{}, nil, err
	}
	if leg.SignLoginData.ID == uuid.Nil {
		return time.Time{}, nil, fmt.Errorf("token redis: unrecognized or empty payload")
	}
	d := leg.SignLoginData
	return leg.ExpiresAt, &d, nil
}

func (t *TokenService) tryDB(ctx context.Context, rawToken string) (*types.SignLoginData, error) {
	var row sso.AccessToken
	err := t.ssoRepos().Query().Where("token = ?", rawToken).Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTokenInvalid
		}
		return nil, err
	}
	if time.Now().UTC().After(row.ExpiresAt) {
		t.purgeRedisToken(ctx, rawToken)
		return nil, ErrTokenInvalid
	}
	session, err := LoadSignLoginData(t.store, row.UserID)
	if err != nil {
		return nil, err
	}
	if err := t.writeRedis(ctx, rawToken, session, row.ExpiresAt); err != nil {
		log.Printf("token: redis refill failed (%v)", err)
	}
	return session, nil
}

// RevokeAccessToken은 쿠키에 실린 raw 토큰과 동일 문자열 기준으로 Redis 캐시·sso.access_token 행을 제거합니다.
func (t *TokenService) RevokeAccessToken(ctx context.Context, raw string) error {
	raw = sanitizeToken(raw)
	if raw == "" {
		return nil
	}
	if t.store != nil && t.store.Redis != nil {
		if err := t.store.Redis.Del(ctx, redisTokenKey(raw)).Err(); err != nil {
			log.Printf("token: revoke redis del failed (%v)", err)
		}
	}
	var zero sso.AccessToken
	return t.ssoRepos().Query().Where("token = ?", raw).Delete(&zero).Error
}

func deriveOpaqueAccessToken(secret string) (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(id[:])
	sig := hex.EncodeToString(mac.Sum(nil))
	// 형식: {uuid-v7}.{hmac 앞 32자} — 무작위 길이·쿠키 전송 가능
	return fmt.Sprintf("%s.%s", id.String(), sig[:32]), nil
}

func redisTokenKey(token string) string {
	return redisTokenKeyPrefix + token
}

// redisSessionTTL: Redis TTL until auth expiry, capped by tokenCacheTTL.
func redisSessionTTL(authExpires time.Time) time.Duration {
	now := time.Now().UTC()
	d := authExpires.Sub(now)
	if d < time.Second {
		return time.Second
	}
	if d > tokenCacheTTL {
		return tokenCacheTTL
	}
	return d
}

func (t *TokenService) purgeRedisToken(ctx context.Context, raw string) {
	raw = sanitizeToken(raw)
	if raw == "" || t.store == nil || t.store.Redis == nil {
		return
	}
	_ = t.store.Redis.Del(ctx, redisTokenKey(raw)).Err()
}

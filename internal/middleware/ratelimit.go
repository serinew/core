package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/serinew/core/internal/config"
	"github.com/serinew/core/internal/service"
	httputil "github.com/serinew/core/internal/util/http"
)

// RouteBurstLimiter는 동일 라우트(템플릿 경로)에 대한 반복 호출 속도를 제한합니다.
// 설정 시 429 연속 발생·임계 초과 분에는 accessToken 쿠키·Redis·sso.access_token을 무효화합니다.
func RouteBurstLimiter(cfg config.Config, rdb *redis.Client, tokens *service.TokenService) gin.HandlerFunc {
	if !cfg.RateLimitEnabled {
		return func(c *gin.Context) { c.Next() }
	}
	max := int64(cfg.RateLimitMax)
	if max < 1 {
		max = 1
	}
	window := cfg.RateLimitWindow
	if window <= 0 {
		window = time.Minute
	}

	mem := newMemoryLimiter(window)
	strikeMem := newMemStrike(window)
	strikeThreshold := cfg.RateLimitRevokeAfterStrikes

	return func(c *gin.Context) {
		keyMaterial := burstKeyMaterial(c)

		var allowed bool
		var err error
		ctx := c.Request.Context()
		if rdb != nil {
			allowed, err = allowRedisBurst(ctx, rdb, keyMaterial, max, window)
			if err != nil {
				log.Printf("middleware ratelimit: redis error (%v), using in-memory fallback", err)
				allowed = mem.allow(keyMaterial, int(max))
			}
		} else {
			allowed = mem.allow(keyMaterial, int(max))
		}

		if !allowed {
			var revoked bool
			if strikeThreshold >= 1 && tokens != nil {
				if raw, terr := c.Cookie(service.AccessTokenCookieName); terr == nil && raw != "" {
					sMat := strikeKeyMaterial(c, raw)
					var n int64
					if rdb != nil {
						n, err = redisIncrStrike(ctx, rdb, sMat, window)
						if err != nil {
							log.Printf("middleware ratelimit: strike redis (%v), mem fallback", err)
							n = int64(strikeMem.bump(sMat))
						}
					} else {
						n = int64(strikeMem.bump(sMat))
					}
					if n >= int64(strikeThreshold) {
						if rerr := tokens.RevokeAccessToken(ctx, raw); rerr != nil {
							log.Printf("middleware ratelimit: revoke token failed (%v)", rerr)
						} else {
							clearAccessTokenCookie(c, cfg)
							revoked = true
							if rdb != nil {
								_ = rdb.Del(ctx, redisStrikeKey(sMat)).Err()
							} else {
								strikeMem.clear(sMat)
							}
						}
					}
				}
			}

			winSec := int(window.Round(time.Second).Seconds())
			msg := fmt.Sprintf(
				"Request limit exceeded. The same path allows up to %d requests within %ds.",
				max, winSec,
			)
			if revoked {
				msg += " Your session cookie has been cleared due to repeated abuse."
			}
			httputil.Http.TooManyRequests(c, &httputil.ErrOpts{Message: msg})
			c.Abort()
			return
		}
		c.Next()
	}
}

func clearAccessTokenCookie(c *gin.Context, cfg config.Config) {
	c.SetCookie(service.AccessTokenCookieName, "", -1, "/", "", cfg.CookieSecure, true)
}

func burstKeyMaterial(c *gin.Context) string {
	path := routeTemplateOrPath(c)
	return c.ClientIP() + "|" + c.Request.Method + "|" + path
}

func strikeKeyMaterial(c *gin.Context, rawToken string) string {
	fp := sha256.Sum256([]byte(rawToken))
	return burstKeyMaterial(c) + "|tok:" + hex.EncodeToString(fp[:16])
}

func routeTemplateOrPath(c *gin.Context) string {
	if fp := c.FullPath(); fp != "" {
		return fp
	}
	return c.Request.URL.Path
}

func allowRedisBurst(ctx context.Context, rdb *redis.Client, keyMat string, max int64, window time.Duration) (bool, error) {
	sum := sha256.Sum256([]byte(keyMat))
	k := fmt.Sprintf("{rl}b:%s", hex.EncodeToString(sum[:12]))

	n, err := rdb.Incr(ctx, k).Result()
	if err != nil {
		return false, err
	}
	if n == 1 {
		if err := rdb.Expire(ctx, k, window).Err(); err != nil {
			return false, err
		}
	}
	return n <= max, nil
}

func redisStrikeKey(mat string) string {
	sum := sha256.Sum256([]byte(mat))
	return fmt.Sprintf("{rl}s:%s", hex.EncodeToString(sum[:12]))
}

func redisIncrStrike(ctx context.Context, rdb *redis.Client, mat string, window time.Duration) (int64, error) {
	k := redisStrikeKey(mat)
	n, err := rdb.Incr(ctx, k).Result()
	if err != nil {
		return 0, err
	}
	if n == 1 {
		if err := rdb.Expire(ctx, k, window).Err(); err != nil {
			return n, err
		}
	}
	return n, nil
}

type memoryLimiter struct {
	mu     sync.Mutex
	window time.Duration
	bucket map[string]*memBucket
}

type memBucket struct {
	n     int
	until time.Time
}

func newMemoryLimiter(window time.Duration) *memoryLimiter {
	return &memoryLimiter{
		window: window,
		bucket: make(map[string]*memBucket),
	}
}

func (m *memoryLimiter) allow(keyMat string, max int) bool {
	sum := sha256.Sum256([]byte(keyMat))
	k := hex.EncodeToString(sum[:16])
	now := time.Now().UTC()

	m.mu.Lock()
	defer m.mu.Unlock()

	b, ok := m.bucket[k]
	if !ok || now.After(b.until) {
		m.bucket[k] = &memBucket{n: 1, until: now.Add(m.window)}
		return true
	}
	b.n++
	return b.n <= max
}

type memStrike struct {
	mu     sync.Mutex
	window time.Duration
	m      map[string]*memBucket
}

func newMemStrike(window time.Duration) *memStrike {
	return &memStrike{window: window, m: make(map[string]*memBucket)}
}

func (ms *memStrike) bump(mat string) int {
	sum := sha256.Sum256([]byte(mat))
	k := hex.EncodeToString(sum[:16])
	now := time.Now().UTC()

	ms.mu.Lock()
	defer ms.mu.Unlock()

	b, ok := ms.m[k]
	if !ok || now.After(b.until) {
		ms.m[k] = &memBucket{n: 1, until: now.Add(ms.window)}
		return 1
	}
	b.n++
	return b.n
}

func (ms *memStrike) clear(mat string) {
	sum := sha256.Sum256([]byte(mat))
	k := hex.EncodeToString(sum[:16])

	ms.mu.Lock()
	defer ms.mu.Unlock()

	delete(ms.m, k)
}

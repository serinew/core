package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds runtime configuration loaded from the environment.
type Config struct {
	HTTPPort     string
	PostgresDSN  string // libpq keyword/value form (safe for special chars in password)
	RedisURL     string
	TokenSecret  string // accessToken HMAC 재료 (민 16바이트 권장)
	CookieSecure bool   // HTTPS에서 Secure 쿠키 (COOKIE_SECURE=true)

	RateLimitEnabled bool          // RATE_LIMIT_ENABLED
	RateLimitMax     int           // RATE_LIMIT_MAX (윈도우당 허용 횟수)
	RateLimitWindow  time.Duration // RATE_LIMIT_WINDOW_SEC

	// RATE_LIMIT_REVOKE_AFTER_STRIKES — 429 이후 같은 토큰으로 추가 시도 시 횟수(0이면 비활성).
	RateLimitRevokeAfterStrikes int
}

func getenv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

// Load reads configuration from environment variables.
func Load() Config {
	port := getenv("PORT", "8080")

	host := getenv("DB_HOST", "localhost")
	dbPort := getenv("DB_PORT", "5432")
	user := getenv("DB_USER", "postgres")
	pass := getenv("DB_PASSWORD", "")
	name := getenv("DB_NAME", "postgres")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, dbPort, user, pass, name, getenv("DB_SSLMODE", "disable"),
	)

	rlMax := getenvIntPositive("RATE_LIMIT_MAX", 120)
	rlWin := getenvDurationSeconds("RATE_LIMIT_WINDOW_SEC", time.Minute)
	rlRevoke := getenvNonNegativeInt("RATE_LIMIT_REVOKE_AFTER_STRIKES", 0)

	return Config{
		HTTPPort:     port,
		PostgresDSN:  dsn,
		RedisURL:     getenv("REDIS_URL", "redis://localhost:6379"),
		TokenSecret:  strings.TrimSpace(os.Getenv("TOKEN_SECRET")),
		CookieSecure: getenvBool("COOKIE_SECURE", false),

		RateLimitEnabled:            getenvBool("RATE_LIMIT_ENABLED", false),
		RateLimitMax:                rlMax,
		RateLimitWindow:             rlWin,
		RateLimitRevokeAfterStrikes: rlRevoke,
	}
}

func getenvNonNegativeInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return fallback
	}
	return n
}

func getenvIntPositive(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return fallback
	}
	return n
}

func getenvDurationSeconds(key string, fallback time.Duration) time.Duration {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return fallback
	}
	return time.Duration(n) * time.Second
}

func getenvBool(key string, fallback bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	v = strings.ToLower(v)
	switch v {
	case "1", "true", "t", "yes", "y", "on":
		return true
	case "0", "false", "f", "no", "n", "off":
		return false
	default:
		return fallback
	}
}

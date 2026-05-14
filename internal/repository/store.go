package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const redisPingTimeout = 2 * time.Second

// Store holds Postgres (GORM) and optional Redis handles.
type Store struct {
	DB    *gorm.DB
	Redis *redis.Client
}

// Connect opens PostgreSQL via GORM and Redis (Redis failure is logged; Redis may be nil).
func Connect(ctx context.Context, postgresDSN, redisURL string) (*Store, func(), error) {
	db, err := gorm.Open(postgres.Open(postgresDSN), &gorm.Config{})
	if err != nil {
		return nil, nil, fmt.Errorf("postgres: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("postgres sql: %w", err)
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, nil, fmt.Errorf("postgres ping: %w", err)
	}
	applyPostgresPool(sqlDB)

	rdb := openRedisOptional(ctx, redisURL)

	stop := func() {
		if rdb != nil {
			_ = rdb.Close()
		}
		_ = sqlDB.Close()
	}

	log.Println("repository: postgres (gorm) connected")
	return &Store{DB: db, Redis: rdb}, stop, nil
}

func openRedisOptional(ctx context.Context, redisURL string) *redis.Client {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Printf("repository: redis URL invalid (%v), skipping", err)
		return nil
	}
	rdb := redis.NewClient(opts)
	pctx, cancel := context.WithTimeout(ctx, redisPingTimeout)
	defer cancel()
	if err := rdb.Ping(pctx).Err(); err != nil {
		log.Printf("repository: redis unreachable (%v), skipping", err)
		_ = rdb.Close()
		return nil
	}
	log.Println("repository: redis connected")
	return rdb
}

// applyPostgresPool caps concurrent server-side connections and recycles them so
// “many Connect()” / 핸들러 폭주 시에도 DB가 열 연결로 터지지 않게 합니다.
// (한 프로세스당 풀은 sql.DB 하나 — repository.Connect는 한 번만 호출하는 전제.)
func applyPostgresPool(sqlDB *sql.DB) {
	maxOpen := 25
	if v := os.Getenv("DB_MAX_OPEN_CONNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxOpen = n
		}
	}
	maxIdle := 5
	if v := os.Getenv("DB_MAX_IDLE_CONNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			maxIdle = n
		}
	}
	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(maxIdle)
	// Below typical Pg max session lifetime; renew before server drops the socket.
	sqlDB.SetConnMaxLifetime(55 * time.Minute)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)
	log.Printf("repository: postgres pool max_open=%d max_idle=%d", maxOpen, maxIdle)
}

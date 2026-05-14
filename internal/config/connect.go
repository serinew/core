package config

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const redisPingTimeout = 2 * time.Second

// SetupPostgresPool connects and pings Postgres. On failure logs and returns nil (server may still run).
func SetupPostgresPool(ctx context.Context, dsn string) *pgxpool.Pool {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Printf("postgres: init skipped (%v)", err)
		return nil
	}
	if err := pool.Ping(ctx); err != nil {
		log.Printf("postgres: unreachable (%v)", err)
		pool.Close()
		return nil
	}
	log.Println("postgres: connected")
	return pool
}

// SetupRedis connects and pings Redis.
// The returned cleanup must be deferred when non-nil; otherwise both returns are nil.
func SetupRedis(ctx context.Context, redisURL string) (_ *redis.Client, cleanup func()) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Printf("redis: invalid REDIS_URL (%v)", err)
		return nil, nil
	}
	rdb := redis.NewClient(opts)
	pctx, cancel := context.WithTimeout(ctx, redisPingTimeout)
	defer cancel()
	if err := rdb.Ping(pctx).Err(); err != nil {
		log.Printf("redis: unreachable (%v)", err)
		_ = rdb.Close()
		return nil, nil
	}
	log.Println("redis: connected")
	return rdb, func() { _ = rdb.Close() }
}

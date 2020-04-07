package rate

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gomodule/redigo/redis"
)

const defaultWindow = time.Second

// Limiter performs distributed control of events frequency using redis.
type Limiter struct {
	// Limit settings: allow `Limit` events per `Window` time interval
	Window time.Duration
	Limit  float64

	// Redis settings: connections pool, ZSET key and loaded script digest
	Pool   *redis.Pool
	Key    string
	Digest string

	// Error logging
	Log Logger
}

// NewLimiter creates new rate limiter with default redis pool settings.
// By default limit is a number of events per second. To change window
// time interval, change `Window` field manualy.
func NewLimiter(addr, redisKey string, limit float64) (*Limiter, error) {
	lim := &Limiter{
		Pool:   defaultRedisPool(addr),
		Key:    redisKey,
		Window: defaultWindow,
		Limit:  limit,
		Log:    defaultLogger(),
	}
	if err := lim.Init(); err != nil {
		return nil, fmt.Errorf("init limiter: %v", err)
	}
	return lim, nil
}

// Init loads script for calculating bucket size to redis.
func (lim *Limiter) Init() error {
	conn := lim.Pool.Get()
	defer conn.Close()

	digest, err := redis.String(conn.Do("SCRIPT", "LOAD", script))
	if err != nil {
		return fmt.Errorf("load script: %v", err)
	}
	lim.Digest = digest
	return nil
}

// Allow checks whether event is allowed to happen.
func (lim *Limiter) Allow() bool {
	conn := lim.Pool.Get()
	defer conn.Close()

	now := time.Now().UnixNano()
	allow, err := redis.Bool(conn.Do(
		"EVALSHA", lim.Digest, 1,
		lim.Key, lim.Window.Nanoseconds(), lim.Limit, now,
	))
	if err != nil {
		lim.Log.Printf("Redis request failed: %v", err)
		return true
	}
	return allow
}

// Logger describes general purpose logger.
type Logger interface {
	Printf(msg string, args ...interface{})
}

func defaultRedisPool(addr string) *redis.Pool {
	return &redis.Pool{
		Wait:        true,
		MaxActive:   1000,
		MaxIdle:     1000,
		IdleTimeout: 5 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", addr)
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

func defaultLogger() *log.Logger {
	return log.New(os.Stderr, "", log.LstdFlags)
}

// Lua script that makes main allow/deny descision. It uses redis ZSET
// structure for storing list of events that happened. On each call it
// removes events older than window interval. If limit is not reached,
// than new event is added, and script returns true. Otherwise event
// is not added, and false is returned.
const script = `
	local key = KEYS[1]
	local window = tonumber(ARGV[1])
	local limit = tonumber(ARGV[2])
	local now = tonumber(ARGV[3])

	redis.call('ZREMRANGEBYSCORE', key, '-inf', now-window)

	local amount = redis.call('ZCARD', key)
	if amount < limit then
		redis.call('ZADD', key, now, now)
		return 1
	end
	return 0
`

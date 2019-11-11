package storage

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"time"
)

var (
	RedisMaxIdle   = 30  // 池中的最大空闲连接数
	RedisMaxActive = 300 // 池在给定时间分配的最大连接数。为零时，池中的连接数没有限制。
)

const (
	redisIdleTimeoutSec = 30 // 在此期间保持空闲后关闭连接。如果该值为零，则不关闭空闲连接。应用程序应将超时设置为小于服务器超时的值。
)

// NewRedisPool  a new Redis connection pool.
// TestOnBorrow是一个可选的应用程序提供的函数，用于在应用程序再次使用连接之前检查空闲连接的运行状况。参数t是连接返回池的时间。如果函数返回错误，则关闭连接。
func NewRedisPool(redisURL string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     RedisMaxIdle,
		MaxActive:   RedisMaxActive,
		IdleTimeout: redisIdleTimeoutSec * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialURL(redisURL)
			if err != nil {
				log.Errorf("redis connection error: %v", err)
				return nil, fmt.Errorf("redis connection error: %s", err)
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			if err != nil {
				log.Errorf("ping redis error: %v", err)
				return fmt.Errorf("ping redis error: %v", err)
			}
			return nil
		},
	}
}

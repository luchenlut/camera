package storage

import (
	"camera/config"
	"github.com/garyburd/redigo/redis"
	"github.com/pkg/errors"
)

func Get(key string) (interface{}, error) {
	c := config.C.Redis.Pool.Get()
	defer c.Close()
	return c.Do("GET", key)
}

func SetEX(key string, expire int, value interface{}) error {
	c := config.C.Redis.Pool.Get()
	defer c.Close()
	_, err := c.Do("SETEX", key, expire, value)
	return err
}

func Set(key string, value interface{}) error {
	c := config.C.Redis.Pool.Get()
	defer c.Close()
	_, err := c.Do("SET", key, value)
	return err
}

func EXPIRE(key string, expire int) error {
	c := config.C.Redis.Pool.Get()
	defer c.Close()
	_, err := c.Do("EXPIRE", key, expire)
	return err
}

func Exists(key string) (bool, error) {
	c := config.C.Redis.Pool.Get()
	defer c.Close()

	r, err := redis.Int(c.Do("EXISTS", key))
	if err != nil {
		return false, errors.Wrap(err, "get exists error")
	}
	if r == 1 {
		return true, nil
	}
	return false, nil
}

func Delete(key string) error {
	c := config.C.Redis.Pool.Get()
	defer c.Close()

	_, err := c.Do("DEL", key)
	return err
}

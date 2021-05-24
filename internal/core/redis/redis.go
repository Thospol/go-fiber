package redis

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/garyburd/redigo/redis"
)

const (
	redisMaxIdle  int           = 3
	redisIdleTime time.Duration = 5 * time.Minute
)

var (
	c = &client{}
)

// Client regis client interface
type Client interface {
	Ping() error
	Get(key string, value interface{}) error
	GetKeys(pattern string) ([]string, error)
	Set(key string, value interface{}, expiredTime time.Duration) error
	Delete(key string) error
	Close()
	MapRedisKey(r *http.Request, data interface{}, prefixKey string) string
}

// Configuration config redis
type Configuration struct {
	Host     string
	Port     int
	Password string
}

// Init start redis connection
func Init(config Configuration) error {
	pool := &redis.Pool{
		MaxIdle:     redisMaxIdle,
		IdleTimeout: redisIdleTime,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port),
				redis.DialPassword(config.Password),
			)
		},
	}

	c = &client{
		pool: pool,
	}

	err := c.Ping()
	if err != nil {
		return err
	}

	return nil
}

// GetConnection get client connection
func GetConnection() Client {
	return c
}

// Client redis cache
type client struct {
	pool *redis.Pool
}

// Ping ping servier
func (cache *client) Ping() error {
	conn := cache.pool.Get()
	defer func() {
		_ = conn.Close()
	}()

	_, err := redis.String(conn.Do("PING"))
	if err != nil {
		return err
	}
	return nil
}

// Get get value from key
func (cache *client) Get(key string, value interface{}) error {
	conn := cache.pool.Get()
	defer func() {
		_ = conn.Close()
	}()
	str, err := redis.String(conn.Do("GET", key))
	if err != nil {
		return err
	}

	b := bytes.Buffer{}
	b.Write([]byte(str))
	d := gob.NewDecoder(&b)
	err = d.Decode(value)
	if err != nil {
		return err
	}

	return nil
}

// GetKeys get keys
func (cache *client) GetKeys(pattern string) ([]string, error) {
	conn := cache.pool.Get()
	defer func() {
		_ = conn.Close()
	}()

	iter := 0
	keys := []string{}
	for {
		arr, err := redis.Values(conn.Do("SCAN", iter, "MATCH", pattern))
		if err != nil {
			return keys, fmt.Errorf("error retrieving '%s' keys", pattern)
		}

		iter, _ = redis.Int(arr[0], nil)
		k, _ := redis.Strings(arr[1], nil)
		keys = append(keys, k...)

		if iter == 0 {
			break
		}
	}

	return keys, nil
}

// Set set value to key
func (cache *client) Set(key string, value interface{}, expiredTime time.Duration) error {
	conn := cache.pool.Get()
	defer func() {
		_ = conn.Close()
	}()

	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	err := e.Encode(value)
	if err != nil {
		return err
	}

	_, err = conn.Do("SET", key, b.Bytes())
	if err != nil {
		return err
	}

	if expiredTime.Seconds() > 1 {
		_, err = conn.Do("EXPIRE", key, expiredTime.Seconds())
		if err != nil {
			return err
		}
	}

	return err
}

// Delete delete key
func (cache *client) Delete(key string) error {
	conn := cache.pool.Get()
	defer func() {
		_ = conn.Close()
	}()

	_, err := conn.Do("DEL", key)
	return err
}

// Close close pool redis
func (cache *client) Close() {
	_ = cache.pool.Close()
}

// MapRedisKey map redis key
func (cache *client) MapRedisKey(r *http.Request, data interface{}, prefixKey string) string {
	key := prefixKey
	formValue := reflect.ValueOf(data)
	if formValue.Kind() == reflect.Ptr {
		formValue = formValue.Elem()
	}

	t := reflect.TypeOf(formValue.Interface())
	for i := 0; i < formValue.NumField(); i++ {
		if tag := t.Field(i).Tag.Get("form"); tag != "" {
			if formValue.FieldByName(t.Field(i).Name).IsValid() {
				if v := r.URL.Query().Get(tag); v != "" {
					key = key + fmt.Sprintf("_%s_%s", tag, v)
				}
			}
		}
	}

	return key
}

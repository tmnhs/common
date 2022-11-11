package redisclient

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/tmnhs/common/logger"
	"strconv"
	"time"
)

//redis连接
var _defaultRedis *redis.Client

func Init(addr, password string, db int) (r *redis.Client, err error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password, // no password set
		DB:       db,       // use default DB
	})
	_, err = client.Ping(context.Background()).Result()
	_defaultRedis = client
	return
}

func GetRedis() *redis.Client {
	if _defaultRedis == nil {
		logger.GetLogger().Error("redis is not initialized")
		return nil
	}
	return _defaultRedis
}

func GetIntFromRedis(ctx context.Context, key string) (int64, error) {
	s, err := GetStringFromRedis(ctx, key)
	if err != nil {
		return 0, err
	}
	appletID, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return appletID, nil
}

func HgetIntFromRedis(ctx context.Context, key string, value string) (int64, error) {
	s, err := HgetStringFromRedis(ctx, key, value)
	if err != nil {
		return 0, err
	}
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func GetStringFromRedis(ctx context.Context, key string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3000*time.Millisecond)
	defer cancel()

	result, err := _defaultRedis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", ErrRedisNotFound
		}
		return "", err
	}
	return result, nil
}

func HgetStringFromRedis(ctx context.Context, key string, value string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3000*time.Millisecond)
	defer cancel()

	result, err := _defaultRedis.HGet(ctx, key, value).Result()
	if err != nil {
		if err == redis.Nil {
			return "", ErrRedisNotFound
		}
		return "", err
	}
	return result, nil
}

func GetFromRedis(ctx context.Context, key string, object interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, 3000*time.Millisecond)
	defer cancel()

	result, err := _defaultRedis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrRedisNotFound
		}
		return err
	}

	if err := json.Unmarshal([]byte(result), object); err != nil {
		return fmt.Errorf("unmarshal error")
	}
	return nil
}

func SetToRedis(ctx context.Context, key string, object interface{}, expire time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, 3000*time.Millisecond)
	defer cancel()

	buf, err := json.Marshal(object)
	if err != nil {
		return err
	}
	_, err = _defaultRedis.Set(ctx, key, buf, expire).Result()
	if err != nil {
		return err
	}
	return nil
}

func DelFromRedis(ctx context.Context, key string) error {
	ctx, cancel := context.WithTimeout(ctx, 3000*time.Millisecond)
	defer cancel()

	_, err := _defaultRedis.Del(ctx, key).Result()
	return err
}

func GetIntArrayFromRedis(ctx context.Context, key string) ([]int64, error) {
	var a = make([]int64, 0)
	err := GetFromRedis(ctx, key, &a)
	if err == nil { // Redis Found
		return a, nil
	}
	if err != ErrRedisNotFound { // Redis Error
		return nil, err
	}
	// Redis Not Found
	return a, nil
}

package db

import (
	"context"
	"log"

	conf "api-pay/config"
	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client
var Ctx = context.Background()

// InitRedis 初始化Redis客户端
func InitRedis() error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     conf.AppConfig.Redis.Addr,
		Password: conf.AppConfig.Redis.Password,
		DB:       conf.AppConfig.Redis.DB,
	})

	// 测试连接
	if err := RedisClient.Ping(Ctx).Err(); err != nil {
		return err
	}

	log.Println("Connected to Redis successfully")
	return nil
}

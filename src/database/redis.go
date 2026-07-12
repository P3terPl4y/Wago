package database

import (
	"App/src/global"
	"context"
	"fmt"
	"log"
	"os"

	goredis "github.com/redis/go-redis/v9"
)

func InitRedis() {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		log.Println("⚠️ REDIS_URL no configurada, no se usará caché Redis.")
		return
	}
	opt, err := goredis.ParseURL(redisURL)
	if err != nil {
		log.Printf("❌ Error parseando REDIS_URL: %v, sin caché", err)
		return
	}
	global.RedisClient = goredis.NewClient(opt)
	ctx := context.Background()
	if err := global.RedisClient.Ping(ctx).Err(); err != nil {
		log.Printf("❌ Error conectando a Redis: %v, sin caché", err)
		global.RedisClient = nil
	} else {
		fmt.Println("✅ Conectado a Redis para caché de historial")
	}
}

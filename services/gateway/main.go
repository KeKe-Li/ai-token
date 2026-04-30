package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"

	"github.com/KeKe-Li/ai-token/services/gateway/internal/config"
	"github.com/KeKe-Li/ai-token/services/gateway/internal/model"
	"github.com/KeKe-Li/ai-token/services/gateway/internal/router"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	db, err := model.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Printf("警告: 数据库连接失败，部分功能不可用: %v", err)
	} else {
		defer db.Close()
		log.Println("数据库连接成功")
	}

	var rdb *redis.Client
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err == nil {
		rdb = redis.NewClient(opt)
		if err := rdb.Ping(ctx).Err(); err != nil {
			log.Printf("警告: Redis 连接失败，限流功能不可用: %v", err)
			rdb = nil
		} else {
			defer rdb.Close()
			log.Println("Redis 连接成功")
		}
	}

	r := router.Setup(cfg, db, rdb)

	go func() {
		log.Printf("AI Token Gateway 启动于 :%s", cfg.Port)
		if err := r.Run(":" + cfg.Port); err != nil {
			log.Fatalf("服务启动失败: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("正在关闭服务...")
}

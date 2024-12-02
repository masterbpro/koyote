package redis

import (
	"context"
	"fmt"
	"github.com/grbit/go-json"
	"github.com/koyote/pkg/telegram"
	"github.com/savsgio/gotils/uuid"
	"time"

	"github.com/go-redis/redis/v8"
	log "github.com/gookit/slog"
	"github.com/koyote/pkg/config"
)

var ctx = context.Background()
var redisClient *redis.Client

func ConnectToRedis() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%v:%v", config.GlobalAppConfig.Redis.Host, config.GlobalAppConfig.Redis.Port),
		Password: config.GlobalAppConfig.Redis.Auth.Password,
		DB:       config.GlobalAppConfig.Redis.DB,
	})

	err := redisClient.Set(ctx, "check", "connection", 5*time.Second).Err()
	if err != nil {
		log.Fatal("Error while connect to redis. Error: ", err)
	}

	ProcessRedisData(redisClient, 3*time.Second)
}

func ProcessRedisData(redisClient *redis.Client, pollInterval time.Duration) {
	go func() {
		for {
			keys, err := redisClient.Keys(ctx, "event:*").Result()
			if err != nil {
				log.Printf("Error fetching keys from Redis: %v\n", err)
				time.Sleep(pollInterval)
				continue
			}

			for _, key := range keys {
				value, err := redisClient.Get(ctx, key).Result()
				if err != nil {
					log.Printf("Error fetching value for key %s: %v\n", key, err)
					continue
				}

				if !json.Valid([]byte(value)) {
					log.Printf("Invalid JSON data for key %s: %s\n", key, value)
					continue
				}

				var event MessageType
				err = json.Unmarshal([]byte(value), &event)
				if err != nil {
					log.Printf("Error unmarshaling data for key %s: %v\n", key, err)
					continue
				}

				err = telegram.SendEventMessage(event.ChatID, event.ThreadID, event.Message)
				if err != nil {
					log.Printf("Can't send message from Redis", err)
				}

				err = redisClient.Del(ctx, key).Err()
				if err != nil {
					log.Printf("Error deleting key %s: %v\n", key, err)
				} else {
					log.Printf("Successfully processed and deleted key: %s\n", key)
				}
			}
			time.Sleep(pollInterval)
		}
	}()
}

func PublishEventToRedisChannel(data MessageType) {
	log.Info("Saving received event to Redis")
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("Error serializing event data: %v", err)
	}

	err = redisClient.Set(ctx, fmt.Sprintf("event:%s", uuid.V4()), jsonData, 0).Err()
	if err != nil {
		log.Fatalf("Error saving event to Redis: %v", err)
	}
}

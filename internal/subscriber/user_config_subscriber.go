package subscriber

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/nikhilthakur8/advoid/services"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func SubscribeToUserConfigUpdatesRedisChannel() {

	opt, _ := redis.ParseURL(os.Getenv("REDIS_URI"))
	opt.TLSConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	rdb := redis.NewClient(opt)

	_, err := rdb.Ping(ctx).Result()

	if err != nil {
		panic(fmt.Sprintf("Failed to connect to Redis: %v", err))
	}

	sub := rdb.Subscribe(ctx, "user_config_updates")

	go func() {
		ch := sub.Channel()

		for msg := range ch {
			var userId struct {
				UserID int `json:"userId"`
			}
			err := json.Unmarshal([]byte(msg.Payload), &userId)
			if err != nil {
				fmt.Printf("Error unmarshaling message: %v\n", err)
				continue
			}
			fmt.Println("Received user config update for user ID:", userId.UserID)
			go services.InvalidateUserConfigCache(strconv.Itoa(userId.UserID))
		}
	}()

}

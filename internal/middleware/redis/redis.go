package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
)

//实现某个聊天室最新的10条聊天记录

// 进去聊天室获取历史记录，page=1时首先先去redis中查询，如果命中则返回前十条，
var RdbRoomMessageList *redis.Client
var RdbVistorList *redis.Client
var Ctx = context.Background()

func InitRedis() {
	RdbRoomMessageList = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	RdbVistorList = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1,
	})
}

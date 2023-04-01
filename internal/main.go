package main

import (
	"chat/internal/middleware/casbin"
	rabbitmq "chat/internal/middleware/rabbitmq"
	redis "chat/internal/middleware/redis"
	"chat/internal/router"
	"chat/internal/service"
)

func main() {
	go service.Manager.Start()
	e := router.Router()
	rabbitmq.InitRabbitMQ()
	rabbitmq.InitAddRoomMQ()
	rabbitmq.InitDeleteRoomMQ()
	redis.InitRedis()
	casbin.InitEnforcer()
	e.Run(":3000")
}

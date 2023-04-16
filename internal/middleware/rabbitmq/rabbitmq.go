package rabbitmq

import (
	"chat/logs"

	"github.com/streadway/amqp"
)

const MQURL = "amqp://guest:guest@127.0.0.1:5672/"

type RabbitMQ struct {
	conn  *amqp.Connection
	mqurl string
}

var Rmq *RabbitMQ

// InitRabbitMQ 初始化RabbitMQ的连接和通道。
func InitRabbitMQ() {

	Rmq = &RabbitMQ{
		mqurl: MQURL,
	}
	dial, err := amqp.Dial(Rmq.mqurl)
	if err != nil {
		logs.ReLogrusObj(logs.Path).Error("rabbitmq init err:", err)
		return
	}
	Rmq.conn = dial

}

// 关闭mq通道和mq的连接。
func (r *RabbitMQ) destroy() {
	r.conn.Close()
}

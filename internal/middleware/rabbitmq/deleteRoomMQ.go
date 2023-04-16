package rabbitmq

import (
	"chat/internal/model"
	"chat/logs"
	"strings"

	"github.com/streadway/amqp"
)

type DeleteRoomMQ struct {
	RabbitMQ
	channel   *amqp.Channel
	exchange  string
	queueName string
}

func NewDeleteRoomMQ(queueName string) *DeleteRoomMQ {
	deleteRoomMQ := &DeleteRoomMQ{
		RabbitMQ:  *Rmq,
		queueName: queueName, //friendQue groupQue
	}

	ch, err := deleteRoomMQ.conn.Channel()
	deleteRoomMQ.channel = ch
	if err != nil {
		logs.ReLogrusObj(logs.Path).Error("new deleteRoomMQ err:", err)
	}
	return deleteRoomMQ
}
func (c *DeleteRoomMQ) Publish(message string) {

	_, err := c.channel.QueueDeclare(
		c.queueName,
		//是否持久化
		false,
		//是否为自动删除
		false,
		//是否具有排他性
		false,
		//是否阻塞
		false,
		//额外属性
		nil,
	)
	if err != nil {
		logs.ReLogrusObj(logs.Path).Warn("DeleteRoomMQ publish-->QueueDeclare err:", err)
		return
	}

	err1 := c.channel.Publish(
		c.exchange,
		c.queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
	if err1 != nil {
		logs.ReLogrusObj(logs.Path).Warn("DeleteRoomMQ publish-->Publish err:", err1)
		return
	}
}
func (r *DeleteRoomMQ) Consumer() {

	_, err := r.channel.QueueDeclare(r.queueName, false, false, false, false, nil)

	if err != nil {
		logs.ReLogrusObj(logs.Path).Warn("escGroup consume error-->QueueDeclare :", err)
		return
	}

	//2、接收消息
	msg, err := r.channel.Consume(
		r.queueName,
		//用来区分多个消费者
		"",
		//是否自动应答
		true,
		//是否具有排他性
		false,
		//如果设置为true，表示不能将同一个connection中发送的消息传递给这个connection中的消费者
		false,
		//消息队列是否阻塞
		false,
		nil,
	)
	if err != nil {
		logs.ReLogrusObj(logs.Path).Warn("escGroup consume error-->Consume:", err)
		return
	}

	forever := make(chan bool)
	go r.EscGroup(msg)
	//log.Printf("[*] Waiting for messages,To exit press CTRL+C")
	<-forever

}
func (r *DeleteRoomMQ) EscGroup(msg <-chan amqp.Delivery) {
	for req := range msg {
		params := strings.Split(string(req.Body), ",")
		//roomid,uid

		err := model.EscRoom(params[0], params[1])
		// redis.RdbRoomUserList.LRem(redis.Ctx,params[0],)
		if err != nil {
			logs.ReLogrusObj(logs.Path).Warn("escGroup consume error-->EscRoom:", err)
		}
	}

}
func InitDeleteRoomMQ() {
	drm2 := NewDeleteRoomMQ("escGroup")
	go drm2.Consumer()

}

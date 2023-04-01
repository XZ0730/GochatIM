package rabbitmq

import (
	"chat/internal/model"
	"log"
	"strings"
	"time"

	"github.com/streadway/amqp"
)

type AddRoomMQ struct {
	RabbitMQ
	channel   *amqp.Channel
	exchange  string
	queueName string
}

// NewAddRoomMQ 获取AddRoomMQ的对应队列。
func NewAddRoomMQ(queueName string) *AddRoomMQ {
	addRoomMQ := &AddRoomMQ{
		RabbitMQ:  *Rmq,
		queueName: queueName, //friendQue groupQue
	}

	ch, err := addRoomMQ.conn.Channel()
	addRoomMQ.channel = ch
	Rmq.failOnErr(err, "获取通道失败")
	return addRoomMQ
}

// addroom Produce
func (c *AddRoomMQ) Publish(message string) {

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
		panic(err)
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
		panic(err)
	}
}

// addroom消费者
func (r *AddRoomMQ) Consumer() {

	_, err := r.channel.QueueDeclare(r.queueName, false, false, false, false, nil)

	if err != nil {
		panic(err)
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
		panic(err)
	}

	forever := make(chan bool)
	switch {
	case r.queueName == "friendQue":
		go r.AddFriend(msg)
	case r.queueName == "groupQue":
		go r.InsertToGroup(msg)
	}
	//log.Printf("[*] Waiting for messages,To exit press CTRL+C")

	<-forever

}

// 加好友 ————>将用户塞入群聊：传入roomid-roomname-type1-uid1-uid2 创建房间信息，创建用户房间关联信息
func (r *AddRoomMQ) AddFriend(msg <-chan amqp.Delivery) {
	for req := range msg {
		params := strings.Split(string(req.Body), ",")
		//params[0]:roomId 1:Name 2:Type
		roomId := params[0]
		name := params[1]
		type1 := params[2]
		cr := &model.Room_Basic{
			RoomID:    roomId,
			Name:      name,
			Info:      "",
			Type:      type1,
			Create_At: time.Now().Unix(),
			Update_At: time.Now().Unix(),
		}
		params1 := strings.Split(name, "-")
		err2 := model.CreatePrivateRoom(cr, params1[0], params1[1])
		if err2 != nil {
			log.Println("adroommq.go line 121 err2:", err2)
		}
	}
}

// 加入群聊 ————>将用户塞入群聊
func (r *AddRoomMQ) InsertToGroup(msg <-chan amqp.Delivery) {
	for req := range msg {
		params := strings.Split(string(req.Body), ",")
		ur := &model.UserRoom{
			UserID:    params[0],
			RoomID:    params[1],
			Type:      params[2],
			Create_At: time.Now().Unix(),
			Update_At: time.Now().Unix(),
		}
		err2 := model.CreateUserRoom(ur)
		if err2 != nil {
			log.Println("addroommq.go line 139 err2:", err2)
		}
	}
}

// 开启协程监听队列
func InitAddRoomMQ() {
	friendmq := NewAddRoomMQ("friendQue")
	go friendmq.Consumer()
	groupmq := NewAddRoomMQ("groupQue")
	go groupmq.Consumer()
}

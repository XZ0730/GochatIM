package service

import (
	"chat/internal/middleware/casbin"
	"chat/internal/model"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID     string
	Socket *websocket.Conn
	Send   chan []byte
	Judge  bool
}
type BroadCast struct {
	Client  *Client
	Message []byte
	RoomID  string
	Type    string
}
type ClientManager struct {
	Clients    map[string]*Client
	BroadCast  chan *BroadCast
	ReplyMsg   chan *Client
	Register   chan *Client
	Unregister chan *Client
}

var Manager = ClientManager{ //链接管理员
	Clients:    make(map[string]*Client),
	BroadCast:  make(chan *BroadCast),
	Register:   make(chan *Client),
	Unregister: make(chan *Client),
}

func (manager *ClientManager) Start() {
	for {
		log.Println("<---监听管道通信--->")
		select {
		case client := <-Manager.Register: // 建立连接
			log.Printf("建立新连接: %v", client.ID)
			Manager.Clients[client.ID] = client
			msg := []byte("链接成功")
			_ = client.Socket.WriteMessage(websocket.TextMessage, msg)

		case client := <-Manager.Unregister: // 断开连接
			log.Printf("连接失败:%v", client.ID)
			if _, ok := Manager.Clients[client.ID]; ok {
				msg := []byte("链接断开")
				_ = client.Socket.WriteMessage(websocket.TextMessage, msg)
				//删除policy -------删除游客policy
				if client.Judge { //判断是否是游客
					_, err := casbin.Enfocer.RemoveGroupingPolicy(client.ID, "vistor")
					if err != nil {
						log.Println("游客policy")
					}
				}
				fmt.Println("uuid:", client.ID)
				close(client.Send)
				delete(Manager.Clients, client.ID)
			}
		case broadcast := <-Manager.BroadCast: //广播，读协程读到自身用户发送的消息后，
			//传入广播通道，由广播通道根据roomid广播到特定用户
			users, err2 := model.GetUsersByRid(broadcast.RoomID) //查询房间内用户
			if err2 != nil {
				log.Println(err2)
				panic(err2)
			}
			mb := &model.Meesage_Basic{
				UserID:    broadcast.Client.ID,       //发送者的id
				RoomID:    broadcast.RoomID,          //房间号
				Data:      string(broadcast.Message), //信息
				Create_At: time.Now().Unix(),         //发送时间
				Update_At: time.Now().Unix(),
			}
			err := model.InsertOne(mb) //插入数据
			if err != nil {
				fmt.Println(err)
				panic(err)
			}
			for _, user := range users { //遍历房间用户
				if client, ok := Manager.Clients[user.UserID]; ok { //在线
					select {
					case client.Send <- broadcast.Message: //往clientsend写数据用于后面的写协程发送
						continue
					default:
						close(client.Send)
						delete(Manager.Clients, client.ID)
					}
				}

			}

		}
	}
}

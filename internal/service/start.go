package service

import (
	"chat/internal/middleware/casbin"
	"chat/internal/middleware/redis"
	"chat/internal/model"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/panjf2000/ants"
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
	Clients         sync.Map
	BroadCast       chan *BroadCast
	BroadCastPublic chan *BroadCast
	ReplyMsg        chan *Client
	Register        chan *Client
	Unregister      chan *Client
}

var Manager = ClientManager{ //链接管理员
	BroadCast:       make(chan *BroadCast),
	BroadCastPublic: make(chan *BroadCast),
	Register:        make(chan *Client),
	Unregister:      make(chan *Client),
}
var (
	pool       *ants.PoolWithFunc
	publicpool *ants.PoolWithFunc
	publicwait sync.WaitGroup
	wait       sync.WaitGroup
)

func init() {
	//协程池--->负责读取
	var err error
	pool, err = ants.NewPoolWithFunc(500, func(payload interface{}) { //连接池 获取userid并返回
		request, ok := payload.(*job) //断言
		if !ok {
			return
		}

		if client, ok := Manager.Clients.Load(request.Userid); ok { //在线
			select {
			case client.(*Client).Send <- request.Message: //往clientsend写数据用于后面的写协程发送
				log.Println("send ok")
			default:
				// close(client.(*Client).Send)
				// Manager.Clients.Delete(request.Userid)
				Manager.Unregister <- client.(*Client)
			}
		}
		fmt.Println("-------------------------")
		wait.Done()
	})
	if err != nil {
		log.Println("ants pool error:new pool error")
		panic("ants pool error:new pool error")
	}
	publicpool, err = ants.NewPoolWithFunc(1000, func(payload interface{}) { //连接池 获取userid并返回
		request, ok := payload.(*job) //断言
		if !ok {
			return
		}

		if client, ok := Manager.Clients.Load(request.Userid); ok { //在线
			select {
			case client.(*Client).Send <- request.Message: //往clientsend写数据用于后面的写协程发送
				log.Println("send ok")
			default:
				Manager.Unregister <- client.(*Client)
			}
		}
		publicwait.Done()
	})
	if err != nil {
		log.Println("ants pool error:new pool error")
		panic("ants pool error:new pool error")
	}
}

type job struct {
	Message []byte
	Userid  string
}

func (manager *ClientManager) Start() {
	for {
		log.Println("<---监听管道通信--->")
		select {
		case client := <-Manager.Register: // 建立连接
			log.Printf("建立新连接: %v", client.ID)
			Manager.Clients.Store(client.ID, client)
			msg := []byte("链接成功")
			_ = client.Socket.WriteMessage(websocket.TextMessage, msg)

		case client := <-Manager.Unregister: // 断开连接
			log.Printf("连接失败:%v", client.ID)
			if _, ok := Manager.Clients.Load(client.ID); ok {
				msg := []byte("链接断开")
				fmt.Println("msg", string(msg))
				_ = client.Socket.WriteMessage(websocket.TextMessage, msg)
				//删除policy -------删除游客policy
				if client.Judge { //判断是否是游客
					_, err := casbin.Enfocer.RemoveGroupingPolicy(client.ID, "vistor")
					if err != nil {
						log.Println("游客policy")
					}
				}
				fmt.Println("uuid:", client.ID)
				_ = client.Socket.Close()
				close(client.Send)
				Manager.Clients.Delete(client.ID)
			}
		case broadcast := <-Manager.BroadCast: //广播，读协程读到自身用户发送的消息后，
			//传入广播通道，由广播通道根据roomid广播到特定用户
			// _, err := casbin.Enfocer.RemoveGroupingPolicy(broadcast.Client.ID, "vistor")
			// if err != nil {
			// 	log.Println("游客policy")
			// }
			var err3 error
			log.Println("----time", time.Now().Unix())
			usersID := make([]string, 0)
			usersID, err3 = redis.RdbRoomUserList.LRange(redis.Ctx, broadcast.RoomID, 0, -1).Result()
			fmt.Println("---------------------------------------")
			fmt.Println("usersid:", usersID)
			if err3 != nil || len(usersID) == 0 { //没有取到值
				log.Println("broadcast get user error")
				users, err2 := model.GetUsersByRid(broadcast.RoomID) //查询房间内用户
				if err2 != nil {
					log.Println("mongodb get user error by roomid:", broadcast.RoomID)
				}
				for _, v := range users {
					usersID = append(usersID, v.UserID)
					err := redis.RdbRoomUserList.LPush(redis.Ctx, broadcast.RoomID, v.UserID).Err()
					if err != nil {
						log.Println("redis RdbRoomUserList push error")
						break
					}
					err4 := redis.RdbRoomUserList.Expire(redis.Ctx, broadcast.RoomID, time.Duration(time.Now().Day()*30)*time.Second).Err()
					if err4 != nil {
						log.Println("redis RdbRoomUserList ExpireSet error")
						break
					}
				}
			}
			//得到users存入redis
			fmt.Println("---------------------------------------")
			fmt.Println("usersid:", usersID)
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
			for _, userid := range usersID { //遍历房间用户
				job := &job{
					Message: broadcast.Message,
					Userid:  userid,
				}
				wait.Add(1)
				pool.Invoke(job)
			}
			wait.Wait()
			log.Println("----time", time.Now().Unix())
		case broadcast := <-Manager.BroadCastPublic:
			//将公共频道进行拆分，单独进行消息发送，
			//理由：首先公共频道访问人数多，处理的消息请求较多，
			//此时和登录用户用同一个通道显然不合适，会大大拖慢消息的发送速度
			//其次 公共频道的逻辑和私有频道逻辑混杂在一起会使得代码可读性变差
			//同时增加代码逻辑的复杂度
			//公共频道逻辑应该如此：在保证并发安全的情况下，遍历所有sync.map中的值并进行发送消息
			//问：为什么使用sync.map而不去使用原生map加锁的组合
			//答：因为sync.map在读多写少(只在用户链接和断开的时候会进行写的操作)的情况下要比另外一种组合更优，能够大大降低锁的竞争
			//:
			msg := broadcast.Message
			Manager.Clients.Range(func(key, value any) bool {
				//这边有个问题，如果发生中途插入了大量的key，sync.map的遍历是如何处理的呢
				//显然这个不适合我们这个场景，为什么，因为sunc.map本身就是适合写多的场景，
				//当前场景下的map是只在链接和退出的时候会进行写入,而且这边是用的非阻塞态的select多路复用的通道
				//因此这边保证了一个安全就是这边代码没有执行完，并不会进行一个写的操作也就是进行下一轮的轮询
				//同时我认为这边下面的代码也应该交给协程去处理，因为在上面的协程池逻辑中我们将启动协程去给send通道去发送消息
				//因此send不可避免地可能陷入阻塞，倘若发生阻塞，则会大大影响select非阻塞态下其他路的执行--也就是降低响应的速度

				//还有一个问题是为什么用两个池子，理由其实和上面一样，共有和私有进行分离，共有处理的请求要更多
				job := &job{
					Message: msg,
					Userid:  key.(string),
				}
				publicwait.Add(1)
				publicpool.Invoke(job)
				return true
			})
			publicwait.Wait()
		}
	}
}

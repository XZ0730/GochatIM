package service

import (
	"chat/internal/middleware/redis"
	"chat/internal/model"
	"chat/util"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type MsgDTO struct {
	Message interface{} `json:"data"`
	RoomID  string      `json:"room_id"`
}

func Sendmsg(c *gin.Context) {
	fmt.Println("-----------------------")
	conn, err := (&websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}).Upgrade(c.Writer, c.Request, nil) //ws升级
	fmt.Println("=============================")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "ws异常",
		})
	}
	id := ""
	judge := false
	claims, _ := util.ParseToken(c.GetHeader("token"))
	if claims == nil {
		judge = true //判断是不是游客
		id, err = redis.RdbVistorList.Get(c, c.Request.RemoteAddr).Result()
		if err != nil {
			log.Println("redis get error:", err)
		}
	} else {
		id = claims.ID
	}
	client := &Client{
		ID:     id,
		Socket: conn,
		Send:   make(chan []byte),
		Judge:  judge,
	}
	fmt.Println("-------------------------")
	Manager.Register <- client
	go client.ReadMsg()
	go client.WriteMsg()
}
func (c *Client) ReadMsg() {
	defer func() {
		_ = c.Socket.Close()
	}()
	for {
		ms := new(MsgDTO)
		err2 := c.Socket.ReadJSON(ms)
		if err2 != nil {
			panic(err2)
		}
		//TODO:判断用户是否属于消息体房间
		ok := model.JudgeIsInROOM(c.ID, ms.RoomID)
		if !ok {
			fmt.Println("user is no exist in the room")
			return
		}
		Manager.BroadCast <- &BroadCast{
			Client:  c,
			Message: []byte(ms.Message.(string)),
			RoomID:  ms.RoomID,
		}
	}
}
func (c *Client) WriteMsg() {
	defer func() {
		_ = c.Socket.Close()
	}() //房间内广播消息
	for {
		select { //监听send管道
		case message, ok := <-c.Send:
			if !ok {
				_ = c.Socket.WriteMessage(websocket.TextMessage, []byte{})
			}
			_ = c.Socket.WriteMessage(websocket.TextMessage, message)

		}
	}
}

//cannot convert ms.Message (variable of type interface{}) to []byte (need type assertion)

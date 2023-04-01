package service

import (
	"chat/internal/middleware/rabbitmq"
	"chat/internal/middleware/redis"
	"chat/internal/model"
	"chat/internal/vo"
	"chat/util"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ChatService struct {
	RoomName string `json:"name" form:"name"`
	RoomID   string `json:"room_id" form:"room_id"`
}

var pageSize int64 = 3

// 获取聊天历史-->传入rid和page
func GetHistoryList(c *gin.Context) {
	rid, ok := c.GetQuery("rid")
	if !ok || rid == "" {
		c.JSON(http.StatusOK, gin.H{
			"status": 322,
			"msg":    "房间号错误",
		})
	}
	claims, _ := util.ParseToken(c.GetHeader("token"))
	_, err3 := model.GetURinfo(claims.ID, rid)
	if err3 != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": 325,
			"msg":    err3.Error(),
		})
	}
	pageStr, ok1 := c.GetQuery("page")
	if !ok1 {
		pageStr = "1"
	}
	page, err := strconv.ParseInt(pageStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": 323,
			"msg":    err.Error(),
		})
	}
	skip := (page - 1) * pageSize
	//TODO:查询某个房间的消息记录//默认最近的3条
	// mbs, err2 := model.GetMsgByRid(rid, &pageSize, &skip)
	// if err2 != nil {
	// 	c.JSON(http.StatusOK, gin.H{
	// 		"status": 324,
	// 		"msg":    err2.Error(),
	// 	})
	// }
	if skip > 100 {
		mbs, err2 := model.GetMsgByRid(rid, &pageSize, &skip)
		if err2 != nil {
			c.JSON(http.StatusOK, gin.H{
				"status": 324,
				"msg":    err2.Error(),
			})
		}
		c.JSON(http.StatusOK, vo.BuildList(mbs, len(mbs), 0))
		//一百条以后的数据通过数据库查找
	}
	//返回最新的一百条-->一百条以上或者是redis数据过期，就从数据库中查找
	//首先查看redis中能否命中从redis中读取
	if cnt, err := redis.RdbRoomMessageList.Exists(redis.Ctx, rid).Result(); cnt > 0 {
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"status": 323,
				"err":    err.Error(),
			})
		}
		//存在
		msgs, err := redis.RdbRoomMessageList.LRange(c, rid, 0, -1).Result()
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"status": 324,
				"err":    err.Error(),
			})
		}
		c.JSON(http.StatusOK, vo.BuildList(msgs, len(msgs), 0))

	} else {
		if _, err := redis.RdbRoomMessageList.LPush(redis.Ctx, rid, -1).Result(); err != nil {
			redis.RdbRoomMessageList.Del(redis.Ctx, rid)
			c.JSON(http.StatusOK, gin.H{
				"status": 323,
				"err":    err.Error(),
			})
		} //防止脏读
		var pagesize int64 = 100
		var skipto int64 = 0
		mbs, err := model.GetMsgByRid(rid, &pagesize, &skipto)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"status": 324,
				"err":    err.Error(),
			})
		}
		for _, msg := range mbs {
			redis.RdbRoomMessageList.LPush(redis.Ctx, rid, msg.Data)
		}
		_, err2 := redis.RdbRoomMessageList.Expire(redis.Ctx, rid,
			time.Duration(time.Now().Day()*30)*time.Second).
			Result()
		if err2 != nil {
			log.Println("redis save one room-msg expire failed")
		}
		c.JSON(http.StatusOK, vo.BuildList(mbs, len(mbs), 0))
	}
}

// 创建群聊--->创建群聊房间--type=>"2"
func (cs *ChatService) CreateChatRoom(uid string) vo.Response {
	room := &model.Room_Basic{
		Admin_id:  uid,
		RoomID:    uid + strconv.Itoa(int(time.Now().Unix())),
		Name:      cs.RoomName,
		Info:      "",
		Type:      "2",
		Create_At: time.Now().Unix(),
		Update_At: time.Now().Unix(),
	}
	err := model.CreateChatRoom(room)
	if err != nil {
		return vo.Response{
			Status: 400,
			Error:  err.Error(),
		}
	}
	return vo.Response{
		Status: 200,
		Msg:    "ok",
	}
}

// 搜索好友，按好友account搜索--->添加好友--->创建单聊房间--type=>"1"

// 获取当前聊天室具体信息-->rid->roombasic  //输入roomid查询群聊房间--
func (cs *ChatService) GetRoomInfo() vo.Response {
	room, err := model.GetRoomByRid(cs.RoomID)
	if err != nil {
		return vo.Response{
			Status: 391,
			Msg:    "获取房间信息错误",
		}
	}
	users, err := model.GetUsersByRid(cs.RoomID)
	if err != nil {
		return vo.Response{
			Status: 391,
			Msg:    "获取房间信息错误",
		}
	}

	return vo.BuildList(users, len(users), room)
}

// 加入群聊--->查询群聊房间-->点击加入-->创建userroom记录
func (cs *ChatService) InsertUserToRoom(uid string, rid string) vo.Response {
	ok := model.JudgeIsInROOM(uid, rid)
	if ok {
		return vo.Response{
			Status: 392,
			Msg:    "已经在群聊中了",
		}
	}
	r, err := model.GetRoomByRid(rid)
	fmt.Println("rid:", rid)
	fmt.Println("len:", len(rid))
	if err != nil {
		fmt.Println("r:", r)
		return vo.Response{
			Status: 391,
			Msg:    "查询不到此房间",
			Error:  err.Error(),
		}
	}
	msg := uid + "," + r.RoomID + "," + "2"
	rabbitmq.NewAddRoomMQ("groupQue").Publish(msg)
	return vo.Response{
		Status: 200,
		Msg:    "ok",
	}
}

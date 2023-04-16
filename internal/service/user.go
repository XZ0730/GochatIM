package service

import (
	"chat/internal/middleware/casbin"
	rabbitmq "chat/internal/middleware/rabbitmq"
	"chat/internal/model"
	"chat/internal/vo"
	"chat/logs"
	"chat/util"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/panjf2000/ants"
)

type UserService struct {
	Account  string `json:"account" form:"account"`
	Passwrod string `json:"password" form:"password"`
	Nickname string `json:"nickname" form:"nickname"`
}
type TestXH struct {
	Result chan int64
}

var IDpool *ants.PoolWithFunc
var workderc *util.Worker
var wait2 sync.WaitGroup

func init() {
	workderc, _ = util.NewWorker(int64(1))
	var err error
	IDpool, err = ants.NewPoolWithFunc(500, func(payload interface{}) { //连接池 获取userid并返回
		request, ok := payload.(*TestXH) //断言
		if !ok {
			return
		}
		id := workderc.GetId()
		request.Result <- id
		wg.Done()
	})
	if err != nil {
		log.Println("ants error:", err)
	}
}
func (service *UserService) Register() vo.Response {
	if service.Account == "" || service.Nickname == "" || service.Passwrod == "" {
		return vo.Response{
			Status: 397,
			Msg:    "不能为空",
		}
	}
	ok := model.JudgeIsExist(service.Account)
	if ok {
		return vo.Response{
			Status: 398,
			Msg:    "用户名已存在",
		}
	}
	addChant := make(chan bool)                                       //数据库结果通知通道
	pool, _ := ants.NewPoolWithFunc(300, func(userInfo interface{}) { //执行数据库插入
		user, ok := userInfo.(*model.User_Basic)
		if !ok {
			logs.ReLogrusObj(logs.Path).Debug("User_Basic assert fail")
			wait2.Done()
			return
		}
		if user != nil { //传递数据添加
			addStatus := model.Create_User(user)
			addChant <- addStatus
		}
		wait2.Done()
	})
	request := &TestXH{Result: make(chan int64)}
	//snowflake
	ub := &model.User_Basic{
		Account:   service.Account,
		Password:  service.Passwrod,
		Nickname:  service.Nickname,
		Create_At: time.Now().Unix(),
		Update_At: time.Now().Unix(),
	}
	wait2.Add(1)
	IDpool.Invoke(request)

	uid := <-request.Result
	fmt.Println("uid:", uid)
	ub.ID = strconv.FormatInt(uid, 10)

	wg.Add(1)
	pool.Invoke(ub)

	addsta := <-addChant
	if !addsta {
		log.Println("注册失败")
		return vo.Response{
			Status: 397,
			Msg:    "注册失败",
		}
	}
	msg := ub.ID + "," + "38324" + "," + "2"
	rabbitmq.NewAddRoomMQ("groupQue").Publish(msg) //用户注册加入公共频道
	//加入游客策略---加入用户策略
	_, err2 := casbin.Enfocer.AddGroupingPolicy(ub.ID, "vistor")
	if err2 != nil {
		logs.ReLogrusObj(logs.Path).Debug("[用户", ub.ID, "]", "加入游客策略失败")
	}

	_, err := casbin.Enfocer.AddGroupingPolicy(ub.ID, "user")
	if err != nil {
		logs.ReLogrusObj(logs.Path).Debug("[用户", ub.ID, "]", "加入用户策略失败")
	}
	wait2.Wait()
	return vo.Response{
		Status: 200,
		Msg:    "ok",
	}
}

// 登录service
func (service *UserService) Login() vo.Response {
	if service.Account == "" || service.Passwrod == "" {
		return vo.Response{
			Status: 500,
			Msg:    "账户密码空",
		}
	}
	ub, err := model.GetUserBasicByAccount(service.Account)
	if err != nil {
		return vo.Response{
			Status: 500,
			Error:  err.Error(),
		}
	}
	if ub.Password != service.Passwrod {
		return vo.Response{
			Status: 333,
			Msg:    "密码错误",
		}
	}
	data, _ := util.GenerateToken(ub.ID, service.Passwrod, service.Account)
	return vo.Response{
		Status: http.StatusOK,
		Msg:    "ok",
		Data:   data,
	}
}

// 获取用户信息service
func (service *UserService) GetUserInfo(uid string) vo.Response {
	u, err := model.GetUserBasicByAccount(service.Account)
	if err != nil {
		return vo.Response{
			Status: 333,
			Error:  err.Error(),
		}
	}

	return vo.Response{
		Data:   vo.BuildUser(u, uid),
		Status: http.StatusOK,
		Msg:    "ok",
	}
}

// 获得当前用户的聊天室列表-->返回rid，讲rid
func (service *UserService) GetChatRooms(uid string) vo.Response {
	//根据uid查询userroom 返回所有userroom
	rooms, err := model.GetRoomsByuid(uid)
	if err != nil {
		return vo.Response{
			Status: 339,
			Msg:    "拉取聊天室失败",
		}
	}
	return vo.Response{
		Status: 200,
		Data:   rooms, //待会序列化一下
		Msg:    "ok",
	}
}
func (*UserService) AddFriend(uid1 string, account string) vo.Response {
	//判断账号是否为空
	if account == "" {
		return vo.Response{
			Status: 334,
			Msg:    "账号为空",
		}
	}
	//查询账号是否存在
	u, err := model.GetUserBasicByAccount(account)
	if err != nil {
		return vo.Response{
			Status: 336,
			Msg:    "账号不存在",
		}
	}
	//查询是否是好友
	fmt.Println(u)
	b := model.JudgeIsFriend(uid1, u.ID)
	if b {
		return vo.Response{
			Status: 337,
			Msg:    "已经是好友，请勿重复添加",
		}
	}
	//创建私聊房间
	// cr := &model.Room_Basic{
	// 	RoomID:    uid1 + u.ID + strconv.Itoa(int(time.Now().Unix())),
	// 	Name:      uid1 + "-" + u.ID,
	// 	Info:      "",
	// 	Type:      "1",
	// 	Create_At: time.Now().Unix(),
	// 	Update_At: time.Now().Unix(),
	// }
	message := uid1 + u.ID + strconv.Itoa(int(time.Now().Unix())) + "," + uid1 + "-" + u.ID +
		"," + "1"
	rabbitmq.NewAddRoomMQ("friendQue").Publish(message)
	return vo.Response{
		Status: 200,
		Msg:    "ok",
	}
}

// 删除好友--->删除单聊房间和聊天记录 :
func (service *UserService) EscRoom(uid, roomid string) vo.Response {
	//直接传入消息队列，然后删除
	msg := roomid + "," + uid
	rabbitmq.NewDeleteRoomMQ("escGroup").Publish(msg)

	return vo.Response{
		Status: 200,
		Msg:    "ok",
	}
}

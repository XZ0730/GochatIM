package casbin

import (
	"chat/internal/middleware/redis"
	"chat/logs"
	"chat/util"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

var Enfocer *casbin.Enforcer

func InitEnforcer() {
	a, err := gormadapter.NewAdapter("mysql", "root:root@tcp(43.136.122.18:3307)/casbin", true)
	if err != nil {
		fmt.Println("----")
		panic(err)
	}
	var err1 error

	Enfocer, err1 = casbin.NewEnforcer("D:/Golang/chat/internal/middleware/casbin/model.conf", a)
	// fmt.Println("---------------------------")
	if err1 != nil {
		// fmt.Println("----------------")
		panic(err1)
	}
	// sub := "zhangsan"
	// obj := "data3"
	// act := "read"
	// e.AddPolicy(sub, obj, act)
	// e.AddGroupingPolicy("zhangsan", "jolo")
	// ok, err := e.Enforce(sub, obj, act)
	// if err != nil {
	// 	panic(err)
	// }
	return
	// if ok {
	// 	fmt.Println("ok")
	// } else {
	// 	fmt.Println("no")
	// }
}

func CheckAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		var id string
		cliam2, _ := util.ParseToken(c.GetHeader("token"))
		if cliam2 == nil { //游客处理
			log.Println("[游客登录]")
			//这边可以不直接abort  如果是作为游客，可以开放一个公共频道去聊天
			//发送消息前的中间件，发送消息之前先判断你token对不对，如果不对，则为游客登录，发送消息
			//公共频道的消息不做记录
			//游客登录分配一个id，然后加入policy
			//游客断开链接的时候进行销毁
			//将用户的addr 作为key value为用户的uuid
			var err error

			params := strings.Split(c.Request.RemoteAddr, ":")
			// fmt.Println("params", params)
			// c.IsWebsocket()
			id, err = redis.RdbVistorList.Get(redis.Ctx, params[0]).Result()
			if err != nil { //redis中没有--就set
				id = util.UUID()
				//缓存有效期，三个月
				err2 := redis.RdbVistorList.Set(redis.Ctx, params[0], id, time.Duration(time.Now().Month()*3)*time.Second).Err()
				if err2 != nil {
					log.Println("redis set error:", err2)
				}
				fmt.Println(id)
				_, err3 := Enfocer.AddGroupingPolicy(id, "vistor")
				if err3 != nil {
					log.Println("vistor group add err1:", err)
				}
			} else { //redis中存在kv，直接取出构造用户角色映射关系
				b2, err2 := Enfocer.AddGroupingPolicy(id, "vistor")
				if err2 != nil || !b2 {
					log.Println("vistor group add err2:", err2)
				}
			}

		} else { //claims不为空---存在token||token鉴权成功
			log.Println("[用户登录]")
			if time.Now().Unix() > cliam2.ExpiresAt {
				c.JSON(http.StatusOK, gin.H{
					"status": 401,
					"msg":    "token过期",
				})
				c.Abort()
				return
			}
			id = cliam2.ID
		}
		//统一获取url---method --id
		obj_uri := c.Request.URL
		act := c.Request.Method
		sub := id
		path := obj_uri.Path
		if strings.HasPrefix(obj_uri.Path, "/u/room") {
			path = "/u/room"
		}
		fmt.Println(sub, ":", path, ":", act)
		b, _ := Enfocer.Enforce(sub, path, act)
		if b {
			logs.ReLogrusObj(logs.Path).Debug("[user]", sub, "--[obj]", obj_uri, "--[act]", act, "------[timestramp]", time.Now().Format("2006-05-04 15:02:01"))
			c.Next()
			return
		} else {
			logs.ReLogrusObj(logs.Path).Debug("[autheicaton lose user]", sub, "--[obj]", obj_uri, "--[act]", act, "------[timestramp]", time.Now().Format("2006-05-04 15:02:01"))
			c.Abort()
			return
		}
	}
}

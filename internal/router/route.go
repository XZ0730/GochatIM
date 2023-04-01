package router

import (
	api "chat/internal/api"
	"chat/internal/middleware/casbin"
	"chat/internal/service"

	"github.com/gin-gonic/gin"
)

func Router() *gin.Engine {
	r := gin.Default()
	r.POST("/login", api.Login)
	r.POST("/register", api.Register)

	r.Use(casbin.CheckAuth())                //鉴权授权
	r.GET("/ws/msg", service.Sendmsg)        //发送消息
	r.POST("/msg/file", api.FileUpload)      //发送图片和语音-->分片上传
	r.GET("/u/chat", service.GetHistoryList) //获取历史消息
	r.GET("/u/user", api.GetUserInfo)        //获取用户信息

	r.POST("/u/chat", api.CreateChatRoom) //创建群聊
	r.GET("/u/room", api.GetUserRooms)    //获取房间列表
	r.POST("/u/room", api.GetRoomInfo)    //发送rid,获取房间信息

	r.POST("/u/room/:rid", api.InsertUserToRoom) //通过rid加入群聊
	r.DELETE("/u/room/:rid", api.EscGroup)       //删除好友---退出群聊

	r.POST("/u/user", api.AddFriend) //添加好友
	return r
}

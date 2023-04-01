package api

import (
	"chat/internal/service"
	"chat/internal/vo"
	"chat/util"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
	var service *service.UserService
	fmt.Println(c.Request.RemoteAddr)
	fmt.Println(c.Request.URL)
	if err := c.ShouldBind(&service); err == nil {
		rsp := service.Login()
		c.JSON(http.StatusOK, rsp)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"err": err.Error(),
		})
	}
}
func Register(c *gin.Context) {
	var user *service.UserService
	if err := c.ShouldBind(&user); err == nil {
		r := user.Register()
		c.JSON(http.StatusOK, r)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}
}
func GetUserInfo(c *gin.Context) {
	var service *service.UserService
	fmt.Println("---------------------------------")
	if err := c.ShouldBind(&service); err == nil {
		fmt.Println("---------------------------------1")
		c2, err2 := util.ParseToken(c.GetHeader("token"))
		fmt.Println("---------------------------------2")
		if err2 != nil {
			fmt.Println("---------------------------------")
			panic(err2)
		}
		rsp := service.GetUserInfo(c2.ID)
		c.JSON(http.StatusOK, rsp)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"err": err.Error(),
		})
	}
}
func GetUserRooms(c *gin.Context) {
	claims, err := util.ParseToken(c.GetHeader("token"))
	if err != nil {
		c.JSON(http.StatusBadRequest, vo.Response{
			Status: http.StatusBadRequest,
			Msg:    "token解析错误",
		})
	}
	var Userservice *service.UserService
	rsp := Userservice.GetChatRooms(claims.ID)
	c.JSON(http.StatusOK, rsp)

}
func AddFriend(c *gin.Context) {
	var service *service.UserService
	claims, err := util.ParseToken(c.GetHeader("token"))
	if err != nil {
		c.JSON(http.StatusBadRequest, vo.Response{
			Status: http.StatusBadRequest,
			Msg:    "token解析错误",
		})
	}
	if err := c.ShouldBind(&service); err == nil {
		rsp := service.AddFriend(claims.ID, service.Account)
		c.JSON(http.StatusOK, rsp)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"err": err.Error(),
		})
	}
}
func EscGroup(c *gin.Context) {
	var service *service.UserService
	claims, err := util.ParseToken(c.GetHeader("token"))
	if err != nil {
		c.JSON(http.StatusBadRequest, vo.Response{
			Status: http.StatusBadRequest,
			Msg:    "token解析错误",
		})
	}
	rid := c.Param("rid")
	if rid == "" {
		c.JSON(http.StatusOK, vo.Response{
			Status: 488,
			Msg:    "群聊号为空",
		})
	}
	rsp := service.EscRoom(claims.ID, rid)
	c.JSON(http.StatusOK, rsp)
}

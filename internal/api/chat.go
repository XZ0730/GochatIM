package api

import (
	"chat/internal/service"
	"chat/internal/vo"
	"chat/util"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateChatRoom(c *gin.Context) {
	var cs *service.ChatService
	claims, _ := util.ParseToken(c.GetHeader("token"))
	if err := c.ShouldBind(&cs); err == nil {
		fmt.Println("--------------------")
		rsp := cs.CreateChatRoom(claims.ID)
		c.JSON(http.StatusOK, rsp)
	} else {
		c.JSON(http.StatusBadRequest, err)
	}
}
func GetRoomInfo(c *gin.Context) {
	var cs *service.ChatService
	if err := c.ShouldBind(&cs); err == nil {
		fmt.Println("--------------------")
		rsp := cs.GetRoomInfo()
		c.JSON(http.StatusOK, rsp)
	} else {
		c.JSON(http.StatusBadRequest, err)
	}
}

func InsertUserToRoom(c *gin.Context) {
	var cs *service.ChatService
	fmt.Println(c.Request.URL.Path)
	claims, _ := util.ParseToken(c.GetHeader("token"))

	fmt.Println("--------------------")
	rid := c.Param("rid")
	if rid == "" {
		c.JSON(http.StatusOK, vo.Response{
			Status: 488,
			Msg:    "群聊号为空",
		})
	}
	fmt.Println("rid:", rid)
	rsp := cs.InsertUserToRoom(claims.ID, rid)
	c.JSON(http.StatusOK, rsp)
}
func FileUpload(c *gin.Context) {
	form, _ := c.MultipartForm()
	files := form.File["file"]
	var fileservice *service.FileService
	c2, err := util.ParseToken(c.GetHeader("token"))
	if err != nil {
		c.JSON(http.StatusBadRequest, vo.Response{
			Status: http.StatusBadRequest,
			Msg:    "鉴权失败",
			Error:  err.Error(),
		})
	}
	rid, ok := c.GetQuery("room_id")
	if !ok || rid == "" {
		c.JSON(http.StatusBadRequest, vo.Response{
			Status: 3398,
			Error:  "房间号不存在",
		})
	}
	if err := c.ShouldBind(&fileservice); err == nil {
		rsp := fileservice.SendFile(files, c2.ID, rid)
		c.JSON(http.StatusOK, rsp)
	} else {
		c.JSON(http.StatusBadRequest, vo.Response{
			Status: http.StatusBadRequest,
			Error:  err.Error(),
		})
	}

}

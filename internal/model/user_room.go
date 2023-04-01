package model

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

type UserRoom struct {
	UserID    string `bson:"user_id"`
	RoomID    string `bson:"room_id"`
	Type      string `bson:"room_type"`
	Create_At int64  `bson:"create_at"`
	Update_At int64  `bson:"update_at"`
}

func (UserRoom) CollectionName() string {
	return "user_room"
}

func GetURinfo(userid, roomid string) (*UserRoom, error) {
	ur := new(UserRoom)
	err := Mongo.Collection(UserRoom{}.CollectionName()).
		FindOne(context.Background(), bson.D{{Key: "user_id", Value: userid}, {Key: "room_id", Value: roomid}}).
		Decode(ur)
	return ur, err
}
func GetUsersByRid(roomid string) (users []*UserRoom, err error) {
	cursor, err := Mongo.Collection(UserRoom{}.CollectionName()).
		Find(context.Background(), bson.D{{Key: "room_id", Value: roomid}})
	for cursor.Next(context.Background()) {
		ur := new(UserRoom)
		err = cursor.Decode(ur)
		if err != nil {
			return nil, err
		}
		users = append(users, ur)
	}
	return
}
func JudgeIsFriend(uid1 string, uid2 string) bool {
	cursor, err := Mongo.Collection(UserRoom{}.CollectionName()).
		Find(context.Background(),
			bson.D{{Key: "user_id", Value: uid1},
				{Key: "room_type", Value: "1"}})
	if err != nil {
		fmt.Println("err:", err)
	}
	roomId := make([]string, 0)
	for cursor.Next(context.Background()) {
		ur := new(UserRoom)
		err := cursor.Decode(ur)
		if err != nil {
			fmt.Println("err:", err)
			return false
		}
		roomId = append(roomId, ur.RoomID)
	}
	fmt.Println("roomid:", roomId)
	cnt, err := Mongo.Collection(UserRoom{}.CollectionName()).
		CountDocuments(context.Background(), bson.M{"user_id": uid2, "room_id": bson.M{"$in": roomId}})
	if err != nil {
		fmt.Println("err:", err)
		return false
	}
	if cnt > 0 {
		return true
	}
	return false
}

func JudgeIsInROOM(uid string, roomid string) bool {
	cnt, err := Mongo.Collection(UserRoom{}.CollectionName()).
		CountDocuments(context.Background(),
			bson.D{{Key: "user_id", Value: uid},
				{Key: "room_id", Value: roomid}})
	if err != nil {
		return false
	}
	if cnt > 0 {
		return true
	}
	return false
}
func GetRoomsByuid(uid string) (rooms []*UserRoom, err error) {

	cursor, err := Mongo.Collection(UserRoom{}.CollectionName()).
		Find(context.Background(), bson.D{{Key: "user_id", Value: uid}})

	for cursor.Next(context.Background()) {
		ur := &UserRoom{}
		err = cursor.Decode(ur)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, ur)
	}
	return
}
func CreateUserRoom(ur *UserRoom) error {
	_, err := Mongo.Collection(UserRoom{}.CollectionName()).
		InsertOne(context.Background(), ur)

	return err
}
func EscRoom(roomid, uid string) (err error) {
	//删除用户房间关联
	_, err = Mongo.Collection(UserRoom{}.CollectionName()).DeleteMany(context.Background(),
		bson.D{{Key: "room_id", Value: roomid}, {Key: "user_id", Value: uid}})
	if err != nil {
		return err
	}
	cnt, _ := Mongo.Collection(UserRoom{}.CollectionName()).CountDocuments(context.Background(),
		bson.D{{Key: "room_id", Value: roomid}})
	fmt.Println("cnt:", cnt)
	if cnt != 0 { //用户为空删除聊天记录
		return nil
	}
	//删除聊天记录
	_, err = Mongo.Collection(Meesage_Basic{}.CollectionName()).DeleteMany(context.Background(),
		bson.D{{Key: "room_id", Value: roomid}})
	if err != nil {
		return
	}
	//删除房间信息
	_, err = Mongo.Collection(Room_Basic{}.CollectionName()).DeleteOne(context.Background(),
		bson.D{{Key: "room_id", Value: roomid}})
	return
}

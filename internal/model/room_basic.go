package model

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

type Room_Basic struct {
	Admin_id  string `bson:"user_id"`
	RoomID    string `bson:"room_id"`
	Name      string `bson:"name"`
	Info      string `bson:"info"`
	Type      string `bson:"type"`
	Create_At int64  `bson:"create_at"`
	Update_At int64  `bson:"update_at"`
}

func (Room_Basic) CollectionName() string {
	return "room_basic"
}
func CreateChatRoom(cr *Room_Basic) error {
	_, err := Mongo.Collection(Room_Basic{}.CollectionName()).
		InsertOne(context.Background(), cr)
	if err != nil {
		return err
	}
	ur := &UserRoom{
		UserID:    cr.Admin_id,
		RoomID:    cr.RoomID,
		Create_At: cr.Create_At,
		Update_At: cr.Update_At,
	}
	_, err = Mongo.Collection(UserRoom{}.CollectionName()).
		InsertOne(context.Background(), ur)
	return err
}
func CreatePrivateRoom(cr *Room_Basic, uid1, uid2 string) error {
	_, err := Mongo.Collection(Room_Basic{}.CollectionName()).
		InsertOne(context.Background(), cr)
	if err != nil {
		return err
	}
	ur1 := &UserRoom{
		UserID:    uid1,
		RoomID:    cr.RoomID,
		Type:      cr.Type,
		Create_At: cr.Create_At,
		Update_At: cr.Update_At,
	}
	ur2 := &UserRoom{
		UserID:    uid2,
		RoomID:    cr.RoomID,
		Type:      cr.Type,
		Create_At: cr.Create_At,
		Update_At: cr.Update_At,
	}

	_, err = Mongo.Collection(UserRoom{}.CollectionName()).
		InsertOne(context.Background(), ur1)
	if err != nil {
		return err
	}
	_, err = Mongo.Collection(UserRoom{}.CollectionName()).
		InsertOne(context.Background(), ur2)
	return err
}
func GetRoomByRid(rid string) (room *Room_Basic, err error) {
	room = new(Room_Basic)
	err = Mongo.Collection(Room_Basic{}.CollectionName()).
		FindOne(context.Background(), bson.D{{Key: "room_id", Value: rid},
			{Key: "type", Value: "2"}}).
		Decode(room)
	fmt.Println("ridtooto:", rid)
	fmt.Println("err:", err)
	return
}

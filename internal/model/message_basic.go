package model

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Meesage_Basic struct {
	UserID    string      `bson:"user_id"`
	RoomID    string      `bson:"room_id"`
	Data      interface{} `bson:"data"`
	Create_At int64       `bson:"create_at"`
	Update_At int64       `bson:"update_at"`
}

func (Meesage_Basic) CollectionName() string {
	return "message_basic"
}
func InsertOne(mb *Meesage_Basic) error {
	_, err := Mongo.Collection(mb.CollectionName()).
		InsertOne(context.Background(), mb)
	return err
}
func GetMsgByRid(roomid string, pageSize, skip *int64) (mbs []*Meesage_Basic, err error) {
	cursor, err := Mongo.Collection(Meesage_Basic{}.CollectionName()).
		Find(context.Background(), bson.M{"room_id": roomid},
			&options.FindOptions{
				Limit: pageSize,
				Skip:  skip,
				Sort: bson.D{{
					Key: "create_at", Value: -1,
				}},
			})
	if err != nil {
		return nil, err
	}

	// Msg := make(bson.D, 1)
	// Msg = append(Msg, bson.E{Key: "a", Value: "d"})
	for cursor.Next(context.Background()) {
		mb := new(Meesage_Basic)
		err2 := cursor.Decode(mb)
		if err2 != nil {
			return nil, err2
		}
		mbs = append(mbs, mb)
	}
	return
}

//options.Find (value of type func() *"go.mongodb.org/mongo-driver/mongo/options".FindOptions) is not a type

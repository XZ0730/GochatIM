package model

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
)

type User_Basic struct {
	ID        string `bson:"user_id"`
	Account   string `bson:"account"`
	Password  string `bson:"password"`
	Nickname  string `bson:"nickname"`
	Email     string `bson:"email"`
	Sex       uint   `bson:"sex"`
	Avatar    string `bson:"avatar"`
	Create_At int64  `bson:"create_at"`
	Update_At int64  `bson:"update_at"`
}

func (User_Basic) CollectionName() string {
	return "user_basic"
}
func Create_User(user *User_Basic) bool {
	_, err := Mongo.Collection(User_Basic{}.CollectionName()).
		InsertOne(context.Background(), user)
	return err == nil
}
func JudgeIsExist(account string) bool {
	cnt, _ := Mongo.Collection(User_Basic{}.CollectionName()).
		CountDocuments(context.Background(), bson.M{"account": account})
	
	return cnt > 0

}
func GetUserBasicByAccount(account string) (*User_Basic, error) {
	ub := new(User_Basic)
	err := Mongo.Collection(User_Basic{}.CollectionName()).
		FindOne(context.Background(), bson.D{{Key: "account", Value: account}}).
		Decode(ub)

	return ub, err
}

//cannot convert (User_Basic literal).UserMongoName() (value of type string) to mongo.Collection

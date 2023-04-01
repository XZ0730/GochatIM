package vo

import "chat/internal/model"

type UserInfo struct {
	NickName string
	Account  string
	Avatar   string
	Isfriend bool
}

func BuildUser(ub *model.User_Basic, uid string) *UserInfo {
	return &UserInfo{
		NickName: ub.Nickname,
		Account:  ub.Account,
		Avatar:   ub.Avatar,
		Isfriend: model.JudgeIsFriend(uid, ub.ID),
	}
}

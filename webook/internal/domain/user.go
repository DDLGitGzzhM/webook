package domain

import "time"

type User struct {
	Id       int64
	Email    string
	PassWord string
	Phone    string
	Nickname string
	Ctime    time.Time
	WeChatInfo
}

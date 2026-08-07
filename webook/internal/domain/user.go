package domain

import "time"

type User struct {
	Id    int64
	Email string
	// <input type={{.input}}>
	PassWord string `fe:"input=password"`
	Phone    string
	Nickname string
	Ctime    time.Time
	WeChatInfo
}

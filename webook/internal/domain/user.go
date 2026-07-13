package domain

import "time"

type User struct {
	Id       int64
	Email    string
	PassWord string
	Ctime    time.Time
}

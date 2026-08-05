package domain

import "time"

// Msg 本地消息表领域对象
type Msg struct {
	Id      int64
	Content string
	Ctime   time.Time
}

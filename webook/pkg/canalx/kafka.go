package canalx

// Message 可以根据需要把其它字段也加入进来。
// T 直接对应到表结构
type Message[T any] struct {
	Data     []T    `json:"data"`
	Database string `json:"database"`
	Table    string `json:"table"`
	Type     string `json:"type"`
}

type MessageV1 struct {
	Data []struct {
		Id      string      `json:"id"`
		Uid     string      `json:"uid"`
		Biz     string      `json:"biz"`
		BizId   string      `json:"biz_id"`
		RootId  interface{} `json:"root_id"`
		Pid     interface{} `json:"pid"`
		Content string      `json:"content"`
		Ctime   string      `json:"ctime"`
		Utime   string      `json:"utime"`
	} `json:"data"`
	Database  string `json:"database"`
	Es        int64  `json:"es"`
	Gtid      string `json:"gtid"`
	Id        int    `json:"id"`
	IsDdl     bool   `json:"isDdl"`
	MysqlType struct {
		Id      string `json:"id"`
		Uid     string `json:"uid"`
		Biz     string `json:"biz"`
		BizId   string `json:"biz_id"`
		RootId  string `json:"root_id"`
		Pid     string `json:"pid"`
		Content string `json:"content"`
		Ctime   string `json:"ctime"`
		Utime   string `json:"utime"`
	} `json:"mysqlType"`
	Old     interface{} `json:"old"`
	PkNames []string    `json:"pkNames"`
	Sql     string      `json:"sql"`
	SqlType struct {
		Id      int `json:"id"`
		Uid     int `json:"uid"`
		Biz     int `json:"biz"`
		BizId   int `json:"biz_id"`
		RootId  int `json:"root_id"`
		Pid     int `json:"pid"`
		Content int `json:"content"`
		Ctime   int `json:"ctime"`
		Utime   int `json:"utime"`
	} `json:"sqlType"`
	Table string `json:"table"`
	Ts    int64  `json:"ts"`
	Type  string `json:"type"`
}

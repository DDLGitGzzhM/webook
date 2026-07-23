package domain

// Interactive 总体交互计数
type Interactive struct {
	//Biz   string
	//BizId int64

	ReadCnt    int64 `json:"read_cnt"`
	LikeCnt    int64 `json:"like_cnt"`
	CollectCnt int64 `json:"collect_cnt"`
	// Liked / Collected 表示当前用户是否点赞或收藏
	// 也可拆成独立结构体
	Liked     bool `json:"liked"`
	Collected bool `json:"collected"`
}

type Self struct {
	Liked     bool `json:"liked"`
	Collected bool `json:"collected"`
}

type Collection struct {
	Name  string
	Uid   int64
	Items []Resource
}

// max(发送者总速率/单一分区写入速率, 发送者总速率/单一消费者速率) + buffer

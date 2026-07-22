package ginx

// Result 可通过定义更多字段来配合 Wrap 方法
type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

package events

// SyncDataEvent 搜索通用同步事件。
type SyncDataEvent struct {
	IndexName string
	DocID     string
	Data      string
}

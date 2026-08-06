package events

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
)

//go:generate mockgen -source=./producer.go -package=evtmocks -destination=mocks/producer.mock.go Producer

// Producer 标签同步事件生产者。
type Producer interface {
	ProduceSyncEvent(ctx context.Context, data BizTags) error
}

// SaramaSyncProducer 基于 Sarama 的同步生产者。
type SaramaSyncProducer struct {
	client sarama.SyncProducer
}

// NewSaramaSyncProducer 创建标签同步事件生产者。
func NewSaramaSyncProducer(client sarama.SyncProducer) Producer {
	return &SaramaSyncProducer{client: client}
}

// ProduceSyncEvent 发送标签数据到搜索同步 topic。
func (p *SaramaSyncProducer) ProduceSyncEvent(ctx context.Context, tags BizTags) error {
	data, _ := json.Marshal(tags)
	evt := SyncDataEvent{
		IndexName: "tags_index",
		// 构成一个唯一的 doc id
		// 要确保后面打了新标签的时候，搜索那边也会有对应的修改
		DocID: fmt.Sprintf("%d_%s_%d", tags.Uid, tags.Biz, tags.BizId),
		Data:  string(data),
	}
	data, _ = json.Marshal(evt)
	_, _, err := p.client.SendMessage(&sarama.ProducerMessage{
		Topic: "search_sync_data",
		Value: sarama.ByteEncoder(data),
	})
	return err
}

// BizTags 业务资源上的标签集合。
type BizTags struct {
	Uid   int64    `json:"uid"`
	Biz   string   `json:"biz"`
	BizId int64    `json:"biz_id"`
	// 只传递 string
	Tags []string `json:"tags"`
}

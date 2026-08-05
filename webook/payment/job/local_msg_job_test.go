package job

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"webook/webook/payment/domain"
	"webook/webook/payment/events"
)

type fakeLocalMsgRepo struct {
	msgs        []domain.Msg
	successIDs  []int64
	failedIDs   []int64
	findErr     error
	markFailErr error
}

func (f *fakeLocalMsgRepo) AddMsg(ctx context.Context, content string) (int64, error) {
	return 0, nil
}

func (f *fakeLocalMsgRepo) MarkSuccess(ctx context.Context, id int64) error {
	f.successIDs = append(f.successIDs, id)
	return nil
}

func (f *fakeLocalMsgRepo) MarkFailed(ctx context.Context, id int64) error {
	f.failedIDs = append(f.failedIDs, id)
	return f.markFailErr
}

func (f *fakeLocalMsgRepo) FindInitMsg(ctx context.Context, offset, limit int) ([]domain.Msg, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	if offset >= len(f.msgs) {
		return nil, nil
	}
	end := offset + limit
	if end > len(f.msgs) {
		end = len(f.msgs)
	}
	return f.msgs[offset:end], nil
}

type fakeProducer struct {
	err error
	cnt int
}

func (f *fakeProducer) ProducePaymentEvent(ctx context.Context, evt events.PaymentEvent) error {
	f.cnt++
	return f.err
}

func TestLocalMsgResendJob_Run(t *testing.T) {
	evt := events.PaymentEvent{BizTradeNO: "t1", Status: 2}
	content, err := json.Marshal(evt)
	require.NoError(t, err)

	t.Run("发送成功标记 success", func(t *testing.T) {
		repo := &fakeLocalMsgRepo{
			msgs: []domain.Msg{{Id: 1, Content: string(content), Ctime: time.Now()}},
		}
		producer := &fakeProducer{}
		job := NewLocalMsgResendJob(repo, producer, time.Minute*10)
		require.NoError(t, job.Run())
		assert.Equal(t, []int64{1}, repo.successIDs)
		assert.Empty(t, repo.failedIDs)
	})

	t.Run("超时失败标记 failed", func(t *testing.T) {
		repo := &fakeLocalMsgRepo{
			msgs: []domain.Msg{{
				Id:      2,
				Content: string(content),
				Ctime:   time.Now().Add(-time.Minute * 20),
			}},
		}
		producer := &fakeProducer{err: errors.New("mq down")}
		job := NewLocalMsgResendJob(repo, producer, time.Minute*10)
		require.NoError(t, job.Run())
		assert.Equal(t, []int64{2}, repo.failedIDs)
		assert.Empty(t, repo.successIDs)
	})
}

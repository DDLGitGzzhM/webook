package validator

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/ecodeclub/ekit/syncx/atomicx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"webook/webook/internal/pkg/logger"
	"webook/webook/pkg/migrator"
)

type testEntity struct {
	Id    int64
	Value string
}

func (t testEntity) ID() int64 {
	return t.Id
}

func (t testEntity) CompareTo(dst migrator.Entity) bool {
	other, ok := dst.(testEntity)
	if !ok {
		return false
	}
	return t.Id == other.Id && t.Value == other.Value
}

func TestValidator_toMap(t *testing.T) {
	v := &Validator[testEntity]{}
	data := []testEntity{
		{Id: 1, Value: "a"},
		{Id: 2, Value: "b"},
		{Id: 3, Value: "c"},
	}

	got := v.toMap(data)
	require.Len(t, got, 3)
	assert.Equal(t, testEntity{Id: 1, Value: "a"}, got[1])
	assert.Equal(t, testEntity{Id: 2, Value: "b"}, got[2])
	assert.Equal(t, testEntity{Id: 3, Value: "c"}, got[3])
}

func TestValidator_toMap_Empty(t *testing.T) {
	v := &Validator[testEntity]{}
	got := v.toMap(nil)
	require.NotNil(t, got)
	assert.Empty(t, got)
}

func TestIsPoolHighLoad(t *testing.T) {
	t.Parallel()
	assert.False(t, isPoolHighLoad(sql.DBStats{}, 0.8))
	assert.False(t, isPoolHighLoad(sql.DBStats{
		MaxOpenConnections: 10,
		InUse:              7,
	}, 0.8))
	assert.True(t, isPoolHighLoad(sql.DBStats{
		MaxOpenConnections: 10,
		InUse:              8,
	}, 0.8))
}

func TestIsThreadsHighLoad(t *testing.T) {
	t.Parallel()
	assert.False(t, isThreadsHighLoad(10, 50))
	assert.True(t, isThreadsHighLoad(50, 50))
	assert.False(t, isThreadsHighLoad(100, 0))
}

func TestIsMemHighLoad(t *testing.T) {
	t.Parallel()
	assert.False(t, isMemHighLoad(100, 200))
	assert.True(t, isMemHighLoad(200, 200))
	assert.False(t, isMemHighLoad(1<<30, 0))
}

func TestWaitIfHighLoad_Resume(t *testing.T) {
	t.Parallel()
	v := &Validator[testEntity]{
		highLoad:        atomicx.NewValueOf(true),
		highLoadBackoff: 10 * time.Millisecond,
		l:               logger.NopLogger{},
	}
	go func() {
		time.Sleep(30 * time.Millisecond)
		v.highLoad.Store(false)
	}()
	assert.False(t, v.waitIfHighLoad(context.Background()))
}

func TestWaitIfHighLoad_Cancel(t *testing.T) {
	t.Parallel()
	v := &Validator[testEntity]{
		highLoad:        atomicx.NewValueOf(true),
		highLoadBackoff: time.Second,
		l:               logger.NopLogger{},
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	assert.True(t, v.waitIfHighLoad(ctx))
}

func TestMonitorLoad_StopsOnCancel(t *testing.T) {
	t.Parallel()
	v := &Validator[testEntity]{
		highLoad:          atomicx.NewValueOf(false),
		loadCheckInterval: 20 * time.Millisecond,
		heapAllocThreshold: 1, // 几乎必然触发内存高负载
		l:                 logger.NopLogger{},
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		v.monitorLoad(ctx)
		close(done)
	}()
	time.Sleep(50 * time.Millisecond)
	assert.True(t, v.highLoad.Load())
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("monitorLoad did not stop after cancel")
	}
}

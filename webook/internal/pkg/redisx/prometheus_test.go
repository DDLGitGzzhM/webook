package redisx

import (
	"context"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPrometheusHook_ProcessHook(t *testing.T) {
	registry := prometheus.NewRegistry()
	old := prometheus.DefaultRegisterer
	prometheus.DefaultRegisterer = registry
	t.Cleanup(func() {
		prometheus.DefaultRegisterer = old
	})

	hook := NewPrometheusHook(prometheus.SummaryOpts{
		Namespace: "test",
		Subsystem: "redis",
		Name:      "cmd",
		Help:      "test",
	})

	called := false
	process := hook.ProcessHook(func(ctx context.Context, cmd redis.Cmder) error {
		called = true
		return redis.Nil
	})

	cmd := redis.NewStringCmd(context.Background(), "get", "k")
	err := process(context.Background(), cmd)
	assert.ErrorIs(t, err, redis.Nil)
	assert.True(t, called)

	families, err := registry.Gather()
	require.NoError(t, err)
	require.NotEmpty(t, families)
	assert.Equal(t, "test_redis_cmd", families[0].GetName())
}

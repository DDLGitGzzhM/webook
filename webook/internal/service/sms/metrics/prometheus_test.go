package metrics

import (
	"context"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubSMS struct {
	called bool
	biz    string
}

func (s *stubSMS) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	s.called = true
	s.biz = biz
	return nil
}

func TestNewPrometheusDecorator_Send(t *testing.T) {
	registry := prometheus.NewRegistry()
	old := prometheus.DefaultRegisterer
	prometheus.DefaultRegisterer = registry
	t.Cleanup(func() {
		prometheus.DefaultRegisterer = old
	})

	stub := &stubSMS{}
	decorator := NewPrometheusDecorator(stub)
	err := decorator.Send(context.Background(), "login", []string{"123456"}, "13800138000")
	require.NoError(t, err)
	assert.True(t, stub.called)
	assert.Equal(t, "login", stub.biz)

	families, err := registry.Gather()
	require.NoError(t, err)
	require.NotEmpty(t, families)
	assert.Equal(t, "geekbang_daming_webook_sms_resp_time", families[0].GetName())
}

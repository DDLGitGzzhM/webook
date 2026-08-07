package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtendFields_Get(t *testing.T) {
	ext := ExtendFields{
		"uid": "123",
	}
	val, err := ext.Get("uid").AsInt64()
	require.NoError(t, err)
	assert.Equal(t, int64(123), val)

	missing := ext.Get("missing")
	assert.True(t, errors.Is(missing.Err, errKeyNotFound))
	assert.Equal(t, int64(0), missing.Int64OrDefault(0))
}

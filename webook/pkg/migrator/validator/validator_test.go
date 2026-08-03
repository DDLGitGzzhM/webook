package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

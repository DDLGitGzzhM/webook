package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"webook/webook/account/domain"
)

type fakeAccountRepo struct {
	checkErr error
	addErr   error
	setErr   error
	checked  int
	added    int
	set      int
}

func (f *fakeAccountRepo) AddCredit(ctx context.Context, c domain.Credit) error {
	f.added++
	return f.addErr
}

func (f *fakeAccountRepo) CheckUnique(ctx context.Context, c domain.Credit) error {
	f.checked++
	return f.checkErr
}

func (f *fakeAccountRepo) SetUnique(ctx context.Context, c domain.Credit) error {
	f.set++
	return f.setErr
}

func TestAccountService_Credit(t *testing.T) {
	t.Run("重复记账直接返回", func(t *testing.T) {
		repo := &fakeAccountRepo{checkErr: errors.New("该业务已经处理过了")}
		svc := NewAccountService(repo)
		err := svc.Credit(context.Background(), domain.Credit{Biz: "reward", BizId: 1})
		assert.Error(t, err)
		assert.Equal(t, 1, repo.checked)
		assert.Equal(t, 0, repo.added)
	})

	t.Run("成功记账并写入幂等标记", func(t *testing.T) {
		repo := &fakeAccountRepo{}
		svc := NewAccountService(repo)
		err := svc.Credit(context.Background(), domain.Credit{Biz: "reward", BizId: 1})
		require.NoError(t, err)
		assert.Equal(t, 1, repo.checked)
		assert.Equal(t, 1, repo.added)
		assert.Equal(t, 1, repo.set)
	})

	t.Run("记账失败不写幂等标记", func(t *testing.T) {
		repo := &fakeAccountRepo{addErr: errors.New("db error")}
		svc := NewAccountService(repo)
		err := svc.Credit(context.Background(), domain.Credit{Biz: "reward", BizId: 1})
		assert.Error(t, err)
		assert.Equal(t, 1, repo.added)
		assert.Equal(t, 0, repo.set)
	})
}

package connpool

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
)

// WriteSplit 主从模式
type WriteSplit struct {
	master gorm.ConnPool
	slaves []gorm.ConnPool
}

func (w *WriteSplit) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return w.master.(gorm.TxBeginner).BeginTx(ctx, opts)
}

func (w *WriteSplit) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	// 可以默认返回 master，也可以默认返回 slave
	return w.master.PrepareContext(ctx, query)
}

func (w *WriteSplit) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return w.master.ExecContext(ctx, query, args...)
}

func (w *WriteSplit) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	// slaves 要考虑负载均衡, 搞个轮询 slaves
	// 这边可以玩骚操作，轮询，加权轮询，平滑的加权轮询，随机，加权随机
	// 动态判定 slaves 健康情况的负载均衡策略
	//（永远挑选最快返回响应的那个 slave，或者暂时禁用超时的 slaves）
	panic("implement me")
}

func (w *WriteSplit) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	//TODO implement me
	panic("implement me")
}

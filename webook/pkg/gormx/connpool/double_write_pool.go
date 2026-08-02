package connpool

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ecodeclub/ekit/syncx/atomicx"
	"gorm.io/gorm"
)

var errUnknownPattern = errors.New("未知的双写模式")

type DoubleWritePool struct {
	src     gorm.ConnPool
	dst     gorm.ConnPool
	pattern *atomicx.Value[string]
}

func NewDoubleWritePool(src gorm.ConnPool,
	dst gorm.ConnPool, pattern string) *DoubleWritePool {
	return &DoubleWritePool{src: src, dst: dst, pattern: atomicx.NewValueOf(pattern)}
}

func (d *DoubleWritePool) BeginTx(ctx context.Context, opts *sql.TxOptions) (gorm.ConnPool, error) {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		tx, err := d.src.(gorm.TxBeginner).BeginTx(ctx, opts)
		return &DoubleWritePoolTx{
			src:     tx,
			pattern: pattern,
		}, err
	case PatternSrcFirst:
		srcTx, err := d.src.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}
		dstTx, err := d.dst.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			// 记录日志，然后不做处理

			// 可以考虑回滚
			// err = srcTx.Rollback()
			// return err
		}
		return &DoubleWritePoolTx{
			src:     srcTx,
			dst:     dstTx,
			pattern: pattern,
		}, nil

	case PatternDstOnly:
		tx, err := d.dst.(gorm.TxBeginner).BeginTx(ctx, opts)
		return &DoubleWritePoolTx{
			src:     tx,
			pattern: pattern,
		}, err
	case PatternDstFirst:
		dstTx, err := d.dst.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}
		srcTx, err := d.src.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			// 记录日志，然后不做处理

			// 可以考虑回滚
			// err = dstTx.Rollback()
			// return err
		}
		return &DoubleWritePoolTx{
			src:     srcTx,
			dst:     dstTx,
			pattern: pattern,
		}, nil
	default:
		return nil, errors.New("未知的双写模式")
	}
}

// PrepareContext Prepare 的语句会进来这里
func (d *DoubleWritePool) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	// sql.Stmt 是一个结构体，你没有办法说返回一个代表双写的 Stmt
	panic("implement me")
	//return nil, errors.New("双写模式下不支持")
	//switch d.pattern.Load() {
	//case PatternSrcOnly, PatternSrcFirst:
	//	return d.src.PrepareContext(ctx, query)
	//case PatternDstOnly, PatternDstFirst:
	//	return d.dst.PrepareContext(ctx, query)
	//default:
	//	panic("未知的双写模式")
	//	//return nil, errors.New("未知的双写模式")
	//}
}

// 在增量校验的时候，我能不能利用这个方法？
// 1.1 我能不能从 query 里面抽取出来主键， WHERE id= xxx ，然后我就知道哪些数据被影响了？
// 1.2 可以尝试的思路是：用抽象语法树来分析 query， 而后找出 query 里面的条件，执行一个 SELECT，判定有哪些 id
// 1.2.1 UPDATE xx set b = xx WHERE a = 1
// 1.2.2 UPDATE xx set a = xx WHERE a = 1 LIMIT 10; DELETE from xxx WHERE aa OFFSET abc LIMIT cde
// 1.2.3 INSERT INTO ON CONFLICT, upsert 语句
func (d *DoubleWritePool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	switch d.pattern.Load() {
	case PatternSrcOnly:
		return d.src.ExecContext(ctx, query, args...)
	case PatternSrcFirst:
		res, err := d.src.ExecContext(ctx, query, args...)
		if err != nil {
			return res, err
		}
		_, err = d.dst.ExecContext(ctx, query, args...)
		if err != nil {
			// 记日志
			// dst 写失败，不被认为是失败
		}
		return res, err
	case PatternDstOnly:
		return d.dst.ExecContext(ctx, query, args...)
	case PatternDstFirst:
		res, err := d.dst.ExecContext(ctx, query, args...)
		if err != nil {
			return res, err
		}
		_, err = d.src.ExecContext(ctx, query, args...)
		if err != nil {
			// 记日志
			// dst 写失败，不被认为是失败
		}
		return res, err
	default:
		panic("未知的双写模式")
		//return nil, errors.New("未知的双写模式")
	}
}

func (d *DoubleWritePool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	switch d.pattern.Load() {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryContext(ctx, query, args...)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.QueryContext(ctx, query, args...)
	default:
		panic("未知的双写模式")
		//return nil, errors.New("未知的双写模式")
	}
}

func (d *DoubleWritePool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	switch d.pattern.Load() {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryRowContext(ctx, query, args...)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.QueryRowContext(ctx, query, args...)
	default:
		// 这里有一个问题，我怎么返回一个 error
		// unsafe 可以
		panic("未知的双写模式")
	}
}

func (d *DoubleWritePool) UpdatePattern(pattern string) {
	d.pattern.Store(pattern)
	// 我能不能，有事务未提交的情况下，我禁止修改
	// 能，但是性能问题比较严重，你需要维持住一个已开事务的计数，要用锁了
}

type DoubleWritePoolTx struct {
	src     *sql.Tx
	dst     *sql.Tx
	pattern string
}

// Commit 和 PPT 不一致
func (d *DoubleWritePoolTx) Commit() error {
	switch d.pattern {
	case PatternSrcOnly:
		return d.src.Commit()
	case PatternSrcFirst:
		// 源库上的事务失败了，我目标库要不要提交
		// commit 失败了怎么办？
		err := d.src.Commit()
		if err != nil {
			// 要不要提交？
			return err
		}
		if d.dst != nil {
			err = d.dst.Commit()
			if err != nil {
				// 记录日志
			}
		}
		return nil
	case PatternDstOnly:
		return d.dst.Commit()
	case PatternDstFirst:
		err := d.dst.Commit()
		if err != nil {
			// 要不要提交？
			return err
		}
		if d.src != nil {
			err = d.src.Commit()
			if err != nil {
				// 记录日志
			}
		}
		return nil
	default:
		return errUnknownPattern
	}
}

func (d *DoubleWritePoolTx) Rollback() error {
	switch d.pattern {
	case PatternSrcOnly:
		return d.src.Rollback()
	case PatternSrcFirst:
		// 源库上的事务失败了，我目标库要不要提交
		// commit 失败嘞怎么办？
		err := d.src.Rollback()
		if err != nil {
			// 要不要提交？
			// 我个人觉得 可以尝试 rollback
			return err
		}
		if d.dst != nil {
			err = d.dst.Rollback()
			if err != nil {
				// 记录日志
			}
		}
		return nil
	case PatternDstOnly:
		return d.dst.Rollback()
	case PatternDstFirst:
		err := d.dst.Rollback()
		if err != nil {
			// 要不要提交？
			return err
		}
		if d.src != nil {
			err = d.src.Rollback()
			if err != nil {
				// 记录日志
			}
		}
		return nil
	default:
		return errUnknownPattern
	}
}

func (d *DoubleWritePoolTx) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	panic("implement me")
}

func (d *DoubleWritePoolTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	switch d.pattern {
	case PatternSrcOnly:
		return d.src.ExecContext(ctx, query, args...)
	case PatternSrcFirst:
		res, err := d.src.ExecContext(ctx, query, args...)
		if err != nil {
			return res, err
		}
		if d.dst == nil {
			return res, err
		}
		_, err = d.dst.ExecContext(ctx, query, args...)
		if err != nil {
			// 记日志
			// dst 写失败，不被认为是失败
		}
		return res, err
	case PatternDstOnly:
		return d.dst.ExecContext(ctx, query, args...)
	case PatternDstFirst:
		res, err := d.dst.ExecContext(ctx, query, args...)
		if err != nil {
			return res, err
		}
		if d.src == nil {
			return res, err
		}
		_, err = d.src.ExecContext(ctx, query, args...)
		if err != nil {
			// 记日志
			// dst 写失败，不被认为是失败
		}
		return res, err
	default:
		panic("未知的双写模式")
		//return nil, errors.New("未知的双写模式")
	}
}

func (d *DoubleWritePoolTx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	switch d.pattern {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryContext(ctx, query, args...)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.QueryContext(ctx, query, args...)
	default:
		panic("未知的双写模式")
		//return nil, errors.New("未知的双写模式")
	}
}

func (d *DoubleWritePoolTx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	switch d.pattern {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryRowContext(ctx, query, args...)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.QueryRowContext(ctx, query, args...)
	default:
		panic("未知的双写模式")
		//return nil, errors.New("未知的双写模式")
	}
}

const (
	PatternDstOnly  = "DST_ONLY"
	PatternSrcOnly  = "SRC_ONLY"
	PatternDstFirst = "DST_FIRST"
	PatternSrcFirst = "SRC_FIRST"
)

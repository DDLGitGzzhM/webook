package validator

import (
	"context"
	"database/sql"
	"runtime"
	"strconv"
	"time"

	"github.com/ecodeclub/ekit/slice"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"

	"webook/webook/internal/pkg/logger"
	"webook/webook/pkg/migrator"
	events2 "webook/webook/pkg/migrator/events"
)

const (
	defaultThreadsRunningThreshold = 50
	defaultPoolBusyRatio           = 0.8
	defaultHeapAllocThreshold      = 512 << 20 // 512MB
	defaultLoadCheckInterval       = time.Second
	defaultHighLoadBackoff         = time.Second
)

// Validator T 必须实现了 Entity 接口
type Validator[T migrator.Entity] struct {
	baseValidator

	batchSize int

	highLoad *atomicx.Value[bool]
	// 高负载时挂起多久再重新检查
	highLoadBackoff time.Duration
	// 多久采样一次负载
	loadCheckInterval time.Duration

	// MySQL Threads_running 超过该值视为 DB 高负载
	threadsRunningThreshold int
	// 连接池 InUse/MaxOpen 超过该比例视为高负载
	poolBusyRatio float64
	// 本进程 HeapAlloc 超过该值视为内存高负载
	heapAllocThreshold uint64

	// 在这里加字段，比如说，在查询 base 根据什么列来排序，在 target 的时候，根据什么列来查询数据
	// 最极端的情况，是这样

	utime int64
	// <=0 说明直接退出校验循环
	// > 0 真的 sleep
	sleepInterval time.Duration

	fromBase func(ctx context.Context, offset int) (T, error)
}

func NewValidator[T migrator.Entity](
	base *gorm.DB,
	target *gorm.DB,
	direction string,
	l logger.Logger,
	p events2.Producer) *Validator[T] {
	res := &Validator[T]{
		baseValidator: baseValidator{
			base:      base,
			target:    target,
			direction: direction,
			l:         l,
			producer:  p,
		},
		batchSize:               100,
		highLoad:                atomicx.NewValueOf[bool](false),
		highLoadBackoff:         defaultHighLoadBackoff,
		loadCheckInterval:       defaultLoadCheckInterval,
		threadsRunningThreshold: defaultThreadsRunningThreshold,
		poolBusyRatio:           defaultPoolBusyRatio,
		heapAllocThreshold:      defaultHeapAllocThreshold,
	}
	res.fromBase = res.fullFromBase
	return res
}

func (v *Validator[T]) SleepInterval(i time.Duration) *Validator[T] {
	v.sleepInterval = i
	return v
}

func (v *Validator[T]) Utime(utime int64) *Validator[T] {
	v.utime = utime
	return v
}

func (v *Validator[T]) Incr() *Validator[T] {
	v.fromBase = v.intrFromBase
	return v
}

func (v *Validator[T]) Validate(ctx context.Context) error {
	// 负载监控跟随校验生命周期：ctx 取消后停止，避免 New 时起永不退出的 goroutine
	go v.monitorLoad(ctx)

	var eg errgroup.Group
	eg.Go(func() error {
		v.validateBaseToTarget(ctx)
		return nil
	})

	eg.Go(func() error {
		v.validateTargetToBase(ctx)
		return nil
	})
	return eg.Wait()
}

// monitorLoad 周期性采样 DB / 本机负载，写入 highLoad。
// 高负载是可逆状态：恢复后校验会继续，不会永久卡住。
func (v *Validator[T]) monitorLoad(ctx context.Context) {
	ticker := time.NewTicker(v.loadCheckInterval)
	defer ticker.Stop()
	v.refreshHighLoad()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			v.refreshHighLoad()
		}
	}
}

func (v *Validator[T]) refreshHighLoad() {
	busy := v.isDBHighLoad(v.base) ||
		v.isDBHighLoad(v.target) ||
		isMemHighLoad(runtimeHeapAlloc(), v.heapAllocThreshold)
	prev := v.highLoad.Load()
	v.highLoad.Store(busy)
	if busy && !prev {
		v.l.Warn("校验进入高负载挂起")
	}
	if !busy && prev {
		v.l.Info("校验负载恢复，继续执行")
	}
}

func (v *Validator[T]) isDBHighLoad(db *gorm.DB) bool {
	if db == nil {
		return false
	}
	sqlDB, err := db.DB()
	if err != nil {
		return false
	}
	if isPoolHighLoad(sqlDB.Stats(), v.poolBusyRatio) {
		return true
	}
	threads, ok := queryThreadsRunning(db)
	if !ok {
		return false
	}
	return isThreadsHighLoad(threads, v.threadsRunningThreshold)
}

// waitIfHighLoad 在高负载时挂起；ctx 取消返回 true，表示调用方应退出。
func (v *Validator[T]) waitIfHighLoad(ctx context.Context) bool {
	for v.highLoad.Load() {
		select {
		case <-ctx.Done():
			return true
		case <-time.After(v.highLoadBackoff):
		}
	}
	return false
}

func isPoolHighLoad(stats sql.DBStats, ratio float64) bool {
	if stats.MaxOpenConnections <= 0 || ratio <= 0 {
		return false
	}
	return float64(stats.InUse)/float64(stats.MaxOpenConnections) >= ratio
}

func isThreadsHighLoad(threads, threshold int) bool {
	return threshold > 0 && threads >= threshold
}

func isMemHighLoad(heapAlloc, threshold uint64) bool {
	return threshold > 0 && heapAlloc >= threshold
}

func runtimeHeapAlloc() uint64 {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return ms.HeapAlloc
}

type mysqlStatusRow struct {
	VariableName string `gorm:"column:Variable_name"`
	Value        string `gorm:"column:Value"`
}

// queryThreadsRunning 查询 MySQL Threads_running；非 MySQL 或失败时 ok=false。
func queryThreadsRunning(db *gorm.DB) (int, bool) {
	var row mysqlStatusRow
	err := db.Raw("SHOW GLOBAL STATUS LIKE 'Threads_running'").Scan(&row).Error
	if err != nil || row.Value == "" {
		return 0, false
	}
	n, err := strconv.Atoi(row.Value)
	if err != nil {
		return 0, false
	}
	return n, true
}

// validateBaseToTargetV1 批量写法，第十三周答案
func (v *Validator[T]) validateBaseToTargetV1(ctx context.Context) error {
	offset := 0
	// 假设说一次 100 条，复用已有的 batchSize
	for {
		var srcs []T
		// 直接取出来一批
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		err := v.base.WithContext(dbCtx).
			Order("id").
			Where("utime >= ?", v.utime).
			Offset(offset).Limit(v.batchSize).Find(&srcs).Error
		cancel()
		switch err {
		// 在 find 里面其实不会有这个错误
		//case gorm.ErrRecordNotFound:
		case context.Canceled, context.DeadlineExceeded:
			// 超时你可以继续，也可以返回。一般超时都是因为数据库有了问题
			return err
		case nil:
			if len(srcs) == 0 {
				// 结束，没有数据
				return nil
			}
			err = v.dstDiffV1(ctx, srcs)
			if err != nil {
				// 直接中断，你也可以考虑继续重试
				return err
			}
		default:
			v.l.Error("校验数据，查询 base 出错", logger.Error(err.Error()))
		}
		if len(srcs) < v.batchSize {
			// 没有数据了
			return nil
		}
		offset += len(srcs)
	}
}

func (v *Validator[T]) dstDiffV1(ctx context.Context, srcs []T) error {
	dbCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	ids := slice.Map(srcs, func(idx int, src T) int64 {
		return src.ID()
	})
	var dsts []T
	err := v.target.WithContext(dbCtx).Where("id IN ?", ids).
		Find(&dsts).Error
	// 让调用者来决定
	if err != nil {
		return err
	}
	dstMap := v.toMap(dsts)
	for _, src := range srcs {
		dst, ok := dstMap[src.ID()]
		if !ok {
			v.notify(src.ID(), events2.InconsistentEventTypeTargetMissing)
			continue
		}
		if !src.CompareTo(dst) {
			v.notify(src.ID(), events2.InconsistentEventTypeNEQ)
		}
	}
	return nil
}

func (v *Validator[T]) toMap(data []T) map[int64]T {
	res := make(map[int64]T, len(data))
	for _, val := range data {
		res[val.ID()] = val
	}
	return res
}

// Validate 调用者可以通过 ctx 来控制校验程序退出
// 全量校验，是不是一条条比对？
// 所以要从数据库里面一条条查询出来
// utime 上面至少要有一个索引，并且 utime 必须是第一列
// <utime, col1, col2>, <utime> 这种可以
// <col1, utime> 这种就不可以
func (v *Validator[T]) validateBaseToTarget(ctx context.Context) {
	offset := 0
	for {
		if v.waitIfHighLoad(ctx) {
			return
		}

		// 找到了 base 中的数据
		// 例如 .Order("id DESC")，每次插入数据，就会导致你的 offset 不准了
		// 如果我的表没有 id 这个列怎么办？
		// 找一个类似的列，比如说 ctime (创建时间）
		// 作业。你改成批量，性能要好很多
		src, err := v.fromBase(ctx, offset)
		switch err {
		case context.Canceled, context.DeadlineExceeded:
			// 超时或者被人取消了
			return
		case nil:
			// 你真的查到了数据
			// 要去 target 里面找对应的数据
			var dst T
			err = v.target.Where("id = ?", src.ID()).First(&dst).Error
			switch err {
			case context.Canceled, context.DeadlineExceeded:
				// 超时或者被人取消了
				return
			case nil:
				if !src.CompareTo(dst) {
					// 不相等
					// 这时候，我要干嘛？上报给 Kafka，就是告知数据不一致
					v.notify(src.ID(),
						events2.InconsistentEventTypeNEQ)
				}

			case gorm.ErrRecordNotFound:
				// 这意味着，target 里面少了数据
				v.notify(src.ID(),
					events2.InconsistentEventTypeTargetMissing)
			default:
				// 这里，要不要汇报，数据不一致？
				// 你有两种做法：
				// 1. 我认为，大概率数据是一致的，我记录一下日志，下一条
				v.l.Error("查询 target 数据失败", logger.Error(err.Error()))
				// 2. 我认为，出于保险起见，我应该报数据不一致，试着去修一下
				// 如果真的不一致了，没事，修它
				// 如果假的不一致（也就是数据一致），也没事，就是多余修了一次
				// 不好用哪个 InconsistentType
			}

		case gorm.ErrRecordNotFound:
			// 比完了。没数据了，全量校验结束了
			// 同时支持全量校验和增量校验，你这里就不能直接返回
			// 在这里，你要考虑：有些情况下，用户希望退出，有些情况下。用户希望继续
			// 当用户希望继续的时候，你要 sleep 一下
			if v.sleepInterval <= 0 {
				return
			}
			time.Sleep(v.sleepInterval)
			continue
		default:
			// 数据库错误
			v.l.Error("校验数据，查询 base 出错",
				logger.Error(err.Error()))
			// 课堂演示方便，你可以删掉
			time.Sleep(time.Second)
			// offset 最好是挪一下
			// 这里要不要挪
		}
		offset++
	}
}

func (v *Validator[T]) fullFromBase(ctx context.Context, offset int) (T, error) {
	dbCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	var src T
	err := v.base.WithContext(dbCtx).
		Offset(offset).
		Order("id").First(&src).Error
	return src, err
}

func (v *Validator[T]) intrFromBase(ctx context.Context, offset int) (T, error) {
	dbCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	var src T
	err := v.base.WithContext(dbCtx).
		Where("utime > ?", v.utime).
		Offset(offset).
		Order("utime ASC, id ASC").First(&src).Error
	return src, err
}

// 理论上来说，可以利用 count 来加速这个过程，
// 我举个例子，假如说你初始化目标表的数据是 昨天的 23:59:59 导出来的
// 那么你可以 COUNT(*) WHERE ctime < 今天的零点，count 如果相等，就说明没删除
// 这一步大多数情况下效果很好，尤其是那些软删除的。
// 如果 count 不一致，那么接下来，你理论上来说，还可以分段 count
// 比如说，我先 count 第一个月的数据，一旦有数据删除了，你还得一条条查出来
func (v *Validator[T]) validateTargetToBase(ctx context.Context) {
	// 先找 target，再找 base，找出 base 中已经被删除的
	// 理论上来说，就是 target 里面一条条找
	offset := 0
	for {
		if v.waitIfHighLoad(ctx) {
			return
		}
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)

		var dstTs []T
		err := v.target.WithContext(dbCtx).
			Where("utime > ?", v.utime).
			Select("id").
			Offset(offset).Limit(v.batchSize).
			Order("utime").Find(&dstTs).Error
		cancel()
		if len(dstTs) == 0 {
			// 没数据了。直接返回
			if v.sleepInterval <= 0 {
				return
			}
			time.Sleep(v.sleepInterval)
			continue
		}
		switch err {
		case context.Canceled, context.DeadlineExceeded:
			// 超时或者被人取消了
			return
		// 正常来说，gorm 在 Find 方法接收的是切片的时候，不会返回 gorm.ErrRecordNotFound
		case gorm.ErrRecordNotFound:
			// 没数据了。直接返回
			if v.sleepInterval <= 0 {
				return
			}
			time.Sleep(v.sleepInterval)
			continue
		case nil:
			ids := slice.Map(dstTs, func(idx int, t T) int64 {
				return t.ID()
			})
			// 可以直接用 NOT IN
			var srcTs []T
			err = v.base.Where("id IN ?", ids).Find(&srcTs).Error
			switch err {
			case context.Canceled, context.DeadlineExceeded:
				// 超时或者被人取消了
				return
			case gorm.ErrRecordNotFound:
				v.notifyBaseMissing(ctx, ids)
			case nil:
				srcIds := slice.Map(srcTs, func(idx int, t T) int64 {
					return t.ID()
				})
				// 计算差集
				// 也就是，src 里面的咩有的
				diff := slice.DiffSet(ids, srcIds)
				v.notifyBaseMissing(ctx, diff)
			default:
				// 记录日志
			}
		default:
			// 记录日志，continue 掉
			v.l.Error("查询target 失败", logger.Error(err.Error()))
		}
		offset += len(dstTs)
		if len(dstTs) < v.batchSize {
			if v.sleepInterval <= 0 {
				return
			}
			time.Sleep(v.sleepInterval)
		}
	}
}

func (v *Validator[T]) notifyBaseMissing(_ context.Context, ids []int64) {
	for _, id := range ids {
		v.notify(id, events2.InconsistentEventTypeBaseMissing)
	}
}

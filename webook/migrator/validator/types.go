package validator

import (
	"context"
	"time"

	"github.com/ecodeclub/ekit/slice"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"gorm.io/gorm"

	"webook/webook/internal/pkg/logger"
	"webook/webook/migrator"
	"webook/webook/migrator/events"
)

// Validator T 必须实现了 Entity 接口
type Validator[T migrator.Entity] struct {
	// 校验，以 XXX 为准，
	base *gorm.DB
	// 校验的是谁的数据
	target *gorm.DB
	l      logger.Logger

	p events.Producer

	direction string

	batchSize int

	highLoad *atomicx.Value[bool]

	// 在这里加字段，比如说，在查询 base 根据什么列来排序，在 target 的时候，根据什么列来查询数据
	// 最极端的情况，是这样

}

func NewValidator[T migrator.Entity](
	base *gorm.DB,
	target *gorm.DB,
	l logger.Logger,
	p events.Producer,
	direction string,
) *Validator[T] {
	highLoad := atomicx.NewValueOf[bool](false)
	go func() {
		// 在这里，去查询数据库的状态
		// 你的校验代码不太可能是性能瓶颈，性能瓶颈一般在数据库
		// 你也可以结合本地的 CPU，内存负载来判定
	}()
	return &Validator[T]{
		base:      base,
		target:    target,
		l:         l,
		p:         p,
		direction: direction,
		batchSize: 100,
		highLoad:  highLoad,
	}
}

func (v *Validator[T]) Validate(ctx context.Context) {
	v.validateBaseToTarget(ctx)
	v.validateTargetToBase(ctx)
}

// Validate 调用者可以通过 ctx 来控制校验程序退出
// 全量校验，是不是一条条比对？
// 所以要从数据库里面一条条查询出来
func (v *Validator[T]) validateBaseToTarget(ctx context.Context) {
	offset := -1
	for {
		//
		if v.highLoad.Load() {
			// 挂起
		}
		// 进来就更新 offset，比较好控制
		// 因为后面有很多的 continue 和 return
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		offset++
		var src T
		// 找到了 base 中的数据
		// 例如 .Order("id DESC")，每次插入数据，就会导致你的 offset 不准了
		// 如果我的表没有 id 这个列怎么办？
		// 找一个类似的列，比如说 ctime (创建时间）
		err := v.base.WithContext(dbCtx).Offset(offset).
			Order("id").First(&src).Error
		cancel()
		switch err {
		case nil:
			// 你真的查到了数据
			// 要去 target 里面找对应的数据
			var dst T
			err = v.target.Where("id = ?", src.ID()).First(&dst).Error
			// 我在这里，怎么办？
			switch err {
			case nil:
				// 找到了。你要开始比较
				// 你怎么比较？
				// 能不能这么比？
				// 1. src == dst
				// 这是利用反射来比
				// 这个原则上是可以的。
				//if reflect.DeepEqual(src, dst) {
				//
				//}
				//var srcAny any = src
				//if c1, ok := srcAny.(interface {
				//	// 有没有自定义的比较逻辑
				//	CompareTo(c2 migrator.Entity) bool
				//}); ok {
				//	// 有，我就用它的
				//	if !c1.CompareTo(dst) {
				//
				//	}
				//} else {
				//	// 没有，我就用反射
				//	if !reflect.DeepEqual(src, dst) {
				//
				//	}
				//}
				if !src.CompareTo(dst) {
					// 不相等
					// 这时候，我要干嘛？上报给 Kafka，就是告知数据不一致
					v.notify(ctx, src.ID(),
						events.InconsistentEventTypeNEQ)
				}

			case gorm.ErrRecordNotFound:
				// 这意味着，target 里面少了数据
				v.notify(ctx, src.ID(),
					events.InconsistentEventTypeTargetMissing)
			default:
				// 这里，要不要汇报，数据不一致？
				// 你有两种做法：
				// 1. 我认为，大概率数据是一致的，我记录一下日志，下一条
				v.l.Error("查询 target 数据失败", logger.Error(err.Error()))
				continue
				// 2. 我认为，出于保险起见，我应该报数据不一致，试着去修一下
				// 如果真的不一致了，没事，修它
				// 如果假的不一致（也就是数据一致），也没事，就是多余修了一次
				// 不好用哪个 InconsistentType
			}
		case gorm.ErrRecordNotFound:
			// 比完了。没数据了，全量校验结束了
			return
		default:
			// 数据库错误
			v.l.Error("校验数据，查询 base 出错",
				logger.Error(err.Error()))
			// offset 最好是挪一下

			continue

		}
	}
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
	offset := -v.batchSize
	for {
		offset = offset + v.batchSize
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)

		var dstTs []T
		err := v.target.WithContext(dbCtx).
			Select("id").
			// WHERE 条件二分查找 COUNT
			Offset(offset).Limit(v.batchSize).
			Order("id").Find(&dstTs).Error
		cancel()
		if len(dstTs) == 0 {
			return
		}
		switch err {
		case gorm.ErrRecordNotFound:
			// 没数据了。直接返回
			return
		case nil:
			ids := slice.Map(dstTs, func(idx int, t T) int64 {
				return t.ID()
			})
			// 可以直接用 NOT IN
			var srcTs []T
			err = v.base.Where("id IN ?", ids).Find(&srcTs).Error
			switch err {
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
			// 全没有
			default:
				// 记录日志
				continue
			}
		default:
			// 记录日志，continue 掉
			continue
		}
		if len(dstTs) < v.batchSize {
			// 没数据了
			return
		}
	}
}

func (v *Validator[T]) notifyBaseMissing(ctx context.Context, ids []int64) {
	for _, id := range ids {
		v.notify(ctx, id, events.InconsistentEventTypeBaseMissing)
	}
}

func (v *Validator[T]) notify(ctx context.Context, id int64, typ string) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	err := v.p.ProduceInconsistentEvent(ctx,
		events.InconsistentEvent{
			ID:        id,
			Direction: v.direction,
			Type:      typ,
		})
	cancel()
	if err != nil {
		// 这又是一个问题
		// 怎么办？
		// 你可以重试，但是重试也会失败，记日志，告警，手动去修
		// 我直接忽略，下一轮修复和校验又会找出来
		v.l.Error("发送数据不一致的消息失败", logger.Error(err.Error()))
	}
}

package wrr

import (
	"sync"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

const name = "custom_wrr"

// balancer.Balancer 接口
// balancer.Builder 接口
// balancer.Picker 接口
// base.PickerBuilder 接口
// 你可以认为，Balancer 是 Picker 的装饰器
func init() {
	// NewBalancerBuilder 是帮我们把一个 Picker Builder 转化为一个 balancer.Builder
	balancer.Register(base.NewBalancerBuilder(name,
		&PickerBuilder{}, base.Config{HealthCheck: false}))
}

// 传统版本的基于权重的负载均衡算法

type PickerBuilder struct {
}

func (p *PickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]*conn, 0, len(info.ReadySCs))
	// sc => SubConn
	// sci => SubConnInfo
	for sc, sci := range info.ReadySCs {
		cc := &conn{
			cc: sc,
		}
		md, ok := sci.Address.Metadata.(map[string]any)
		if ok {
			weightVal := md["weight"]
			weight, _ := weightVal.(float64)
			cc.weight = int(weight)
		}

		if cc.weight == 0 {
			// 可以给个默认值
			cc.weight = 10
		}
		cc.currentWeight = cc.weight
		conns = append(conns, cc)
	}
	return &Picker{
		conns: conns,
	}
}

type Picker struct {
	//	 这个才是真的执行负载均衡的地方
	conns []*conn
	mutex sync.Mutex
}

// Pick 在这里实现基于权重的负载均衡算法
func (p *Picker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if len(p.conns) == 0 {
		// 没有候选节点
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	var total int
	var maxCC *conn
	// 要计算当前权重
	for _, cc := range p.conns {
		// 性能最好就是在 cc 上用原子操作
		// 但是筛选结果不会严格符合 WRR 算法
		// 整体效果可以
		//cc.lock.Lock()
		total += cc.weight
		cc.currentWeight = cc.currentWeight + cc.weight
		if maxCC == nil || cc.currentWeight > maxCC.currentWeight {
			maxCC = cc
		}
		//cc.lock.Unlock()
	}

	// 更新
	maxCC.currentWeight = maxCC.currentWeight - total
	// maxCC 就是挑出来的
	return balancer.PickResult{
		SubConn: maxCC.cc,
		Done: func(info balancer.DoneInfo) {
			// 很多动态算法，根据调用结果来调整权重，就在这里
		},
	}, nil
}

// conn 代表节点
type conn struct {
	// 权重
	weight        int
	currentWeight int

	//lock sync.Mutex

	//	真正的，grpc 里面的代表一个节点的表达
	cc balancer.SubConn
}

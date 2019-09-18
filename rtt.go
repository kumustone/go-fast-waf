package panda_waf

import (
	"sync/atomic"
)

/*
	RTT round-trip time
*/

var (
	rtt RTT
)

//RTT统计的基础类
type RTT struct {
	name  string // 统计类目名称
	count uint64 // 总个数
	total uint64 // 总耗时
}

func (r *RTT) Add(time uint64) {
	atomic.AddUint64(&r.total, time)
	atomic.AddUint64(&r.count, 1)
}

func (r *RTT) GetAverageRT() uint64 {
	average := uint64(0)
	total := atomic.SwapUint64(&r.total, 0)
	count := atomic.SwapUint64(&r.count, 0)

	if count > 0 {
		average = total / count
	} else {
		average = 0
	}

	return average
}

package share

import (
	"fmt"
	"sync"
	"time"
)

/*
做调试使用，对整个数据流进行打点，以便找到可能的阻塞点；
*/

// 跟踪msg打点
type timeStamp struct {
	time uint64 // 打点时间，单位是us
	name string // 打点名称
}

func newTimeStamp(name string) *timeStamp {
	return &timeStamp{
		time: uint64(time.Now().UnixNano()) / uint64(time.Microsecond),
		name: name,
	}
}

type MsgTrace struct {
	trace []*timeStamp // 流经的每个地点的时间打标；
}

func (mt *MsgTrace) mark(name string) {
	mt.trace = append(mt.trace, &timeStamp{
		time: uint64(time.Now().UnixNano()) / uint64(time.Microsecond),
		name: name,
	})
}

func NewMsgTrace() *MsgTrace {
	mt := &MsgTrace{}
	mt.mark("beginning")
	return mt
}

// 设置消息打点, 单独跟踪一条消息，默认为是线程安全的
func (mt *MsgTrace) MarkTimeStamp(name string) {
	mt.trace = append(mt.trace, &timeStamp{
		time: uint64(time.Now().UnixNano()) / uint64(time.Microsecond),
		name: name,
	})
}

// 打点时间输出
func (mt *MsgTrace) OutputString() string {
	if len(mt.trace) == 0 {
		return "null"
	}

	var output string
	for index, value := range mt.trace {
		if index == 0 {
			continue
		}
		output = output + fmt.Sprintf("->%s(%dus)", value.name, int64(value.time)-int64(mt.trace[0].time))
	}
	return output
}

const sessionMapNum = 8

// key: requestid, value: 同步等待返回的wait
type requestMap struct {
	rwmutex  sync.RWMutex
	sessions map[uint64]*MsgTrace
}

type MsgTraceCache struct {
	sessionMaps [sessionMapNum]requestMap
	//disposeFlag bool
	//disposeOnce sync.Once
	//disposeWait sync.WaitGroup
	current uint64
}

func NewMsgTraceCache() *MsgTraceCache {
	manager := &MsgTraceCache{}
	for i := 0; i < sessionMapNum; i++ {
		manager.sessionMaps[i].sessions = make(map[uint64]*MsgTrace)
	}
	return manager
}

var G_msg_trace = NewMsgTraceCache()

func (m *MsgTraceCache) Cache(request_id uint64, mt *MsgTrace) {
	smap := &m.sessionMaps[request_id%sessionMapNum]
	smap.rwmutex.Lock()
	smap.sessions[request_id] = mt
	smap.rwmutex.Unlock()
}

func (m *MsgTraceCache) Get(request_id uint64) *MsgTrace {
	smap := &m.sessionMaps[request_id%sessionMapNum]

	smap.rwmutex.RLock()
	defer smap.rwmutex.RUnlock()

	mt, exist := smap.sessions[request_id]
	if exist {
		return mt
	}
	return nil
}

func (m *MsgTraceCache) Remove(request_id uint64) {
	smap := &m.sessionMaps[request_id%sessionMapNum]
	smap.rwmutex.Lock()
	defer smap.rwmutex.Unlock()
	_, exist := smap.sessions[request_id]
	if exist {
		//直接把channel关闭掉
		delete(smap.sessions, request_id)
	}
}

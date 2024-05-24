package gate

import (
	"sync"
	"sync/atomic"
	"time"
)

// 目前只支持轮询模式，一般后面是NGINX web服务；
// 一般网关不能重启，支持对后端数据的热加载；
type RouterItem struct {
	Key   string
	Value interface{}
}

type Notify struct {
	Items  []*RouterItem
	Action int
}

var (
	Upstream = NewRouter()
)

type Router struct {
	notify  chan Notify
	pool    []*RouterItem
	stop    chan struct{}
	counter int32
	current uint32 //轮询算法使用
	mutex   sync.RWMutex
	timeout time.Duration
}

func NewRouter() *Router {
	return &Router{
		notify: make(chan Notify, 128),
		stop:   make(chan struct{}),
	}
}

func (r *Router) WaitNotify() {
	go func() {
		for {
			select {
			case n := <-r.notify:
				for _, item := range n.Items {
					if n.Action == 0 {
						r.Add(item)
					} else if n.Action == 1 {
						r.Remove(item.Key)
					}
				}
			case <-r.stop:
				return
			}
		}
	}()
}

func (r *Router) Size() int32 {
	return atomic.LoadInt32(&r.counter)
}

func (r *Router) addSize(d int32) int32 {
	return atomic.AddInt32(&r.counter, d)
}

// 做健康监测使用
func (r *Router) Remove(key string) {
	exist := false
	index := 0
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for i, value := range r.pool {
		if value.Key == key {
			exist = true
			index = i
			break
		}
	}
	if exist {
		r.pool = append(r.pool[:index], r.pool[index+1:]...)
		r.addSize(-1)
	}

}

func (r *Router) Add(item *RouterItem) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	exist := false
	for _, value := range r.pool {
		if value.Key == item.Key {
			exist = true
			return false
		}
	}

	if !exist {
		r.pool = append(r.pool, item)
		r.addSize(1)
	}
	return true
}

func (r *Router) Select() *RouterItem {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if r.Size() <= 0 {
		return nil
	}

	//类型是uint32, 如果int32发生越界
	r.current++
	index := r.current % uint32(r.Size())

	return r.pool[index]
}

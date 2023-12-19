package waf

import (
	"log"
	"sync"
	"time"
)

type CCStat struct {
	count int
}

// CCRule 单个CCRules
type CCRule struct {
	Host       string
	Interval   int
	Count      int
	ForbidTime int
	Key        string

	ttl int
	//到时间后，把整个statics清理掉；
	statics map[string]*CCStat
	mutex   sync.RWMutex
}

type CCServe struct {
	mutex sync.RWMutex
	rules map[string][]*CCRule
}

func NewCCRule(j *JsonCCRule) *CCRule {
	return &CCRule{
		Host:       j.Host,
		Interval:   j.InterVal,
		Count:      j.Count,
		ForbidTime: j.ForbidTime,
		Key:        j.Key,
		ttl:        Now() + j.InterVal,
		statics:    make(map[string]*CCStat),
	}
}

func (r *CCRule) OnReq(req *WafHttpRequest) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	//目前Key只支持IP
	key := req.RemoteAddr
	value, exist := r.statics[key]
	if !exist {
		r.statics[key] = &CCStat{
			count: 1,
		}

	} else {
		value.count = value.count + 1
		if value.count >= r.Count {
			log.Printf("AntiCC add to blacklist: %s:%s\n ", r.Host, key)
			CBlackList.Add(&BlackInfo{
				Host:    r.Host,
				Key:     key,
				AddTime: Now(),
				EndTime: Now() + r.ForbidTime,
			})
			delete(r.statics, key)
		}
	}
	return
}

func (r *CCRule) CleanUp(now int) {
	if now < r.ttl {
		return
	}
	r.ttl = now + r.Interval

	r.mutex.Lock()
	r.statics = make(map[string]*CCStat)
	r.mutex.Unlock()
}

func NewCCServe() *CCServe {
	c := &CCServe{
		rules: make(map[string][]*CCRule),
	}

	go c.CleanLoop()
	return c
}

func (c *CCServe) HandleRule(j *JSONRule) {
	if len(j.CCRule.Host) == 0 {
		return
	}
	if j.Status != "valid" {
		return
	}

	c.mutex.Lock()
	value, exist := c.rules[j.CCRule.Host]
	if exist == false {
		var ruleList []*CCRule
		cr := NewCCRule(&(j.CCRule))
		ruleList = append(ruleList, cr)
		c.rules[j.CCRule.Host] = ruleList
	} else {
		cr := NewCCRule(&(j.CCRule))
		value = append(value, cr)
		c.rules[j.CCRule.Host] = value
	}
	c.mutex.Unlock()

	log.Println("add rule ", *j)
}

func (c *CCServe) CleanRules() {
	c.mutex.Lock()
	c.rules = make(map[string][]*CCRule)
	c.mutex.Unlock()
}

func (c *CCServe) CleanLoop() {
	c.mutex.RLock()
	now := Now()
	for _, rules := range c.rules {
		for _, spiderRule := range rules {
			spiderRule.CleanUp(now)
		}
	}
	c.mutex.RUnlock()

	time.AfterFunc(time.Second, c.CleanLoop)
}

func (c *CCServe) Add(j *JsonCCRule) {
	if len(j.Host) == 0 {
		return
	}

	c.mutex.Lock()
	value, exist := c.rules[j.Host]
	if exist == false {
		var ruleList []*CCRule
		cr := NewCCRule(j)
		ruleList = append(ruleList, cr)
		c.rules[j.Host] = ruleList
	} else {
		cr := NewCCRule(j)
		value = append(value, cr)
		c.rules[j.Host] = value
	}
	c.mutex.Unlock()

	log.Println("AntiCC add rule ", *j)
}

func (c *CCServe) CheckRequest(req *WafHttpRequest) *WafProxyResp {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	host := req.Host

	hostRules, exist := c.rules[host]
	if exist {
		for _, rule := range hostRules {
			rule.OnReq(req)
		}
	}

	return SuccessResp
}

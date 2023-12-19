package waf

import (
	"log"
	"sync"
	"time"
)

// BlackInfo 命中多条规则，那么以最长的时间为准；
type BlackInfo struct {
	Host    string
	Key     string
	AddTime int
	EndTime int
}

type BlackMap struct {
	sessions map[string]*BlackInfo
}

// CacheBlackList 第一层以Host作为区分，第二层的Key值可能是IP或是UID
type CacheBlackList struct {
	sync.RWMutex
	blackMaps map[string]BlackMap
	current   uint64
}

func NewCacheBlackList() *CacheBlackList {
	manager := &CacheBlackList{
		blackMaps: make(map[string]BlackMap),
	}

	go manager.CleanLoop()
	return manager
}

func (c *CacheBlackList) CheckRequest(req *WafHttpRequest) *WafProxyResp {
	if b := c.Match(req.Mark, req.RemoteAddr); b != nil {
		return &WafProxyResp{
			RetCode:  WAF_INTERCEPT,
			RuleName: "CacheBlackList",
			Desc:     "缓存黑名单",
		}
	}
	return SuccessResp
}

func (c *CacheBlackList) HandleRule(j *JSONRule) {

}

func (c *CacheBlackList) CleanRules() {

}

func (c *CacheBlackList) Add(info *BlackInfo) {
	c.Lock()
	defer c.Unlock()

	blackMap, exist := c.blackMaps[info.Host]
	if !exist {
		blackMap = BlackMap{
			sessions: make(map[string]*BlackInfo),
		}
		c.blackMaps[info.Host] = blackMap
	}

	//不管有没有这个IP的blackInfo，都将它覆盖掉；
	blackMap.sessions[info.Key] = info
}

func (c *CacheBlackList) Remove(host string, key string) {
	c.Lock()
	defer c.Unlock()
	blackMap, exist := c.blackMaps[host]
	if !exist {
		return
	}
	delete(blackMap.sessions, key)
}

func (c *CacheBlackList) Match(host string, key string) *BlackInfo {
	c.RLock()
	defer c.RUnlock()

	blackMap, exist := c.blackMaps[host]
	if !exist {
		return nil
	}

	if b, exist := blackMap.sessions[key]; exist {
		return b
	}

	return nil
}

func (c *CacheBlackList) CleanLoop() {
	for {
		c.Lock()
		now := Now()
		for key, value := range c.blackMaps {
			for key1, value1 := range value.sessions {
				if value1.EndTime <= now {
					log.Printf("AntiCC Remove timeout key : %s:%s\n", key, key1)
					delete(value.sessions, key1)
				}
			}
		}
		c.Unlock()
		time.Sleep(time.Second * 1)
	}
}

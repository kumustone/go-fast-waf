package main

import (
	. "go-fast-waf/share"
	"sync"
)

type IPList struct {
	set   map[string]bool
	name  string
	desc  string
	mutex sync.RWMutex
}

func NewIPList(n string, d string) *IPList {
	return &IPList{
		name: n,
		desc: d,
		set:  make(map[string]bool),
	}
}

func (l *IPList) HandleRule(r *JSONRule) {
	for _, i := range r.IPList {
		if r.Status == "valid" {
			l.Add(i)
		}
	}
}

func (l *IPList) CleanRules() {
	l.mutex.Lock()
	l.set = make(map[string]bool)
	l.mutex.Unlock()
}

func (l *IPList) CheckRequest(req *WafHttpRequest) *WafProxyResp {
	ip := req.RemoteAddr
	if l.Contains(ip) {

		resp := &WafProxyResp{
			RetCode:  WAF_INTERCEPT,
			RuleName: l.name,
			Desc:     l.desc,
		}

		if l.name == "IPWhiteList" {
			resp.RetCode = WAF_PASS
		}

		return resp
	}

	return SuccessResp
}

func (l *IPList) Add(k string) {
	l.mutex.Lock()
	l.set[k] = true
	l.mutex.Unlock()
}

func (l *IPList) Remove(k string) {
	l.mutex.Lock()
	delete(l.set, k)
	l.mutex.Unlock()
}

func (l *IPList) Contains(k string) bool {
	l.mutex.RLock()
	_, ok := l.set[k]
	l.mutex.RUnlock()
	return ok
}

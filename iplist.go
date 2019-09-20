package waf

import (
	"sync"
)

type IpSet struct {
	set   map[string]bool
	name  string
	desc  string
	mutex sync.RWMutex
}

//var BlackList IpSet
//var WriteList IpSet

func NewIpSet(name string, desc string) *IpSet {
	return &IpSet{
		name: name,
		desc: desc,
		set:  make(map[string]bool),
	}
}

func (s *IpSet) HandleRule(r *JsonRule) {
	for _, i := range r.IPList {
		if r.Action == "add" {
			s.Add(i)
		} else if r.Action == "remove" {
			s.Remove(i)
		}
	}
}

func (s *IpSet) CheckRequest(req *WafHttpRequest) *WafProxyResp {
	ip := req.RemoteAddr
	if s.Contains(ip) {
		return &WafProxyResp{
			RetCode:  WAF_INTERCEPT,
			RuleName: s.name,
			Desc:     s.desc,
		}
	}

	return SuccessResp
}

func (s *IpSet) Add(k string) {
	s.mutex.Lock()
	s.set[k] = true
	s.mutex.Unlock()
}

func (s *IpSet) Remove(k string) {
	s.mutex.Lock()
	delete(s.set, k)
	s.mutex.Unlock()
}

func (s *IpSet) Contains(k string) bool {
	s.mutex.RLock()
	_, ok := s.set[k]
	s.mutex.RUnlock()
	return ok
}

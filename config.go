package waf

import (
	"sync"
)

// WafCheckList 通过URL的关键字来检查哪些需要上传检测，哪些不需要
type WafCheckList struct {
	// 检测项
	Include []string
	// 排除项
	Exclude []string
	// 默认url是否检测
	CheckDefault bool
}

// WafServerAddr 要连接的waf-server的地址
type WafServerAddr struct {
	Address []string
}

const (
	WAF_SERVER_ADD = iota
	WAF_SERVER_REMOVE
)

type AddrNotify struct {
	Address []string
	Action  int
}

var (
	ServerNotify = make(chan AddrNotify, 128)
)

type Config struct {
	CheckSwitch bool
	CheckList   WafCheckList
	ServerAddr  WafServerAddr
}

var (
	config     Config
	chkInclude map[string]bool
	chkExclude map[string]bool
	mutex      sync.RWMutex
)

// InitConfig 初始化或者发生动态的改变都调用这个接口
func InitConfig(c Config) {
	mutex.Lock()
	defer mutex.Unlock()

	server := config.ServerAddr.Address

	config = c
	chkInclude = make(map[string]bool)
	chkExclude = make(map[string]bool)
	for _, item := range config.CheckList.Include {
		chkInclude[item] = true
	}

	for _, item := range config.CheckList.Exclude {
		chkExclude[item] = true
	}

	add, remove := diffSlice(server, config.ServerAddr.Address)

	ServerNotify <- AddrNotify{
		Address: add,
		Action:  WAF_SERVER_ADD,
	}

	ServerNotify <- AddrNotify{
		Address: remove,
		Action:  WAF_SERVER_REMOVE,
	}
}

// NeedCheck 判断一个标记是否需要经过waf检查
func NeedCheck(mark string) bool {
	mutex.RLock()
	defer mutex.RUnlock()

	if !config.CheckSwitch {
		return false
	}

	if _, ok := chkInclude[mark]; ok {
		return true
	}

	if _, ok := chkExclude[mark]; ok {
		return false
	}
	return config.CheckList.CheckDefault
}

// 比较两个数组的异同点
func diffSlice(oldArrays []string, newArrays []string) (add []string, remove []string) {
	for _, itemOld := range oldArrays {
		exist := false
		for _, itemNew := range newArrays {
			if itemOld == itemNew {
				exist = true
				break
			}
		}
		//已经删除的
		if exist == false {
			remove = append(remove, itemOld)
		}
	}

	for _, itemNew := range newArrays {
		exist := false
		for _, itemOld := range oldArrays {
			if itemNew == itemOld {
				exist = true
				break
			}
		}

		if exist == false {
			add = append(add, itemNew)
		}
	}

	return add, remove
}

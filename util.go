package panda_waf

import (
	"os"
	"runtime/debug"
	"time"
	"log"
	"errors"
)

//获取当前的微秒时间
func GetMicroTime() uint64 {
	return uint64(time.Now().UnixNano()) / uint64(time.Microsecond)
}

func PanicRecovery(quit bool) {
	var err error
	if r := recover(); r != nil {
		switch x := r.(type) {
		case string:
			err = errors.New(x)
			break
		case error:
			err = x
			break
		default:
			err = errors.New("Unknown panic")
			break
		}
		debug.PrintStack()
		log.Println("Panic :", err.Error())

		if quit {
			os.Exit(101)
		}
	}
}

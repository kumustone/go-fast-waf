package main

import (
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/kumustone/panda-waf"
	"github.com/natefinch/lumberjack"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"
)

type GateConfig struct {
	GateHttpAddress  string
	GateHttpsAddress string
	GateAPIAddress   string
	UpstreamList     []string
}

type WafGateConfig struct {
	Gate   GateConfig
	WAFRPC panda_waf.Config
}

var (
	confFile = flag.String("c", "./conf/waf_gate.conf", "Config file")
	logPath  = flag.String("l", "./log", " log path")
)

func main() {
	flag.Parse()

	c := WafGateConfig{}
	if _, err := toml.DecodeFile(*confFile, &c); err != nil {
		log.Fatal("Can not decode config file ", err.Error())
		return
	}

	defer panda_waf.PanicRecovery(true)

	log.SetOutput(&lumberjack.Logger{
		Filename:   "waf_gate.log",
		MaxSize:    1,
		MaxBackups: 10,
		MaxAge:     30,
	})

	//go pprof检测
	go func() {
		log.Println(http.ListenAndServe("0.0.0.0:60060", nil))
	}()

	panda_waf.InitConfig(c.WAFRPC)
	panda_waf.WaitServerNotify()

	for _, it := range c.Gate.UpstreamList {
		panda_waf.UpStream.Add(&panda_waf.RouterItem{
			Key: it,
		})
	}

	panda_waf.UpStream.WaitNotify()

	server := &http.Server{
		Addr:           c.Gate.GateHttpAddress,
		IdleTimeout:    3 * time.Minute,
		ReadTimeout:    5 * time.Minute,
		WriteTimeout:   5 * time.Minute,
		MaxHeaderBytes: 20 * 1024 * 1024,
		Handler:        panda_waf.WafHandler,
	}

	log.Fatal(server.ListenAndServe())

}

package main

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/kumustone/waf"
	"github.com/natefinch/lumberjack"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"
)

type GateConfig struct {
	GateHttpAddress  string
	StartHttps       bool
	CertKeyList      [][]string
	GateHttpsAddress string
	GateAPIAddress   string
	UpstreamList     []string
}

type WafGateConfig struct {
	Gate   GateConfig
	WAFRPC waf.Config
}

var (
	confFile = flag.String("c", "./waf_gate.conf", "Config file")
	logPath  = flag.String("l", "./log", " log path")
)

func main() {
	flag.Parse()

	c := WafGateConfig{}
	if _, err := toml.DecodeFile(*confFile, &c); err != nil {
		log.Fatal("Can not decode config file ", err.Error())
		return
	}

	log.Println(c)

	defer waf.PanicRecovery(true)

	log.SetOutput(&lumberjack.Logger{
		Filename:   *logPath + "/waf_gate.log",
		MaxSize:    1,
		MaxBackups: 10,
		MaxAge:     30,
	})

	//go pprof检测
	go func() {
		log.Println(http.ListenAndServe("0.0.0.0:60060", nil))
	}()

	waf.InitConfig(c.WAFRPC)
	waf.WaitServerNotify()

	for _, it := range c.Gate.UpstreamList {
		waf.UpStream.Add(&waf.RouterItem{
			Key: it,
		})
	}

	waf.UpStream.WaitNotify()
	server := &http.Server{
		Addr:           c.Gate.GateHttpAddress,
		IdleTimeout:    3 * time.Minute,
		ReadTimeout:    5 * time.Minute,
		WriteTimeout:   5 * time.Minute,
		MaxHeaderBytes: 20 * 1024 * 1024,
		Handler:        waf.WafHandler,
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			fmt.Println("Listen and serve error ", err.Error())
		}
	}()

	if c.Gate.StartHttps {

		caData := make(map[string][]string)
		for _, item := range c.Gate.CertKeyList {
			if len(item) != 3 {
				continue
			}
			caData[item[0]] = []string{item[1], item[2]}
		}

		cfg := &tls.Config{
			GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
				data, ok := caData[info.ServerName]
				if !ok {
					return nil, errors.New("Cert Key is not exist")
				}
				cert, err := tls.LoadX509KeyPair(data[0], data[1])
				if err != nil {
					return nil, err
				}
				return &cert, nil
			},
		}

		server := http.Server{
			IdleTimeout:    3 * time.Minute,
			ReadTimeout:    5 * time.Minute,
			WriteTimeout:   5 * time.Minute,
			MaxHeaderBytes: 20 * 1024 * 1024,
			Handler:        waf.WafHandler,
			TLSConfig:      cfg,
		}
		fmt.Println("Https start at ", c.Gate.GateHttpsAddress)
		log.Fatal(server.ListenAndServeTLS("", ""))
	}

	select {}
}

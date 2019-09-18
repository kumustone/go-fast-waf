package panda_waf

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
)

var UpStream = NewRouter()

var httpReverse = NewMultipleHostReverseProxy()

func NewMultipleHostReverseProxy() *httputil.ReverseProxy {
	debugLog := log.New(os.Stdout, "[Debug]", log.Ldate|log.Ltime|log.Llongfile)

	return &httputil.ReverseProxy{
		ErrorLog: debugLog,

		//Modify request
		Director: func(req *http.Request) {
			req.URL.Scheme = "http"
			req.URL.Host = UpStream.Select().Key
			//log.Println("upstream host ", req.URL.Host)
		},

		//Modify response
		ModifyResponse: func(resp *http.Response) error {
			return nil
		},

		Transport: &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return http.ProxyFromEnvironment(req)
			},

			Dial: func(network, addr string) (net.Conn, error) {
				conn, err := (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).Dial(network, addr)
				if err != nil {
					println("Error during DIAL:", err.Error())
				}
				return conn, err
			},

			//must config MaxIdleConnsPerHost： connect: can't assign requested address
			MaxIdleConnsPerHost: 512,
			TLSHandshakeTimeout: 300 * time.Second,
			IdleConnTimeout:     120 * time.Second,
		},
	}
}

type HttpHandler struct{}

var (
	WafHandler = &HttpHandler{}
)

func (h *HttpHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	ret := Check(req)
	if ret != nil && ret.RetCode == WAF_INTERCEPT {
		log.Printf("Intercept : Rule %s %s %s\n", ret.RuleName, ret.Desc, req)
		return
	}

	//可以在这里加一些包头给后端做业务处理
	req.Header.Set("x-waf-scheme", req.URL.Scheme)
	httpReverse.ServeHTTP(resp, req)
}

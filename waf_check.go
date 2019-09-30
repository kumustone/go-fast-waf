package waf

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

func cloneHeader(h http.Header) http.Header {
	h2 := make(http.Header, len(h))
	for k, vv := range h {
		vv2 := make([]string, len(vv))
		copy(vv2, vv)
		h2[k] = vv2
	}
	return h2
}

const MaxLimitBody int64 = 100 * 1024

func GetBody(req *http.Request) []byte {
	if req.ContentLength > MaxLimitBody || req.ContentLength <= 0 {
		return []byte("")
	}

	var originBody []byte
	defer req.Body.Close()
	if body, err := ioutil.ReadAll(req.Body); err != nil {
		log.Println("Get body fail : ", err.Error())
		return body
	} else {
		originBody = make([]byte, req.ContentLength)
		copy(originBody, body)
		req.Body = ioutil.NopCloser(bytes.NewReader(body))
		return originBody
	}
}

func Check(req *http.Request) *WafProxyResp {
	wafReq := &WafHttpRequest{
		Mark:          req.Host,
		Method:        req.Method,
		Scheme:        req.URL.Scheme,
		Url:           req.RequestURI,
		Proto:         req.Proto,
		Host:          req.Host,
		RemoteAddr:    req.RemoteAddr,
		ContentLength: uint64(req.ContentLength),
		Header:        cloneHeader(req.Header),
		Body:          GetBody(req),
	}

	//只保留IP即可
	if s := strings.Split(req.RemoteAddr, ":"); len(s) > 0 {
		wafReq.RemoteAddr = s[0]
	}

	resp, err := WafCheck(wafReq, time.Duration(20*time.Millisecond))
	if err != nil {
		//log.Println("waf check : ", err.Error())
	}
	return resp
}

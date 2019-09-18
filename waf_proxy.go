package panda_waf

import (
	"errors"
	"github.com/kumustone/tcpstream"
	"time"
)

var routerServer = NewRouter()

func WaitServerNotify() {
	go func() {
		for {
			select {
			case n := <-ServerNotify:
				for _, addr := range n.Address {
					if n.Action == WAF_SERVER_ADD {
						routerServer.Add(&RouterItem{
							Key:   addr,
							Value: tcpstream.NewSyncClient(addr),
						})
					} else if n.Action == WAF_SERVER_REMOVE {
						routerServer.Remove(addr)
					}
				}
			}
		}
	}()
}

func WafCheck(request *WafHttpRequest, timeout time.Duration) (*WafProxyResp, error) {
	if !NeedCheck(request.Mark) {
		return nil, errors.New("Need no check")
	}

	buffer, err := request.MarshalJSON()
	if err != nil {
		return nil, errors.New(" request MarshalJSON fail")
	}

	var conn *tcpstream.SyncClient
	for i := 0; i < int(routerServer.Size()); i++ {
		if r := routerServer.Select(); r == nil {
			break
		} else {
			if r.Value.(*tcpstream.SyncClient).Stream.State == tcpstream.CONN_STATE_ESTAB {
				conn = r.Value.(*tcpstream.SyncClient)
			}
			r = nil
		}
	}

	if conn == nil {
		return nil, errors.New("No tcpstream available ")
	}

	respMsg, err := conn.Call(&tcpstream.Message{Body: buffer}, time.Duration(time.Millisecond*200))
	if err != nil {
		return nil, err
	}

	resp := &WafProxyResp{}
	if err := resp.UnmarshalJSON(respMsg.Body); err != nil {
		return nil, err
	}

	return resp, nil
}

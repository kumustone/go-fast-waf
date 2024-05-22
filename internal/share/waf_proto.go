package share

const (
	WAF_PASS = iota
	WAF_INTERCEPT
	WAF_SLIDER
	WAF_INTERNAL_TIMEOUT
	WAF_INTERNAL_SEND_ERR
	WAF_INTERNAL_RECV_ERR
	WAF_INTERNAL_REQUEST_INVALID
)

//easyjson:json
type WafHttpRequest struct {
	Mark          string
	Method        string
	Scheme        string
	Url           string
	Proto         string
	Host          string
	RemoteAddr    string
	ContentLength uint64
	Header        map[string][]string
	Body          []byte
}

//easyjson:json
type WafProxyResp struct {
	RetCode  int
	RuleName string
	Desc     string
}

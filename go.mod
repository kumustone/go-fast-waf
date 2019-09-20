module panda-waf

go 1.13

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/kumustone/tcpstream v1.0.0
	github.com/kumustone/waf v0.0.1
	github.com/mailru/easyjson v0.7.0
	github.com/natefinch/lumberjack v2.0.0+incompatible
)

replace github.com/kumustone/waf v0.0.1 => ./

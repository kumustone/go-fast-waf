module panda-waf

go 1.19

replace github.com/kumustone/waf v0.0.1 => ./

require (
	github.com/BurntSushi/toml v1.2.0
	github.com/kumustone/tcpstream v1.0.2
	github.com/kumustone/waf v0.0.1
	github.com/mailru/easyjson v0.7.7
	github.com/natefinch/lumberjack v2.0.0+incompatible
)

require (
	github.com/josharian/intern v1.0.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

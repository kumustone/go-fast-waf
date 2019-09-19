# panda-waf


## 1. 安装
    - 依赖 go1.12+
    - git clone https://github.com/kumustone/panda-waf.git
    - cd panda-waf 
    - ./build.sh
  
 安装后会生成两个bin文件waf-gate waf-server 放在$GOPATH/bin下；
 
## 2. 运行
 
 > waf-gate -c /YourConfigPath/waf_gate.conf -l /YourLogPath/log 
 
 > waf-server -c /YouConfigPaht/waf_server.conf -l /YourLogPath/log -r /YourRuleDir
 
## 3. 配置说明

waf-gate  

```
[gate]
#Gate Http Listen Address
GateHttpAddress = "0.0.0.0:80"

# Gate Https Listen Addresses
GateHttpsAddress = "0.0.0.0:443"

# Gat API Service
GateAPIAddress = "0.0.0.0:2081"

# Upstream Address， RoundBin
UpstreamList = [
    "1.1.1.1:80"
]

# Waf检测项
[wafrpc]

# 检测开关，如果为false，所有的内容都不发送waf-server监测，直接转发给upstream处理；
# true: 按照Checklist规则处理；
CheckSwitch = true
[wafrpc.CheckList]
    # 需要检测项Host
    Include = ["xxxxxxx"]

    # 排除检测项Host
    Exclude = []

    # Include 和 Exclude以外的Host，是否检测；
    CheckDefault = true
[wafrpc.ServerAddr]
    # waf-server的地址列表；如果是多个，如果有多个waf-server按照轮询策略转发请求；
    Address = ["127.0.0.1:8000"]
```

waf-server 配置

```
[server]
# WafServer的监听地址
WafServerAddress = "127.0.0.1:8000"

# WafServer接口地址
HttpAPIAddress = "127.0.0.1:8001"

```

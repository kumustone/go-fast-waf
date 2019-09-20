# waf

轻量级的Waf检测工具；

基于Go语言包含Http/Https反向代理的网关waf-gate和一个执行检测任务的waf-server组成；

![](https://github.com/kumustone/waf/blob/master/doc/waf-1.jpg)

**网络连接功能**

- waf_gate与waf_server之间通过tcp长链连接；
- waf_gate与waf_server之间采用多对多连接，gate通过轮询策略发送给server检测；
- waf_gate支持检测超时设置，如果网络或者其他异常导致请求没有及时回复，那么waf_gate自动放过；
- waf_gate启动自动重连功能，如果没有当前没有可用的waf-server请求，那么放过所有请求；

**规则功能**

- 支持IP黑白名单；

- 基于Host,Referer,Url,User-Agent,Content-Type 正则表达式的组合规则，各个字段支持为空；

- rule规则通过JSON格式配置，方便后续通过接口规则做扩展；

    比如拦截请求： Host: www.xxx.com； Refer 为空； User-Agent中包nmap关键字；

## 1. 安装

    - 依赖 go1.12+
    - git clone https://github.com/kumustone/waf.git
    - cd waf 
    - ./build.sh

 安装后会生成两个bin文件waf-gate waf-server

## 2. 运行

 > waf-gate -c /YourConfigPath/waf_gate.conf -l /YourLogPath/log 

 > waf-server -c /YouConfigPaht/waf_server.conf -l /YourLogPath/log -r /YourRuleDir

## 3. 配置说明

waf-gate  

```go
[gate]
#Gate Http Listen Address
GateHttpAddress = "0.0.0.0:80"

# Gate Https Listen Addresses
StartHttps = true
GateHttpsAddress = "0.0.0.0:443"

# 多个二级域名使用不同的key
CertKeyList = [
[
    "A.xxx.com",
    "A.pem",
    "A.key"
],
[
    "B.xxx.com",
    "B.pem",
    "B.key"
]
]

CertFile = "xxxxxxx.pem"
KeyFile  = "xxxxx.key"

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

```go
[server]
# WafServer的监听地址
WafServerAddress = "127.0.0.1:8000"

# WafServer接口地址
HttpAPIAddress = "127.0.0.1:8001"

```

## 4. 规则文件说明

ip 黑名单的例子
```
{
  "rules": [
    {
      "type": "IpBlackList",
      "action": "add",
      "rule_name": "IpBlackList-0",
      "iplist": [
        "1.1.1.1",
        "2.2.2.2"
      ]
    }
  ]
}
```

正则规则的例子url 满足正则：\\(?\\s*\\b(alert|prompt|confirm|console\\.log)\\s*\\)?\\s*(\\(|`) 都会被拦截；

```
{
  "rules": [
    {
      "type": "Group",
      "action": "add",
      "rule_name": "xss-1",
      "desc": "this is a test rule",
      "group_rule": [
        {
          "field": "Url",
          "op": "is",
          "empty": false,
          "val": "\\(?\\s*\\b(alert|prompt|confirm|console\\.log)\\s*\\)?\\s*(\\(|`)"
        }
      ]
    }
    ]
  }
```

每一个规则文件里面包含的都是一份完整的json，不同的规则文件，可以叠加使用；

目前支持的规则还比较少，欢迎添加一些规则**


# go-fast-waf

[English](README_en.md)

**轻量级，低延迟，高性能，易于配置，网关和Waf解耦的风险检测服务.**

![](https://github.com/kumustone/go-fast-waf/blob/main/doc/waf-1.jpg)

## 组件介绍
1. waf-gate：一个基于 Go 语言编写的极简 HTTP/HTTPS 反向代理。
2. waf-server：负责执行Waf检测，黑白名单，访问控制，反爬等风险检测任务执行。

**为什么 waf-gate和waf-server 分离**

基于业务场景的考虑前提，即优先保证网关稳定性，高性能和低延迟， waf检测次之。
1. 网关稳定性风险解耦， 主要是稳定性风险和延迟风险。
 * waf-server出现崩溃时，waf-gate和waf-server长连接断开。数据不再发送waf-server检测； 
 * waf-server出现延迟，waf-gate的client超时机制会保证会在最大超时时间内（默认20ms）把数据转发到upstream server上，不影响业务的正常运行。
 * waf-server通过会有计算量大操作和入库等行为，如果waf-server和waf-gate是同一个工程，部署在同一台机器上存在资源征用的问题。响应网关的响应速度。
2. 方便分布式部署扩展，可以部署1对多部署，弥补某些复杂检测业务的waf性能不足。

**waf-gate**

1. **轻量级**：waf-gate 是一个非常轻量级的 HTTP 代理。它在 Go 的反向代理库基础上进行了极少量的修改和功能扩展。最大程度保证waf-gate运行的稳定性。
2. **路由功能**：如果您的网站部署很简单，没有复杂的业务路由，waf-gate 可以直接将请求分发给下游的业务服务器。但如果路由较为复杂，waf-gate 会将请求转发给另一个代理（例如 NGINX）进行进一步的路由。无论哪种情况，waf-gate 对整个业务链条来说都是完全透明的。
3. **Web风险检测** ： waf-gate 将请求转发给 waf-server 进行检测，然后 waf-server 将检测结果返回给 waf-gate，以便拦截或允许请求继续。

**waf-server**

1. **通信方式**：waf-gate 与 waf-server 之间通过长连接的 TCP 通信，使用 tcpstream 库实现。
2. **多连接支持**：waf-gate 与 waf-server 之间建立多个连接，采用轮询策略将请求分发给 waf-server 进行检测。
3. **超时处理**：waf-gate 支持检测超时设置，如果网络或其他异常导致 waf-server 未能及时响应，waf-gate 会自动放行请求。
4. **自动重连**：如果当前没有可用的 waf-server 响应，waf-gate 会自动放行所有请求。


**waf_gate与waf_server的网络通信**，是通过[tcpstream](<https://github.com/kumustone/tcpstream>)的库来实现的

- waf_gate与waf_server之间通过tcp长链连接；
- waf_gate与waf_server之间采用多对多连接，gate通过轮询策略发送给server检测；
- waf_gate支持检测超时设置，如果网络或者其他异常导致请求没有及时回复，那么waf_gate自动放过；
- waf_gate启动自动重连功能，如果没有当前没有可用的waf-server请求，那么放过所有请求；

## 规则功能

- 支持IP黑白名单；

- 基于Host,Referer,Url,User-Agent,Content-Type 正则表达式的组合规则，各个字段支持为空；

- rule规则通过JSON格式配置，方便后续通过接口规则做扩展，和后续进行GUI开发；

    比如拦截请求： Host: www.xxx.com； Refer 为空； User-Agent中包nmap关键字；

## 1. 安装

```
    git clone https://github.com/kumustone/waf.git
    cd waf 
    ./build.sh
```

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

目前支持的规则还比较少，欢迎添加一些规则;


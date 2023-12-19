# waf
[简体中文](README.md)

A lightweight Waf detection tool;

Based on Go language, it consists of a gateway waf-gate with Http/Https reverse proxy and a waf-server that performs detection tasks;

![](https://github.com/deph)



waf_gate is a very lightweight httpproxy reverse proxy, which does a very small amount of encapsulation and functionality addition on the basis of go's own reverseproxy library, with very little performance loss. waf_gate itself has a simple routing distribution function, if the website deployment is very simple, there is no business routing, you can directly distribute to the next level of business server through waf-gate. If the routing is more complex, then waf_gate directly forwards to the next level of proxy (such as NGINX) for routing distribution; in this process, waf_gate is completely transparent to the entire business chain.

waf detection is done by forwarding waf_gate to waf_server, and then waf_server returns the detection result to waf_gate to do interception or release operation. The main reasons for not doing rules directly on waf_gate are based on the following considerations:

1. If the detection rules are too complex, especially in the case of containing a lot of regular expressions, CPU time consumption will be higher, affecting waf_gate's forwarding time, thus increasing the business time of the entire link;
2. waf detection may cache a lot of data, resulting in large memory, GC time is too long;
3. Some rules require data from multiple httpproxies to be aggregated and then processed; this way httpProxy is powerless;
4. In the actual application scenario, request and response data may need to be stored;



waf_gate's supported features:

- Lightweight, performance, RT loss is very small;
- Support rewriting request header, response header, response tail;



**waf_gate and waf_server network communication**, is implemented by [tcpstream](https://github.com/deph) library

- waf_gate and waf_server are connected by tcp long chain;
- waf_gate and waf_server use multiple-to-multiple connections, gate sends to server detection by polling strategy;
- waf_gate supports detection timeout setting, if the network or other abnormality causes the request to not reply in time, then waf_gate automatically let go;
- waf_gate starts automatic reconnection function, if there is no currently available waf-server request, then let go of all requests;

**Rule function**

- Support IP black and white list;

- Based on Host,Referer,Url,User-Agent,Content-Type regular expression combination rules, each field supports empty;

- rule rules are configured by JSON format, which is convenient for subsequent expansion of interface rules and GUI development;

    For example, intercept request: Host: www.xxx.com; Refer is empty; User-Agent contains nmap keyword;

## 1. Installation

```
git clone https://github.com/kumustone/waf.git
cd waf 
./build.sh

```


After installation, two bin files waf-gate waf-server will be generated

## 2. Run

 > waf-gate -c /YourConfigPath/waf_gate.conf -l /YourLogPath/log

 > waf-server -c /YouConfigPaht/waf_server.conf -l /YourLogPath/log -r /YourRuleDir

## 3. Configuration instructions

waf-gate

```go
[gate]
#Gate Http Listen Address
GateHttpAddress = "0.0.0.0:80"

# Gate Https Listen Addresses
StartHttps = true
GateHttpsAddress = "0.0.0.0:443"

# Multiple secondary domains use different keys
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

# Waf detection item
[wafrpc]

# Detection switch, if false, all content is not sent to waf-server detection, directly forwarded to upstream processing;
# true: Process according to Checklist rules;
CheckSwitch = true
[wafrpc.CheckList]
    # Host to be detected
    Include = ["xxxxxxx"]

    # Exclude detection item Host
    Exclude = []

    # Whether to detect Host outside Include and Exclude;
    CheckDefault = true
[wafrpc.ServerAddr]
    # waf-server address list; if there are multiple, if there are multiple waf-server, forward requests by polling strategy;
    Address = ["127.0.0.1:8000"]

    # WafServer interface address
    HttpAPIAddress = "127.0.0.1:8001"
```

##  4. Rule file description

ip blacklist example

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

regular rule example url meets regular: \(?\s*\b(alert|prompt|confirm|console\.log)\s*\)?\s*(\(|`) will be intercepted;

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

Each rule file contains a complete json, different rule files can be superimposed;
The rules currently supported are relatively few, welcome to add some rules;



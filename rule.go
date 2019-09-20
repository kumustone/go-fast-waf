package waf

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
)

var (
	supportField = []string{
		"Host",
		"Referer",
		"Url",
		"User-Agent",
		"Content-Type",
	}
)

var (
	SuccessResp = &WafProxyResp{
		RetCode: WAF_PASS,
	}
)

var (
	IPBlackList = NewIpSet("IPBlackList", "")
	IPWriteList = NewIpSet("IPWhiteList", "")
	GroupRule   = NewRuleList()

	CheckList = []RuleCheckHandler{
		IPWriteList,
		IPBlackList,
		GroupRule,
	}
)

type JsonGroupRule struct {
	Field string `json:"field"`
	Op    string `json:"op"`
	Empty bool   `json:"empty"`
	Val   string `json:"val"`
}

type JsonRule struct {
	Type     string          `json:"type"`
	Action   string          `json:"action"`
	RuleName string          `json:"rulename"`
	Desc     string          `json:"desc,omitempty"`
	IPList   []string        `json:"iplist,omitempty"`
	Rule     []JsonGroupRule `json:"group_rule,omitempty"`
}

type JsonRules struct {
	Rules []*JsonRule `json:"rules"`
}

type RuleCheckHandler interface {
	CheckRequest(req *WafHttpRequest) *WafProxyResp
}

func validIP4(ipAddress string) bool {
	ipAddress = strings.Trim(ipAddress, " ")

	re, _ := regexp.Compile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)
	if re.MatchString(ipAddress) {
		return true
	}
	return false
}

func validField(field string) bool {
	for _, item := range supportField {
		if item == field {
			return true
		}
	}
	return false
}

func handleJsonFile(file string) error {
	log.Println("handle rule file :", file)
	bs, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	var r JsonRules
	if err := json.Unmarshal(bs, &r); err != nil {
		return err
	}
	return HandleRule(&r)
}

func InitRulePath(path string) error {
	log.Println("InitRulePath:", path)

	if f, err := os.Stat(path); err != nil {
		return err
	} else {
		if f.IsDir() {
			files, _ := ioutil.ReadDir(path)
			for _, ff := range files {
				if !ff.IsDir() && strings.HasSuffix(ff.Name(), "json") {
					if err := handleJsonFile(path + "/" + ff.Name()); err != nil {
						return err
					}
				}
			}
		} else {
			return handleJsonFile(path)
		}
	}
	return nil
}

func HandleRule(j *JsonRules) error {
	for _, jsonRule := range j.Rules {
		if err := CheckRule(jsonRule); err != nil {
			log.Println("Error : CheckRule ", err.Error())
			return err
		}
	}

	for _, jsonRule := range j.Rules {
		if jsonRule.Type == "IpBlackList" {
			IPBlackList.HandleRule(jsonRule)
		} else if jsonRule.Type == "IPWhiteList" {
			IPWriteList.HandleRule(jsonRule)
		} else if jsonRule.Type == "Group" {
			GroupRule.HandleRule(jsonRule)
		}
	}

	return nil
}

func CheckRule(j *JsonRule) error {
	//检查ruleName是否已经存在
	//检查action : add/remove
	//IPList中的内容是否全部为IP，或者IP段？
	//检查groupRule每一个field字段是否支持

	if j.Action != "add" && j.Action != "remove" {
		return errors.New("exist invalid action, check your rule config")
	}

	if j.Type != "IpBlackList" && j.Type != "IpWhiteList" && j.Type != "Group" {
		return errors.New(fmt.Sprint("type", j.Type, " is not invalid"))
	}

	for _, ip := range j.IPList {
		if !validIP4(ip) {
			return errors.New(fmt.Sprint("ip ", ip, " is Invalid"))
		}
	}

	for _, rule := range j.Rule {
		if !validField(rule.Field) {
			return errors.New(fmt.Sprint("field ", rule.Field, " is not support"))
		}

		if rule.Op != "is" && rule.Op != "not" {
			return errors.New(fmt.Sprint("op ", rule.Op, " is invalid"))
		}

		//data, err := base64.StdEncoding.DecodeString(rule.Val)
		//if err != nil {
		//	return errors.New(fmt.Sprint("Val", rule.Val, " can not base64 Decode ", err.Error()))
		//}

		if _, err := regexp.Compile(string(rule.Val)); err != nil {
			return errors.New(fmt.Sprint("Val ", rule.Val, " can not ruleExp Compile ", err.Error()))
		}
	}
	return nil
}

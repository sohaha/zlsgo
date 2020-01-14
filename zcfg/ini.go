package zcfg

import (
	"github.com/sohaha/zlsgo/zstring"
	"net/url"
	"strings"
)

type IniSt []map[string]map[string]string

func Ini(cfgPath string) (conflist IniSt, err error) {
	var json []byte
	if u, err := url.Parse(cfgPath); err == nil && u.Host != "" {
		json, _ = GetRemoteCfgContent(cfgPath)
	} else {
		json, _ = GetCfgContent(cfgPath)
	}
	conflist = readIni(json)
	return
}

func readIni(content []byte) (conflist IniSt) {

	var data map[string]map[string]string
	var section string
	l := strings.Split(zstring.Bytes2String(content), "\n")
	for _, line := range l {
		line := strings.TrimSpace(line)
		switch {
		case len(line) == 0:
		case string(line[0]) == "#": // 增加配置文件备注
		case line[0] == '[' && line[len(line)-1] == ']':
			section = strings.TrimSpace(line[1 : len(line)-1])
			data = make(map[string]map[string]string)
			data[section] = make(map[string]string)
		default:
			i := strings.IndexAny(line, "=")
			if i == -1 {
				continue
			}
			value := strings.TrimSpace(line[i+1 : len(line)])
			data[section][strings.TrimSpace(line[0:i])] = value
			for _, v := range conflist {
				for k, _ := range v {
					if k == section {
						continue
					}
				}
			}
			conflist = append(conflist, data)
		}
	}
	return conflist
}

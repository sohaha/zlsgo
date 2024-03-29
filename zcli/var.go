package zcli

import (
	"flag"
	"os"
	"strings"

	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/ztype"
)

type (
	cmdCont struct {
		command       Cmd
		name          string
		desc          string
		Supplement    string
		requiredFlags []string
	}
	runFunc func()
	// Cmd represents a subCommand
	Cmd interface {
		Flags(subcommand *Subcommand)
		Run(args []string)
	}
	errWrite struct {
	}
	v struct {
		name   string
		usage  string
		shorts []string
	}
	// Subcommand sub command
	Subcommand struct {
		CommandLine *flag.FlagSet
		Name        string
		Desc        string
		Supplement  string
		Parameter   string
		cmdCont
	}
)

const cliPrefix = ""

var (
	// BuildTime Build Time
	BuildTime = ""
	// BuildGoVersion Build Go Version
	BuildGoVersion = ""
	// BuildGitCommitID Build Git CommitID
	BuildGitCommitID = ""
	// Log cli logger
	Log *zlog.Logger
	// FirstParameter First Parameter
	FirstParameter   = os.Args[0]
	flagHelp         = new(bool)
	flagDetach       = new(bool)
	flagVersion      = new(bool)
	osExit           = os.Exit
	cmds             = make(map[string]*cmdCont)
	cmdsKey          []string
	matchingCmd      *cmdCont
	args             []string
	requiredFlags    = make([]string, 0)
	defaultLang      = "en"
	unknownCommandFn = func(name string) {
		Error("unknown Command: %s", errorText(name))
	}
	Logo         string
	Name         string
	Version      string
	HideHelp     bool
	EnableDetach bool
	HidePrompt   bool
	Lang         = defaultLang
	varsKey      = map[string]*v{}
	varShortsKey = make([]string, 0)
	ShortValues  = map[string]interface{}{}
	langs        = map[string]map[string]string{
		"en": {
			"command_empty": "Command name cannot be empty",
			"help":          "Show Command help",
			"version":       "View version",
			"detach":        "Running in the background",
			"test":          "Test",
			"restart":       "Restart service",
			"stop":          "Stop service",
			"start":         "Start service",
			"status":        "Service status",
			"uninstall":     "Uninstall service",
			"install":       "Install service",
		},
		"zh": {
			"command_empty": "命令名不能为空",
			"help":          "显示帮助信息",
			"version":       "查看版本信息",
			"detach":        "后台运行",
			"restart":       "重启服务",
			"stop":          "停止服务",
			"start":         "开始服务",
			"status":        "服务状态",
			"uninstall":     "卸载服务",
			"install":       "安装服务",
		},
	}
)

func SetLangText(lang, key, value string) {
	l, ok := langs[lang]
	if !ok {
		l = map[string]string{}
	}
	l[key] = value
	langs[lang] = l
}

func GetLangText(key string, def ...string) string {
	if lang, ok := langs[Lang][key]; ok {
		return lang
	}

	if lang, ok := langs[defaultLang][key]; ok {
		return lang
	}
	if len(def) > 0 {
		return def[0]
	}
	return key
}

func (e *errWrite) Write(p []byte) (n int, err error) {
	Error(strings.Replace(ztype.ToString(p), cliPrefix, "", 1))
	return 1, nil
}

func SetVar(name, usage string) *v {
	v := &v{
		name:  cliPrefix + name,
		usage: usage,
	}
	varsKey[name] = v
	return v
}

func (v *v) short(short string) *v {
	v.shorts = append(v.shorts, short)
	// todo prevent duplicate addition
	varShortsKey = append(varShortsKey, short)
	return v
}

// Required Set flag to be required
func (v *v) Required() *v {
	if matchingCmd != nil {
		matchingCmd.requiredFlags = append(matchingCmd.requiredFlags, v.name)
	} else {
		requiredFlags = append(requiredFlags, v.name)
	}
	return v
}

func (v *v) String(def ...string) *string {
	var value string
	if len(def) > 0 {
		value = def[0]
	}
	v.setFlagbind(func(name string) {
		ShortValues[name] = setFlags(v, value, func() interface{} {
			return flag.String(name, value, v.usage)
		}).(*string)
	})
	return setFlags(v, value, func() interface{} {
		return flag.String(v.name, value, v.usage)
	}).(*string)
}

func (v *v) Int(def ...int) *int {
	var value int
	if len(def) > 0 {
		value = def[0]
	}
	v.setFlagbind(func(name string) {
		ShortValues[name] = setFlags(v, value, func() interface{} {
			return flag.Int(name, value, v.usage)
		}).(*int)
	})
	return setFlags(v, value, func() interface{} {
		return flag.Int(v.name, value, v.usage)
	}).(*int)
}

func (v *v) Bool(def ...bool) *bool {
	var value bool
	if len(def) > 0 {
		value = def[0]
	}
	v.setFlagbind(func(name string) {
		ShortValues[name] = setFlags(v, value, func() interface{} {
			return flag.Bool(name, value, v.usage)
		}).(*bool)
	})
	return setFlags(v, value, func() interface{} {
		return flag.Bool(v.name, value, v.usage)
	}).(*bool)
}

var flags = map[string]interface{}{}

func setFlags(v *v, value interface{}, fn func() interface{}) (p interface{}) {
	p, ok := flags[v.name]
	if !ok {
		flags[v.name] = fn()
		return flags[v.name]
	}

	switch val := value.(type) {
	case bool:
		b, ok := p.(*bool)
		if !ok {
			Error("flag %s type error, it needs to be an bool", v.name)
		}
		*b = val
		return b
	case string:
		s, ok := p.(*string)
		if !ok {
			Error("flag %s type error, it needs to be an string", v.name)
		}
		*s = val
		return s
	case int:
		i, ok := p.(*int)
		if !ok {
			Error("flag %s type error, it needs to be an int", v.name)
		}
		*i = val
		return i
	}
	return nil
}

func (v *v) setFlagbind(fn func(name string)) {
	for _, s := range v.shorts {
		fn(s)
	}
}

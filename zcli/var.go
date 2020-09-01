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
		name          string
		desc          string
		command       Cmd
		requiredFlags RequiredFlags
		Supplement    string
	}
	runFunc func()
	// RequiredFlags RequiredFlags flags
	RequiredFlags []string
	// Cmd represents a subCommand
	Cmd interface {
		Flags(subcommand *Subcommand)
		Run(args []string)
	}
	errWrite struct {
	}
	Var struct {
		name   string
		usage  string
		shorts []string
	}
	Subcommand struct {
		cmdCont
		CommandLine *flag.FlagSet
		Name        string
		Desc        string
		Supplement  string
		Parameter   string
	}
)

const cliPrefix = ""

var (
	BuildTime        = ""
	BuildGoVersion   = ""
	BuildGitCommitID = ""
	// Log cli logger
	Log              *zlog.Logger
	FirstParameter   = os.Args[0]
	flagHelp         = new(bool)
	flagVersion      = new(bool)
	osExit           = os.Exit
	cmds             = make(map[string]*cmdCont)
	cmdsKey          []string
	matchingCmd      *cmdCont
	args             []string
	requiredFlags    = RequiredFlags{}
	defaultLang      = "en"
	unknownCommandFn = func(name string) {
		Error("unknown Command: %s", errorText(name))
	}
	Logo         string
	Name         string
	Version      string
	HideHelp     bool
	HidePrompt   bool
	Lang         = defaultLang
	varsKey      = map[string]*Var{}
	varShortsKey = make([]string, 0)
	ShortValues  = map[string]interface{}{}
	langs        = map[string]map[string]string{
		"en": {
			"command_empty": "Command name cannot be empty",
			"help":          "Show Command help",
			"version":       "View version",
			"test":          "Test",
			"restart":       "Restart service",
			"stop":          "Stop service",
			"start":         "Start service",
			"status":        "ServiceIfe status",
			"uninstall":     "Uninstall service",
			"install":       "Install service",
		},
		"zh": {
			"command_empty": "命令名不能为空",
			"help":          "显示帮助信息",
			"version":       "查看版本信息",
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

func SetVar(name, usage string) *Var {
	v := &Var{
		name:  cliPrefix + name,
		usage: usage,
	}
	varsKey[name] = v
	return v
}

func (v *Var) short(short string) *Var {
	v.shorts = append(v.shorts, short)
	// todo 防止重复添加
	varShortsKey = append(varShortsKey, short)
	return v
}

// Required Set flag to be required
func (v *Var) Required() *Var {
	if matchingCmd != nil {
		matchingCmd.requiredFlags = append(matchingCmd.requiredFlags, v.name)
	} else {
		requiredFlags = append(requiredFlags, v.name)
	}
	return v
}

func (v *Var) String(def ...string) *string {
	var value string
	if len(def) > 0 {
		value = def[0]
	}
	v.setFlagbind(func(name string) {
		ShortValues[name] = flag.String(name, value, v.usage)
	})
	return flag.String(v.name, value, v.usage)
}

func (v *Var) Int(def ...int) *int {
	var value int
	if len(def) > 0 {
		value = def[0]
	}
	v.setFlagbind(func(name string) {
		ShortValues[name] = flag.Int(name, value, v.usage)
	})
	return flag.Int(v.name, value, v.usage)
}

func (v *Var) Bool(def ...bool) *bool {
	var value bool
	if len(def) > 0 {
		value = def[0]
	}
	v.setFlagbind(func(name string) {
		ShortValues[name] = flag.Bool(name, value, v.usage)
	})
	return flag.Bool(v.name, value, v.usage)
}

func (v *Var) setFlagbind(fn func(name string)) {
	for _, s := range v.shorts {
		fn(s)
	}
}

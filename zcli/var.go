package zcli

import (
	"flag"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/ztype"
	"os"
	"strings"
)

const cliPrefix = ""

var (
	BuildTime      = ""
	BuildGoVersion = ""
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
	// appConfig      = &App{
	// Lang: defaultLang,
	// }
	Logo         string
	Description  string
	Version      string
	HideHelp     bool
	HidePrompt   bool
	Lang         = defaultLang
	varsKey      = map[string]*Var{}
	varShortsKey = make([]string, 0)
	ShortValues  = map[string]interface{}{}
)

type (
	// App struct {
	// Logo       string
	// Version    string
	// HideHelp   bool
	// HidePrompt bool
	// Lang       string
	// }
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
		id int
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

func getLangs(key string) string {
	// lang := appConfig.Lang
	// if lang == "" {
	// lang = defaultLang
	// }
	langs := map[string]map[string]string{
		"en": {
			"command_empty": "Command name cannot be empty",
			"help":          "Show Command help",
			"version":       "View version",
			"test":          "Test",
		},
		"zh": {
			"command_empty": "命令名不能为空",
			"help":          "显示帮助信息",
			"version":       "查看版本信息",
		},
	}
	if lang, ok := langs[Lang][key]; ok {
		return lang
	}

	if lang, ok := langs[defaultLang][key]; ok {
		return lang
	}

	return ""
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

func (v *Var) Short(short string) *Var {
	v.shorts = append(v.shorts, short)
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
	shortLen := len(v.shorts)
	if shortLen > 0 {
		for i := 0; i < shortLen; i++ {
			fn(v.shorts[i])
		}
	}
}

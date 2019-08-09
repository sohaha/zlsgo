package zcli

import (
	"flag"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/ztype"
	"os"
)

var (
	// Log cli logger
	Log            *zlog.Logger
	firstParameter = os.Args[0]
	flagHelp       = new(bool)
	flagVersion    = new(bool)
	osExit         = os.Exit
	cmds           = make(map[string]*cmdCont)
	matchingCmd    *cmdCont
	args           []string
	requiredFlags  = RequiredFlags{}
	defaultLang    = "en"
	appConfig      = &App{
		Lang: defaultLang,
	}
)

type (
	App struct {
		Logo       string
		Version    string
		HideHelp   bool
		HidePrompt bool
		Lang       string
	}
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
		name  string
		usage string
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
	lang := appConfig.Lang
	if lang == "" {
		lang = defaultLang
	}
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
	if lang, ok := langs[lang][key]; ok {
		return lang
	}

	if lang, ok := langs[defaultLang][key]; ok {
		return lang
	}

	return ""
}

func (e *errWrite) Write(p []byte) (n int, err error) {
	Error(ztype.ToString(p))
	return 1, nil
}

func SetVar(name, usage string) *Var {
	return &Var{
		name:  name,
		usage: usage,
	}
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

func (v *Var) String(defaultValue ...string) *string {
	var value string
	if len(defaultValue) > 0 {
		value = defaultValue[0]
	}
	return flag.String(v.name, value, v.usage)
}

func (v *Var) Int(defaultValue ...int) *int {
	var value int
	if len(defaultValue) > 0 {
		value = defaultValue[0]
	}
	return flag.Int(v.name, value, v.usage)
}

func (v *Var) Bool(defaultValue ...bool) *bool {
	var value bool
	if len(defaultValue) > 0 {
		value = defaultValue[0]
	}
	return flag.Bool(v.name, value, v.usage)
}

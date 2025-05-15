package zcli

import (
	"flag"
	"os"
	"strings"

	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/ztype"
)

type (
	// cmdCont is an internal structure that holds command information and configuration
	cmdCont struct {
		command       Cmd
		name          string
		desc          string
		Supplement    string
		requiredFlags []string
	}
	// runFunc is a function type for the main application run handler
	runFunc func()
	// Cmd represents a subcommand implementation that can define flags and execute code
	Cmd interface {
		// Flags allows the command to define its own flags and options
		Flags(subcommand *Subcommand)
		// Run executes the command with the provided arguments
		Run(args []string)
	}
	// errWrite is an internal type for custom error output handling
	errWrite struct {
	}
	// v represents a flag variable with its name, usage description, and short aliases
	v struct {
		name   string
		usage  string
		shorts []string
	}
	// Subcommand represents a CLI subcommand with its own flags and parameters
	Subcommand struct {
		// CommandLine is the flag set for this subcommand
		CommandLine *flag.FlagSet
		// Name is the subcommand name as used on the command line
		Name        string
		// Desc is the short description of the subcommand
		Desc        string
		// Supplement provides additional detailed information about the subcommand
		Supplement  string
		// Parameter describes the parameters accepted by the subcommand
		Parameter   string
		cmdCont
	}
)

const cliPrefix = ""

var (
	// BuildTime represents the application build timestamp
	BuildTime = ""
	// BuildGoVersion represents the Go version used to build the application
	BuildGoVersion = ""
	// BuildGitCommitID represents the Git commit ID of the build
	BuildGitCommitID = ""
	// Log is the CLI logger instance used for output formatting
	Log *zlog.Logger
	// FirstParameter contains the executable name as invoked on the command line
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

// SetLangText adds or updates a localized text string for the specified language and key.
// This allows customizing or extending the built-in localization support.
func SetLangText(lang, key, value string) {
	l, ok := langs[lang]
	if !ok {
		l = map[string]string{}
	}
	l[key] = value
	langs[lang] = l
}

// GetLangText retrieves a localized text string for the current language setting.
// If the key is not found in the current language, it falls back to the default language.
// If still not found and a default value is provided, it returns that value.
// Otherwise, it returns the key itself.
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

// Write implements the io.Writer interface for custom error handling.
// It formats and displays error messages through the Error function.
func (e *errWrite) Write(p []byte) (n int, err error) {
	Error(strings.Replace(ztype.ToString(p), cliPrefix, "", 1))
	return 1, nil
}

// SetVar creates a new flag variable with the specified name and usage description.
// It returns a variable object that can be further configured with type, default value, and options.
func SetVar(name, usage string) *v {
	v := &v{
		name:  cliPrefix + name,
		usage: usage,
	}
	varsKey[name] = v
	return v
}

// short adds a short alias for the flag (e.g., -h for --help).
// Returns the variable object for method chaining.
func (v *v) short(short string) *v {
	v.shorts = append(v.shorts, short)
	// todo prevent duplicate addition
	varShortsKey = append(varShortsKey, short)
	return v
}

// Required marks the flag as required, meaning the application will report an error
// if the flag is not provided by the user. Returns the variable object for method chaining.
func (v *v) Required() *v {
	if matchingCmd != nil {
		matchingCmd.requiredFlags = append(matchingCmd.requiredFlags, v.name)
	} else {
		requiredFlags = append(requiredFlags, v.name)
	}
	return v
}

// String defines a string flag with an optional default value.
// Returns a pointer to the string value that will be populated when the flag is parsed.
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

// Int defines an integer flag with an optional default value.
// Returns a pointer to the integer value that will be populated when the flag is parsed.
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

// Bool defines a boolean flag with an optional default value.
// Returns a pointer to the boolean value that will be populated when the flag is parsed.
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

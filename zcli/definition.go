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
	requiredFlags  RequiredFlags
	appConfig      = &App{}
)

type (
	App struct {
		Logo     string
		Version  string
		HideHelp bool
	}
	cmdCont struct {
		name          string
		desc          string
		command       Cmd
		requiredFlags RequiredFlags
	}
	runFunc func()
	// RequiredFlags RequiredFlags flags
	RequiredFlags []string
	// Cmd represents a subCommand
	Cmd interface {
		Flags(*flag.FlagSet) *flag.FlagSet
		Run(args []string)
	}
	errWrite struct {
		id int
	}
	stringArr []string
)

func (e *errWrite) Write(p []byte) (n int, err error) {
	Error(ztype.ToString(p))
	return 1, nil
}

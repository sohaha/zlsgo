package zcli

import (
	"errors"
	"flag"
	"os"
	"testing"

	zls "github.com/sohaha/zlsgo"
)

var globalDebug = SetVar("debug", "是否开启调试").Required().Bool()

// var globalTest = SetVar("test--------------test----------------test", "testtesttesttesttesttesttesttesttesttesttest").Bool(true)

type testCmd struct {
	flag1 *bool
	flag2 *int
	flag3 *string
	run   bool
}

func (cmd *testCmd) Flags(sub *Subcommand) {
	cmd.flag1 = SetVar("flag1", "Name about flag1").Required().Bool()
	cmd.flag2 = SetVar("flag2", "Name about flag2").Int(1)
	cmd.flag3 = SetVar("flag333333333333333333333333333333333333", "Name about flag333333333333333333333333333333").String("666")
}

func (cmd *testCmd) Run(args []string) {
	Log.Debug("run")
	cmd.run = true
}

func resetForTesting(args ...string) {
	os.Args = append([]string{"cmd"}, args...)
	FirstParameter = os.Args[0]
	Log.Debugf("resetForTesting: %s\n", os.Args)
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
}

func testOther(t *testing.T) {
	tt := zls.NewTest(t)
	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()
	myExit := func(code int) {
	}
	osExit = myExit
	Logo = "test"
	Version = "1.0.0"
	HideHelp = true
	Name = "test"
	Lang = "zh"
	s := GetLangText("test-key", "no")
	tt.Equal("no", s)
	s = GetLangText("test-key")
	tt.Equal("test-key", s)
	SetLangText("zh", "isName", "yes")
	s = GetLangText("isName")
	tt.Equal("yes", s)

	tipText("ok")
	errorText("err")
	showText("show")
	warnText("warn")
	Add("", "", &testCmd{})
	Help()
}

func TestUtil(T *testing.T) {
	BuildGoVersion = "--"
	BuildTime = "--"
	showVersionNum(true)
	Version = ""
	showVersion()
	_, _ = GetParentProcessName()
	IsDoubleClickStartUp()
	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()
	myExit := func(code int) {
	}
	osExit = myExit
	CheckErr(nil, true)
	CheckErr(errors.New("err"), true)
	Name = "Name"
	Logo = "Logo"
	showHeadr()
	showFlagsHelp()
	showLogo()
	Error("%s", "err")
	argsIsHelp([]string{"-h"})
}

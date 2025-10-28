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
	tt    *zls.TestUtil
	run   bool
}

func (cmd *testCmd) Flags(sub *Subcommand) {
	cmd.flag1 = SetVar("flag1", "Name about flag1").Required().Bool(true)
	cmd.flag2 = SetVar("flag2", "Name about flag2").Int(1)
	cmd.flag3 = SetVar("flag333333333333333333333333333333333333", "Name about flag333333333333333333333333333333").String("666")
}

func (cmd *testCmd) Run(args []string) {
	Log.Debug("run")
	Log.Debug(Current())
	Log.Debug("flag1", *cmd.flag1)
	cmd.tt.EqualExit(true, *cmd.flag1)
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
	oldLang := Lang
	defer func() { 
		osExit = oldOsExit
		Lang = oldLang
	}()
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

func TestUtil(t *testing.T) {
	BuildGoVersion = "--"
	BuildTime = "--"
	showVersionNum(true)
	Version = ""
	showVersion()
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

func TestDetach(t *testing.T) {
	t.Log(isDetach("detach"))
	t.Log(isDetach("dd"))
}

func TestIsSudo(t *testing.T) {
	t.Log(IsSudo())
}

func TestLockInstance(t *testing.T) {
	clean, ok := LockInstance()
	if ok {
		t.Log("lock ok")
		clean()
	} else {
		t.Log("lock failed")
	}
}

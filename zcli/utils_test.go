package zcli

import (
	"flag"
	"os"
)

var globalDebug = SetVar("debug", "是否开启调试").Required().Bool()
var globalTest = SetVar("test--------------test----------------test", "testtesttesttesttesttesttesttesttesttesttest").Bool(true)

type testCmd struct {
	flag1 *bool
	flag2 *int
	flag3 *string
	run   bool
}

func (cmd *testCmd) Flags(sub *Subcommand) {
	cmd.flag1 = SetVar("flag1", "Description about flag1").Required().Bool()
	cmd.flag2 = SetVar("flag2", "Description about flag2").Int(1)
	cmd.flag3 = SetVar("flag333333333333333333333333333333333333", "Description about flag333333333333333333333333333333").String("666")
}

func (cmd *testCmd) Run(args []string) {
	cmd.run = true
}

func resetForTesting(args ...string) {
	os.Args = append([]string{"cmd"}, args...)
	firstParameter = os.Args[0]
	Log.Debug(os.Args)
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
}

func testOther() {
	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()
	myExit := func(code int) {
	}
	osExit = myExit
	Logo = "test"
	Version = "1.0.0"
	HideHelp = true
	Description = "test"
	Lang = "zh"
	getLangs("test")
	tipText("ok")
	errorText("err")
	showText("show")
	warnText("warn")
	Add("", "", &testCmd{})
	Help()
}

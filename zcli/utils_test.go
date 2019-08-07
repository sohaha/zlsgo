package zcli

import (
	"flag"
	"github.com/sohaha/zlsgo"
	"os"
)

type testCmd struct {
	flag1 *bool
	flag2 *int
	run   bool
}

func (cmd *testCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.flag1 = fs.Bool("flag1", false, "Description about flag1")
	cmd.flag2 = fs.Int("flag2", 1, "Description about flag2")
	return fs
}

func (cmd *testCmd) Run(args []string) {
	cmd.run = true
}

func resetForTesting(args ...string) {
	os.Args = append([]string{"cmd"}, args...)
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
}

func testOther(_ *zlsgo.TestUtil) {
	SetApp(&App{
		Logo:    "test",
		Version: "1.0.0",
	})
	tipText("ok")
	errorText("err")
	showText("show")
	warnText("warn")
	Help()
}

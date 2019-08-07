package zcli

import (
	"flag"
	"github.com/sohaha/zlsgo"
	"testing"
)

var (
	globalFlagText = "-global=hello"
	globalText     = "global"
)

func TestCli(T *testing.T) {
	t := zlsgo.NewTest(T)
	testRun(t)
	defaultGlobalFlags(t)
	testCommand(t)
	testCommand2(t)
	testCommandRequired(t)
	testOther(t)
}

func testRun(_ *zlsgo.TestUtil) {
	resetForTesting("")
	Run()
}

func testCommand2(t *zlsgo.TestUtil) {
	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()
	var got int
	myExit := func(code int) {
		got = code
	}
	osExit = myExit
	resetForTesting(globalFlagText, "testCommand4", "help")
	Add("testCommand4", "", &testCmd{}, RequiredFlags{})
	Run()
	t.Equal(0, got)
}

func testCommandRequired(t *zlsgo.TestUtil) {
	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()
	var got int
	myExit := func(code int) {
		got = code
	}
	osExit = myExit
	GlobalRequired(RequiredFlags{"flag2", globalText})
	resetForTesting(globalFlagText, "testCommand2", "-flag2=1")
	c1 := &testCmd{}
	c1cmd := Add("testCommand2", "", c1, RequiredFlags{"flag", globalText})
	Add("", "", c1, nil)

	showSubcommandUsage(flag.CommandLine, c1cmd)
	Run()
	t.Equal(true, c1.run)
	t.Equal(got, 1)
}

func testCommand(t *zlsgo.TestUtil) {
	resetForTesting(globalFlagText, "testCommand3", "-flag2=1")
	flagGlobal := flag.String(globalText, "", "Description about global")
	c1 := &testCmd{}
	Add("testCommand3", "", c1, RequiredFlags{})
	Run()
	t.Equal(true, c1.run)
	t.Equal(false, *c1.flag1)
	t.Equal(1, *c1.flag2)
	t.Equal("hello", *flagGlobal)
}

func defaultGlobalFlags(t *zlsgo.TestUtil) {
	resetForTesting()
	flagGlobal := flag.String(globalText, "", "Description ")
	parse(true)
	t.Equal(*flagGlobal, "")

	resetForTesting(globalFlagText)
	flagGlobal = flag.String(globalText, "", "Description ")
	parse(true)
	t.Equal(*flagGlobal, "hello")

	resetForTesting(globalFlagText, "-global2=hi")
	flag.String("global1", "1", "")
	flag.String("global2", "2", "")
	total := numOfGlobalFlags()

	t.Equal(total, 2)
}

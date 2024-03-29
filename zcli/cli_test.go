package zcli

import (
	"flag"
	"os"
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestCli(T *testing.T) {
	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()
	myExit := func(code int) {
	}
	osExit = myExit
	t := zlsgo.NewTest(T)
	// resetForTesting("-debug")
	Logo = `
________  ____  .__   .__
\___   /_/ ___\ |  |  |__|
 /    / \  \___ |  |  |  |
/_____ \ \___  >|  |__|  |
      \/     \/ |____/|__|`
	Version = "1.0.1"
	Add("run", "run", &testCmd{})
	Run(func() {
		t.Log("Run", *globalDebug)
	})
}

func TestCli2(T *testing.T) {
	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()
	myExit := func(code int) {
	}
	osExit = myExit
	resetForTesting("-debug")
	Run()
}

func TestVersion(T *testing.T) {
	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()
	myExit := func(code int) {
	}
	osExit = myExit
	t := zlsgo.NewTest(T)
	resetForTesting("-version")
	Version = "1.0.1"
	Run(func() {
		current, b := Current()
		t.Equal(nil, current)
		t.EqualTrue(!b)
		t.Log("Run", *globalDebug)
	})
}

func TestCliOther(t *testing.T) {
	testOther(t)
}

func TestCliCommand(t *testing.T) {
	tt := zlsgo.NewTest(t)
	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()
	myExit := func(code int) {
		t.Log("myExit:", code)
	}
	osExit = myExit
	requiredFlags = []string{}
	resetForTesting("test", "-flag1")
	Add("test", "test", &testCmd{
		tt: tt,
	})
	Run()
	showFlags(flag.CommandLine)
}

func TestCliCommandErr(_ *testing.T) {
	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()
	myExit := func(code int) {
	}
	osExit = myExit
	requiredFlags = []string{}
	resetForTesting("test")
	Add("test", "test", &testCmd{})
	Run()
}

func TestCliCommandHelp(t *testing.T) {
	expectedName := "gopher"
	requiredFlags = []string{}
	resetForTesting("testHelp", "-help")
	matchingCmd := Add("testHelp", "test", &testCmd{})
	expectedErrorHandling := flag.ExitOnError
	expectedOutput := os.Stdout
	parseSubcommand(flag.Args())
	flag.CommandLine.Init(expectedName, expectedErrorHandling)
	flag.CommandLine.SetOutput(expectedOutput)
	showSubcommandUsage(flag.CommandLine, matchingCmd)
	showFlags(flag.CommandLine)
}

func TestCliCommandHelp2(t *testing.T) {
	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()
	myExit := func(code int) {
		t.Log("myExit:", code)
	}
	osExit = myExit
	requiredFlags = []string{}
	resetForTesting("test", "ddd", "-h")
	Add("test", "test", &testCmd{})
	Run()
}

func TestUnknown(_ *testing.T) {
	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()
	myExit := func(code int) {
	}
	osExit = myExit
	resetForTesting("unknown")
	Run()
}

func TestUnknown2(T *testing.T) {
	t := zlsgo.NewTest(T)
	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()

	myExit := func(code int) {
		t.Log(code)
	}
	osExit = myExit
	SetUnknownCommand(func(name string) {
		t.Log(name)
	})
	resetForTesting("unknown")
	Run()
}

func TestInput(t *testing.T) {
	Inputln("test:", false)
}

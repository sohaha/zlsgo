package zcli

import (
	"flag"
	"github.com/sohaha/zlsgo"
	"os"
	"testing"
)

func TestCli(T *testing.T) {
	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()
	myExit := func(code int) {
	}
	osExit = myExit
	t := zlsgo.NewTest(T)
	resetForTesting("-debug")
	SetApp(&App{
		Logo: `
________  ____  .__   .__
\___   /_/ ___\ |  |  |__|
 /    / \  \___ |  |  |  |
/_____ \ \___  >|  |__|  |
      \/     \/ |____/|__|`,
		Version: "1.0.1",
	})
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
	SetApp(&App{
		Version: "1.0.1",
	})
	Run(func() {
		t.Log("Run", *globalDebug)
	})
}

func TestCliOther(_ *testing.T) {
	testOther()
}

func TestCliCommand(_ *testing.T) {
	requiredFlags = RequiredFlags{}
	resetForTesting("test", "-flag1")
	Add("test", "test", &testCmd{})
	Run()
	showFlags(flag.CommandLine)
}

func TestCliCommandErr(_ *testing.T) {
	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()
	myExit := func(code int) {
	}
	osExit = myExit
	requiredFlags = RequiredFlags{}
	resetForTesting("test")
	Add("test", "test", &testCmd{})
	Run()
}

func TestCliCommandHelp(_ *testing.T) {
	expectedName := "gopher"
	requiredFlags = RequiredFlags{}
	resetForTesting("testHelp", "-help")
	matchingCmd := Add("testHelp", "test", &testCmd{})
	expectedErrorHandling := flag.ExitOnError
	expectedOutput := os.Stdout
	parseSubcommand()
	flag.CommandLine.Init(expectedName, expectedErrorHandling)
	flag.CommandLine.SetOutput(expectedOutput)
	showSubcommandUsage(flag.CommandLine, matchingCmd)
	showFlags(flag.CommandLine)
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

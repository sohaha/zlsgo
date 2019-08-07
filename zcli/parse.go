package zcli

import (
	"errors"
	"flag"
	"fmt"
	"github.com/sohaha/zlsgo/zarray"
	"github.com/sohaha/zlsgo/ztype"
	"strings"
)

func parse(outHelp bool) {
	parseCommand(outHelp)
	parseSubcommand()
}

func parseRequiredFlags(fs *flag.FlagSet, requiredFlags RequiredFlags) (err error) {
	requiredFlagsLen := len(requiredFlags)
	if requiredFlagsLen > 0 {
		flagMap := zarray.New(requiredFlagsLen)
		for _, flagName := range requiredFlags {
			flagMap.Push(flagName)
		}
		fs.Visit(func(f *flag.Flag) {
			_, _ = flagMap.RemoveValue(f.Name)
		})
		flagMapLen := flagMap.Length()
		if flagMapLen > 0 && !*flagHelp {
			arr := make([]string, flagMapLen)
			for i := 0; i < flagMapLen; i++ {
				value, _ := flagMap.Get(i)
				arr[i] = "-" + ztype.ToString(value)
			}
			err = errors.New(fmt.Sprintf("required flags: %s", strings.Join(arr, ", ")))
		}
	}
	return
}
func parseCommand(outHelp bool) {
	flag.Parse()
	if len(cmds) < 1 {
		return
	}
	flag.Usage = usage
	requiredErr := parseRequiredFlags(flag.CommandLine, requiredFlags)
	if requiredErr != nil {
		if len(flag.Args()) > 0 {
			Error(requiredErr.Error())
		} else {
			Help()
			osExit(0)
		}
	}
	if flag.NArg() < 1 {
		if outHelp {
			Help()
		}
		return
	}

}

func parseSubcommand() {
	name := flag.Arg(0)
	if cont, ok := cmds[name]; ok {
		firstParameter += " " + name
		fs := cont.command.Flags(flag.NewFlagSet(name, flag.ExitOnError))
		fs.SetOutput(&errWrite{})
		fs.Usage = func() {
			Log.Printf("usage of %s %s:\n", firstParameter, name)
			Log.Printf("  %s", cont.desc)
			showFlags(fs)
			showRequired(fs, cont.requiredFlags)
		}
		_ = fs.Parse(flag.Args()[1:])
		args = fs.Args()
		matchingCmd = cont
		flagMap := zarray.New(len(cont.requiredFlags))
		for _, flagName := range cont.requiredFlags {
			flagMap.Push(flagName)
		}
		fs.Visit(func(f *flag.Flag) {
			_, _ = flagMap.RemoveValue(f.Name)
		})
		flagMapLen := flagMap.Length()
		if flagMapLen > 0 && !*flagHelp {
			arr := make([]string, flagMapLen)
			for i := 0; i < flagMapLen; i++ {
				value, _ := flagMap.Get(i)
				arr[i] = "-" + ztype.ToString(value)
			}
			Error("required flags: %s", strings.Join(arr, ", "))
		}

		flag.CommandLine = fs
	} else if name != "" {
		Error("unknown testCommand: %s", errorText(name))
	}
	return
}

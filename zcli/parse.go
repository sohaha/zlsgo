package zcli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sohaha/zlsgo/zarray"
	"github.com/sohaha/zlsgo/ztype"
)

func parse(outHelp bool) {
	parseCommand(outHelp)
	parseSubcommand(flag.Args())
}

func parseRequiredFlags(fs *flag.FlagSet, requiredFlags RequiredFlags) (err error) {
	requiredFlagsLen := len(requiredFlags)
	if requiredFlagsLen > 0 {
		flagMap := zarray.New(requiredFlagsLen)
		for _, flagName := range requiredFlags {
			flagMap.Push(flagName)
		}
		fs.Visit(func(f *flag.Flag) {
			Log.Error(f.Name)

			_, _ = flagMap.RemoveValue(f.Name)
		})
		flagMapLen := flagMap.Length()
		if flagMapLen > 0 && !*flagHelp {
			arr := make([]string, flagMapLen)
			for i := 0; i < flagMapLen; i++ {
				value, _ := flagMap.Get(i)
				arr[i] = "-" + ztype.ToString(value)
			}
			err = fmt.Errorf("required flags: %s", strings.Join(arr, ", "))
		}
	}
	return
}

// Parse Parse
func Parse(arg ...[]string) {
	if Version != "" {
		flagVersion = SetVar("version", GetLangText("version")).short("V").Bool()
	}
	var argsData []string
	if len(arg) == 1 {
		argsData = arg[0]
	} else {
		argsData = os.Args[1:]
	}
	for k := range argsData {
		s := argsData[k]
		if len(s) < 2 || s[0] != '-' {
			continue
		}
		var prefix = "-"
		if s[1] == '-' {
			s = s[2:]
			prefix = "--"
		} else {
			s = s[1:]
		}
		for key := range varsKey {
			if key == s {
				argsData[k] = prefix + cliPrefix + s
			}
		}
	}
	if len(argsData) > 0 && argsData[0][0] != '-' {
		parseSubcommand(argsData)
		matchingCmd.command.Run(args)
		osExit(0)
	}

	_ = flag.CommandLine.Parse(argsData)
	var v *bool
	var ok bool
	if v, ok = ShortValues["V"].(*bool); ok && *v {
		*flagVersion = *v
	}
	if *flagVersion {
		showVersionNum(*v)
		osExit(0)
		return
	}
}

func parseCommand(outHelp bool) {
	Parse()
	if len(cmds) < 1 {
		return
	}
	flag.Usage = usage
	requiredErr := parseRequiredFlags(flag.CommandLine, requiredFlags)
	if requiredErr != nil {
		if len(flag.Args()) > 0 {
			Error(requiredErr.Error())
		} else if outHelp {
			Help()
		}
	}

	if flag.NArg() < 1 {
		if outHelp {
			Help()
		}
		return
	}
}

func parseSubcommand(Args []string) {
	var name = ""
	if len(Args) > 0 {
		name = Args[0]
	}
	if cont, ok := cmds[name]; ok {
		matchingCmd = cont
		FirstParameter += " " + name
		fsArgs := Args[1:]
		fs := flag.NewFlagSet(name, flag.ExitOnError)
		flag.CommandLine = fs
		subcommand := &Subcommand{
			Name:        cont.name,
			Desc:        cont.desc,
			Supplement:  cont.Supplement,
			Parameter:   FirstParameter,
			CommandLine: fs,
		}
		cont.command.Flags(subcommand)
		fs.SetOutput(&errWrite{})
		fs.Usage = func() {
			Log.Printf("%s\n", subcommand.Desc)
			if subcommand.Supplement != "" {
				Log.Printf("%s\n", subcommand.Supplement)
			}
			Log.Printf("\nusage of %s\n", subcommand.Parameter)
			ShowFlags(fs)
			ShowRequired(fs, cont.requiredFlags)
		}
		_ = fs.Parse(fsArgs)
		args = fs.Args()
		argsIsHelp(args)
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
	} else if name != "" {
		unknownCommandFn(name)
		osExit(1)
	}
}

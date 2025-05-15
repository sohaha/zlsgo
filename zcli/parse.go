package zcli

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/sohaha/zlsgo/zarray"
	"github.com/sohaha/zlsgo/ztype"
)

var runCmd = []string{os.Args[0]}

// parse processes command line arguments and identifies subcommands.
// If outHelp is true, help information will be displayed when appropriate.
func parse(outHelp bool) {
	parseCommand(outHelp)
	parseSubcommand(flag.Args())
}

// parseRequiredFlags checks if all required flags have been provided.
// It returns an error if any required flags are missing.
func parseRequiredFlags(fs *flag.FlagSet, requiredFlags []string) (err error) {
	requiredFlagsLen := len(requiredFlags)
	if requiredFlagsLen > 0 {
		flagMap := zarray.NewArray(requiredFlagsLen)
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

var parseDone sync.Once

// Parse processes command line arguments, handling short flag aliases and special flags like version and detach.
// It returns true if any flags were provided in the arguments.
func Parse(arg ...[]string) (hasflag bool) {
	parseDone.Do(func() {
		if Version != "" {
			flagVersion = SetVar("version", GetLangText("version")).short("V").Bool()
		}
		if EnableDetach {
			flagDetach = SetVar("detach", GetLangText("detach")).Bool()
		}
	})

	var argsData []string
	if len(arg) == 1 {
		argsData = arg[0]
	} else {
		argsData = os.Args[1:]
	}
	for k := range argsData {
		s := argsData[k]
		if !isDetach(s) {
			runCmd = append(runCmd, s)
		}
		if len(s) < 2 || s[0] != '-' {
			continue
		}
		prefix := "-"
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

	hasflag = len(argsData) > 0

	if hasflag && argsData[0][0] != '-' {
		parseSubcommand(argsData)
		if matchingCmd != nil {
			matchingCmd.command.Run(args)
		}
		osExit(0)
	}

	_ = flag.CommandLine.Parse(argsData)
	v, ok := ShortValues["V"].(*bool)
	if ok && *v {
		*flagVersion = *v
	}
	if d, ok := ShortValues["D"].(*bool); ok && *d {
		*flagDetach = *d
	}
	if *flagVersion {
		showVersionNum(*v)
		osExit(0)
		return
	}
	return
}

// parseCommand processes the main command line arguments and validates required flags.
// If outHelp is true and there are errors or no arguments, help information will be displayed.
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

// parseSubcommand identifies and processes a subcommand from the provided arguments.
// It sets up the subcommand's flag set, validates required flags, and prepares for execution.
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
		flag.CommandLine.VisitAll(func(f *flag.Flag) {
			fs.Var(f.Value, f.Name, f.Usage)
		})
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
			Log.Printf("%s\n\n", subcommand.Desc)
			if subcommand.Supplement != "" {
				Log.Printf("%s\n\n", subcommand.Supplement)
			}
			Log.Printf("\nusage of %s\n\n", subcommand.Parameter)
			showFlags(fs)
			showRequired(fs, cont.requiredFlags)
		}
		_ = fs.Parse(fsArgs)
		args = fs.Args()
		argsIsHelp(args)
		flagMap := zarray.NewArray(len(cont.requiredFlags))
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

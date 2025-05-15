// Package zcli provides tools for quickly building command-line applications
// with support for subcommands, flags, help documentation, and interactive prompts.
package zcli

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/zshell"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/ztype"
)

// init initializes the CLI environment by setting up logging and customizing flag behavior
func init() {
	Log = zlog.New()
	Log.ResetFlags(zlog.BitLevel)
	// flag.CommandLine.SetOutput(ioutil.Discard)
	flag.CommandLine.SetOutput(&errWrite{})
	flag.Usage = func() {
		usage()
	}
}

// Add registers a command handler for the provided subcommand name.
// Returns a command container that can be further configured with flags and options.
func Add(name, description string, command Cmd) *cmdCont {
	if name == "" {
		Log.Error(GetLangText("command_empty"))
		return &cmdCont{}
	}
	cmd := &cmdCont{
		name:          name,
		desc:          description,
		command:       command,
		requiredFlags: []string{},
	}
	cmds[name] = cmd
	cmdsKey = append(cmdsKey, name)
	return cmd
}

// SetUnknownCommand sets a handler function to be called when an unknown command is encountered.
// This allows custom handling of command errors or suggestions for similar commands.
func SetUnknownCommand(fn func(_ string)) {
	unknownCommandFn = fn
}

// usage displays the application's usage information including available commands,
// global flags, and required parameters.
func usage() {
	showHeadr()
	showFlagsAndRequired := func() {
		if numOfGlobalFlags() > 0 {
			showFlags(flag.CommandLine)
			showRequired(flag.CommandLine, requiredFlags)
		}
	}
	if len(cmds) == 0 {
		Log.Printf("usage of %s\n\n", showText(FirstParameter))
		showFlagsAndRequired()
		return
	}
	Log.Printf("usage: %s <command>\n\n\n", FirstParameter)
	Log.Println("where <command> is one of:")
	for _, name := range cmdsKey {
		if cont, ok := cmds[name]; ok {
			// for name, cont := range cmds {
			Log.Printf("    "+tipText("%-19s")+" %s\n", name, cont.desc)
		}
	}

	showFlagsAndRequired()
	if !HidePrompt {
		Log.Printf(showText("\nMore Command information, please use: %s <command> --help\n"), FirstParameter)
	}
}

// showFlags displays all flags in the provided flag set with their descriptions,
// types, and default values in a formatted layout.
func showFlags(fg *flag.FlagSet) {
	Log.Printf("\noptional flags:\n")
	max := 40
	showFlagsHelp()
	flagsItems := zstring.Buffer()
	fg.VisitAll(func(f *flag.Flag) {
		s := zstring.Buffer()
		flagsTitle := strings.Replace(f.Name, cliPrefix, "", 1)
		for _, key := range varShortsKey {
			if flagsTitle == key {
				return
			}
		}
		output := false
		if flagsTitle == "version" {
			output = true
		}
		name, usage := flag.UnquoteUsage(f)
		for key, v := range varsKey {
			shorts := v.shorts
			if key == flagsTitle && len(shorts) > 0 {
				for key := range shorts {
					shorts[key] = "-" + shorts[key]
				}
				flagsTitle += ", " + strings.Join(shorts, ", ")
			}
		}
		// if name == "" {
		// 	name = "bool"
		// }
		sf := "    -%-12s"
		if len(name) > 0 {
			newName := showText("<" + name + ">")
			namePadLen := 12 + len(newName) - len(name)
			flagsTitle += " " + newName
			sf = "    -%-" + ztype.ToString(namePadLen) + "s"
		}
		s.WriteString(warnText(fmt.Sprintf(sf, flagsTitle)))
		if zstring.Len(s.String()) <= max {
			s.WriteString("\t")
		} else {
			s.WriteString("\n    \t")
		}
		s.WriteString(strings.Replace(usage, "\n", "\n    \t", -1))
		defValue := ztype.ToString(f.DefValue)
		if defValue != "" && defValue != "0" && defValue != "false" {
			s.WriteString(fmt.Sprintf(" (default %v)", defValue))
		}
		if output {
			Log.Println(s.String())
		} else {
			s.WriteString("\n")
			flagsItems.WriteString(s.String())
		}
	})

	Log.Println(flagsItems.String())
}

// Start executes the matched command or the provided run function.
// It handles help flag display, required flag validation, and background execution mode.
func Start(runFunc ...runFunc) {
	if *flagDetach {
		err := zshell.BgRun(strings.Join(runCmd, " "))
		if err != nil {
			Error(err.Error())
		}
		return
	}
	if matchingCmd != nil {
		if *flagHelp {
			showSubcommandUsage(flag.CommandLine, matchingCmd)
		} else {
			matchingCmd.command.Run(args)
		}
		return
	}
	requiredErr := parseRequiredFlags(flag.CommandLine, requiredFlags)
	if requiredErr != nil {
		Error(requiredErr.Error())
	}

	isRunFunc := len(runFunc) > 0
	if isRunFunc {
		runFunc[0]()
	} else {
		Help()
	}
}

// Run parses command line arguments and starts the application.
// If a run function is provided, it will be executed when no specific subcommand is matched.
// Returns true if execution was successful.
func Run(runFunc ...runFunc) (ok bool) {
	isRunFunc := len(runFunc) > 0
	parse(!isRunFunc)
	Start(runFunc...)
	return
}

// Input prompts the user for input with the given prompt text.
// If required is true, it will continue prompting until non-empty input is provided.
// Returns the user's input as a string.
func Input(problem string, required bool) (text string) {
	if problem != "" {
		fmt.Print(problem)
	}
	reader := bufio.NewReader(os.Stdin)
	text, _ = reader.ReadString('\n')
	if required && zstring.TrimSpace(text) == "" {
		return Input(problem, required)
	}
	return
}

// Inputln is similar to Input but adds a newline after the prompt text.
// Returns the user's input as a string.
func Inputln(problem string, required bool) (text string) {
	if problem != "" {
		problem = problem + "\n"
	}
	return Input(problem, required)
}

// Current returns the currently matched command and a boolean indicating whether a command was matched.
// This can be used to check which command is being executed or to access command-specific data.
func Current() (interface{}, bool) {
	if matchingCmd == nil {
		return nil, false
	}
	return matchingCmd.command, true
}

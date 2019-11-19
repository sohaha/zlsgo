package zcli

import (
	"flag"
	"fmt"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/ztype"
	"strings"
)

func errorText(msg string) string {
	return zlog.ColorTextWrap(zlog.ColorRed, msg)
}

func tipText(msg string) string {
	return zlog.ColorTextWrap(zlog.ColorGreen, msg)
}

func warnText(msg string) string {
	return zlog.ColorTextWrap(zlog.ColorLightYellow, msg)
}

func showText(msg string) string {
	return zlog.ColorTextWrap(zlog.ColorLightGrey, msg)
}

// Help show help
func Help() {
	if matchingCmd != nil {
		showSubcommandUsage(flag.CommandLine, matchingCmd)
	} else {
		flag.Usage()
	}
	osExit(0)
}

func numOfGlobalFlags() (count int) {
	flag.VisitAll(func(_ *flag.Flag) {
		count++
	})
	return
}

func Error(format string, v ...interface{}) {
	Log.Errorf(format, v...)
	if !HidePrompt {
		tip := zstring.Buffer()
		tip.WriteString("\nPlease use ")
		tip.WriteString(tipText("%s --help"))
		tip.WriteString(" for more information")
		Log.Printf(tip.String(), FirstParameter)
	}
	osExit(1)
}

func ShowRequired(_ *flag.FlagSet, requiredFlags RequiredFlags) {
	flagMapLen := len(requiredFlags)
	if flagMapLen > 0 {
		Log.Printf("\n  required flags:\n")
		arr := make([]string, flagMapLen)
		for i := 0; i < flagMapLen; i++ {
			arr[i] = "-" + ztype.ToString(requiredFlags[i])
		}
		Log.Printf("    %s\n\n", errorText(strings.Join(arr, ", ")))
	}
}

func showSubcommandUsage(fs *flag.FlagSet, _ *cmdCont) {
	fs.Usage()
}

func showLogo() bool {
	if Logo != "" {
		Log.Printf("%s\n", strings.Replace(Logo, "\n", "", 1))
		return true
	}
	return false
}

func showFlagsHelp() {
	if !HideHelp {
		help := zstring.Buffer()
		help.WriteString(warnText(fmt.Sprintf("    -%-12s", "help")))
		help.WriteString("\t")
		help.WriteString(getLangs("help"))
		Log.Println(help.String())
	}
}

func showDescription() bool {
	if Description != "" {
		Log.Printf("%s\n", Description)
		return true
	}
	return false
}

func showVersion() bool {
	if Version != "" {
		Log.Printf("Version: %s\n", Version)
		return true
	}
	return false
}

func showVersionNum(info bool) {
	if info {
		Log.Printf("version=%s\n", Version)
		//noinspection GoBoolExpressions
		if BuildGoVersion != "" {
			Log.Printf("goVersion=%s\n", BuildGoVersion)
		}
		//noinspection GoBoolExpressions
		if BuildTime != "" {
			Log.Printf("buildTime=%s\n", BuildTime)
		}
	} else {
		Log.Println(Version)
	}

	return
}

func showHeadr() {
	logoOk := showLogo()
	descriptionOk := showDescription()
	versionOk := showVersion()
	if logoOk || versionOk || descriptionOk {
		Log.Println("")
	}
}

func argsIsHelp(args []string) {
	if !*flagHelp {
		for _, value := range args {
			if value == "-h" || value == "-help" || value == "--help" {
				*flagHelp = true
				return
			}
		}
	}
}

package zcli

import (
	"flag"
	"fmt"
	"strings"

	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/ztype"
	"github.com/sohaha/zlsgo/zutil"
	"github.com/sohaha/zlsgo/zutil/daemon"
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
		tip := "\nPlease use " + tipText("%s --help") + " for more information"
		Log.Printf(tip, FirstParameter)
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
		Log.Printf("%s\n", strings.Replace(Logo+"v"+Version, "\n", "", 1))
		return true
	}
	return false
}

func showFlagsHelp() {
	if !HideHelp {
		Log.Println(warnText(fmt.Sprintf("    -%-12s", "help")) + "\t" + GetLangText("help"))
	}
}

func showDescription(logoOk bool) bool {
	if Name != "" {
		if !logoOk && Version != "" {
			Log.Printf("%s v%s\n", Name, Version)
		} else if !logoOk {
			Log.Printf("%s\n", Name)
		}

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
		//noinspection GoBoolExpressions
		if BuildGitCommitID != "" {
			Log.Printf("GitCommitID=%s\n", BuildGitCommitID)
		}
	} else {
		Log.Println(Version)
	}
}

func showHeadr() {
	var (
		versionOk     bool
		descriptionOk bool
	)
	logoOk := showLogo()
	descriptionOk = showDescription(logoOk)
	if !logoOk && !descriptionOk {
		versionOk = showVersion()
	}
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

func CheckErr(err error, exit ...bool) {
	if serviceErr == daemon.ErrNoServiceSystemDetected {
		err = fmt.Errorf("%s does not support process daemon\n", zutil.GetOs())
		exit = []bool{true}
	}
	if err != nil {
		if len(exit) > 0 && exit[0] {
			Log.Error(err)
			osExit(1)
			return
		}
		Log.Fatal(err)
	}
}

func IsDoubleClickStartUp() bool {
	return zutil.IsDoubleClickStartUp()
}

func isDetach(a string) bool {
	for _, v := range []string{"D", "detach"} {
		if strings.TrimLeft(a, "-") == v {
			return true
		}
	}
	return false
}

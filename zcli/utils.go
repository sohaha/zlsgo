package zcli

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/ztype"
	"github.com/sohaha/zlsgo/zutil"
	"github.com/sohaha/zlsgo/zutil/daemon"
)

// errorText formats a string with red color for error messages.
func errorText(msg string) string {
	return zlog.ColorTextWrap(zlog.ColorRed, msg)
}

// tipText formats a string with green color for tips and success messages.
func tipText(msg string) string {
	return zlog.ColorTextWrap(zlog.ColorGreen, msg)
}

// warnText formats a string with yellow color for warning messages.
func warnText(msg string) string {
	return zlog.ColorTextWrap(zlog.ColorLightYellow, msg)
}

// showText formats a string with light grey color for regular informational text.
func showText(msg string) string {
	return zlog.ColorTextWrap(zlog.ColorLightGrey, msg)
}

// Help displays usage information for the current command or subcommand and exits the program.
// If a subcommand is active, it shows help for that specific subcommand.
func Help() {
	if matchingCmd != nil {
		showSubcommandUsage(flag.CommandLine, matchingCmd)
	} else {
		flag.Usage()
	}
	osExit(0)
}

// numOfGlobalFlags returns the number of global flags defined in the application.
func numOfGlobalFlags() (count int) {
	flag.VisitAll(func(_ *flag.Flag) {
		count++
	})
	return
}

// Error prints a formatted error message, optionally shows a help tip, and exits the program with status code 1.
func Error(format string, v ...interface{}) {
	Log.Errorf(format, v...)
	if !HidePrompt {
		tip := "\nPlease use " + tipText("%s --help") + " for more information\n"
		Log.Printf(tip, FirstParameter)
	}
	osExit(1)
}

// showRequired displays a list of required flags that must be provided by the user.
func showRequired(_ *flag.FlagSet, requiredFlags []string) {
	flagMapLen := len(requiredFlags)
	if flagMapLen > 0 {
		Log.Printf("\n  required flags:\n\n")
		arr := make([]string, flagMapLen)
		for i := 0; i < flagMapLen; i++ {
			arr[i] = "-" + ztype.ToString(requiredFlags[i])
		}
		Log.Printf("    %s\n\n\n", errorText(strings.Join(arr, ", ")))
	}
}

// showSubcommandUsage displays the usage information for a subcommand using its flag set.
func showSubcommandUsage(fs *flag.FlagSet, _ *cmdCont) {
	fs.Usage()
}

// showLogo displays the application logo if one is defined.
// Returns true if a logo was displayed, false otherwise.
func showLogo() bool {
	if Logo != "" {
		Log.Printf("%s\n\n", strings.Replace(Logo+"v"+Version, "\n", "", 1))
		return true
	}
	return false
}

// showFlagsHelp displays help information for the help flag unless help display is disabled.
func showFlagsHelp() {
	if !HideHelp {
		Log.Println(warnText(fmt.Sprintf("    -%-12s", "help")) + "\t" + GetLangText("help"))
	}
}

// showDescription displays the application name and version if defined.
// The logoOk parameter indicates whether a logo was already displayed to avoid redundancy.
// Returns true if a description was displayed, false otherwise.
func showDescription(logoOk bool) bool {
	if Name != "" {
		if !logoOk && Version != "" {
			Log.Printf("%s v%s\n\n", Name, Version)
		} else if !logoOk {
			Log.Printf("%s\n\n", Name)
		}

		return true
	}
	return false
}

// showVersion displays the application version if defined.
// Returns true if a version was displayed, false otherwise.
func showVersion() bool {
	if Version != "" {
		Log.Printf("Version: %s\n\n", Version)
		return true
	}
	return false
}

// showVersionNum displays version information.
// If info is true, it shows detailed build information including Go version and build time.
// Otherwise, it only shows the version number.
func showVersionNum(info bool) {
	if info {
		Log.Printf("version=%s\n\n", Version)
		//noinspection GoBoolExpressions
		if BuildGoVersion != "" {
			Log.Printf("goVersion=%s\n\n", BuildGoVersion)
		}
		//noinspection GoBoolExpressions
		if BuildTime != "" {
			Log.Printf("buildTime=%s\n\n", BuildTime)
		}
		//noinspection GoBoolExpressions
		if BuildGitCommitID != "" {
			Log.Printf("GitCommitID=%s\n\n", BuildGitCommitID)
		}
	} else {
		Log.Println(Version)
	}
}

// showHeadr displays the application header, which may include the logo, name, and version.
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

// argsIsHelp checks if any of the provided arguments is a help flag (-h, -help, --help).
// If found, it sets the global help flag to true.
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

// CheckErr handles errors by logging them and optionally exiting the program.
// If exit is true, it calls os.Exit(1) after logging the error.
// It has special handling for service system detection errors.
func CheckErr(err error, exit ...bool) {
	if serviceErr == daemon.ErrNoServiceSystemDetected {
		err = errors.New(zutil.GetOs() + " does not support process daemon")
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

// isDetach checks if the provided argument is a detach flag (D or detach).
// Returns true if it is a detach flag, false otherwise.
func isDetach(a string) bool {
	for _, v := range []string{"D", "detach"} {
		if strings.TrimLeft(a, "-") == v {
			return true
		}
	}
	return false
}

// IsSudo checks if the current process is running with sudo/administrator privileges.
// Returns true if running with elevated privileges, false otherwise.
func IsSudo() bool {
	return daemon.IsSudo()
}

// IsDoubleClickStartUp detects if the application was started by double-clicking
// rather than from a command line.
// Returns true if started by double-click, false otherwise.
func IsDoubleClickStartUp() bool {
	return zutil.IsDoubleClickStartUp()
}

// LockInstance ensures only one instance of the application can run at a time.
// Returns a cleanup function and a boolean indicating whether the lock was acquired.
// If the lock was not acquired (ok is false), another instance is already running.
func LockInstance() (clean func(), ok bool) {
	ePath := zfile.ExecutablePath()
	lockName := zfile.TmpPath() + "/" + zstring.Md5(ePath) + ".SingleInstance.lock"
	lock := zfile.NewFileLock(lockName)
	if err := lock.Lock(); err != nil {
		return nil, false
	}
	return func() {
		lock.Clean()
	}, true
}

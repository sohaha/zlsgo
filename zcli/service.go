package zcli

import (
	"context"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sohaha/zlsgo/zerror"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zshell"
	"github.com/sohaha/zlsgo/zutil"
	"github.com/sohaha/zlsgo/zutil/daemon"
)

type (
	// app implements the daemon service interface for running the application as a service
	app struct {
		run    func()
		status bool
	}
	// serviceStop implements the Cmd interface for stopping a service
	serviceStop struct {
	}
	// serviceStart implements the Cmd interface for starting a service
	serviceStart struct {
	}
	// serviceRestart implements the Cmd interface for restarting a service
	serviceRestart struct {
	}
	// serviceInstall implements the Cmd interface for installing a service
	serviceInstall struct {
	}
	// serviceUnInstall implements the Cmd interface for uninstalling a service
	serviceUnInstall struct {
	}
	// serviceStatus implements the Cmd interface for checking a service's status
	serviceStatus struct {
	}
)

var (
	service    daemon.ServiceIface
	serviceErr error
	once       sync.Once
)

var s = make(chan struct{})

// Start implements the daemon.ServiceIface Start method for the app type.
// It runs the application function in a goroutine and returns any error that occurs.
func (a *app) Start(daemon.ServiceIface) error {
	a.status = true
	err := make(chan error, 1)
	go func() {
		err <- zerror.TryCatch(func() error {
			a.run()
			return nil
		})
		s <- struct{}{}
	}()
	return <-err
}

// Stop implements the daemon.ServiceIface Stop method for the app type.
// It waits for the application to stop with a timeout of 30 seconds.
func (a *app) Stop(daemon.ServiceIface) error {
	if !a.status {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	select {
	case <-s:
	case <-ctx.Done():
		// return errors.New("forced timeout")
	}
	return nil
}

// Flags implements the Cmd interface for the serviceStatus command.
// It checks for service errors before proceeding.
func (*serviceStatus) Flags(_ *Subcommand) {
	CheckErr(serviceErr, true)
}

// Run implements the Cmd interface for the serviceStatus command.
// It displays the current status of the service.
func (*serviceStatus) Run(_ []string) {
	log.Printf("%s: %s\n", service.String(), service.Status())
}

// Flags implements the Cmd interface for the serviceInstall command.
// It checks for service errors before proceeding.
func (*serviceInstall) Flags(_ *Subcommand) {
	CheckErr(serviceErr, true)
}

// Run implements the Cmd interface for the serviceInstall command.
// It installs and starts the service.
func (*serviceInstall) Run(_ []string) {
	CheckErr(service.Install(), true)
	CheckErr(service.Start(), true)
}

// Flags implements the Cmd interface for the serviceUnInstall command.
// It checks for service errors before proceeding.
func (*serviceUnInstall) Flags(_ *Subcommand) {
	CheckErr(serviceErr, true)
}

// Run implements the Cmd interface for the serviceUnInstall command.
// It uninstalls the service from the system.
func (*serviceUnInstall) Run(_ []string) {
	CheckErr(service.Uninstall(), true)
}

// Flags implements the Cmd interface for the serviceStart command.
// It checks for service errors before proceeding.
func (*serviceStart) Flags(_ *Subcommand) {
	CheckErr(serviceErr, true)
}

// Run implements the Cmd interface for the serviceStart command.
// It starts the service if it is not already running.
func (*serviceStart) Run(_ []string) {
	CheckErr(service.Start(), true)
}

// Flags implements the Cmd interface for the serviceStop command.
// It checks for service errors before proceeding.
func (*serviceStop) Flags(_ *Subcommand) {
	CheckErr(serviceErr, true)
}

// Run implements the Cmd interface for the serviceStop command.
// It stops the service if it is running.
func (*serviceStop) Run(_ []string) {
	CheckErr(service.Stop(), true)
}

// Flags implements the Cmd interface for the serviceRestart command.
// It checks for service errors before proceeding.
func (*serviceRestart) Flags(_ *Subcommand) {
	CheckErr(serviceErr, true)
}

// Run implements the Cmd interface for the serviceRestart command.
// It restarts the service.
func (*serviceRestart) Run(_ []string) {
	CheckErr(service.Restart(), true)
}

// LaunchServiceRun initializes a service with the given name and description,
// and runs it immediately. If the --detach flag is set, it will run the service
// in the background. If no service system is available, it runs the function directly.
func LaunchServiceRun(name string, description string, fn func(), config ...*daemon.Config) error {
	_, _ = LaunchService(name, description, fn, config...)
	Parse()
	if *flagDetach {
		return zshell.BgRun(strings.Join(runCmd, " "))
	}
	if serviceErr != nil && (serviceErr != daemon.ErrNoServiceSystemDetected && !daemon.IsPermissionError(serviceErr)) {
		return serviceErr
	}
	if service == nil {
		fn()
		return nil
	}
	return service.Run()
}

// LaunchService initializes a service with the given name, description, and run function.
// It also registers service management commands (install, uninstall, status, etc.).
// Returns the service interface and any error that occurred during initialization.
func LaunchService(name string, description string, fn func(), config ...*daemon.Config) (daemon.ServiceIface, error) {
	once.Do(func() {
		userService := false
		if zutil.IsMac() {
			userService = true
		}
		daemonConfig := &daemon.Config{
			Name:        name,
			Description: description,
			Options: map[string]interface{}{
				"UserService": userService,
			},
		}
		if len(os.Args) > 2 {
			daemonConfig.Arguments = os.Args[2:]
		}

		if len(config) > 0 {
			nconf := config[0]
			if nconf.Name == "" {
				nconf.Name = name
			}
			if nconf.Description == "" {
				nconf.Description = description
			}
			if len(nconf.Options) == 0 {
				nconf.Options = daemonConfig.Options
			}
			if len(nconf.Arguments) == 0 {
				nconf.Arguments = daemonConfig.Arguments
			}
			daemonConfig = nconf
		}

		// The file path is redirected to the current execution file path
		_, gogccflags, _, _ := zshell.Run("go env GOGCCFLAGS")
		if !strings.Contains(
			gogccflags, zfile.RealPath(zfile.ProgramPath()+"../../../..")) {
			zfile.ProjectPath = zfile.ProgramPath()
		}

		service, serviceErr = daemon.New(&app{
			run: fn,
		}, daemonConfig)

		Add("install", GetLangText("install"), &serviceInstall{})
		Add("uninstall", GetLangText("uninstall"), &serviceUnInstall{})
		Add("status", GetLangText("status"), &serviceStatus{})
		Add("start", GetLangText("start"), &serviceStart{})
		Add("stop", GetLangText("stop"), &serviceStop{})
		Add("restart", GetLangText("restart"), &serviceRestart{})
	})

	return service, serviceErr
}

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
	"github.com/sohaha/zlsgo/zutil/daemon"
)

type (
	app struct {
		run    func()
		status bool
	}
	serviceStop struct {
	}
	serviceStart struct {
	}
	serviceRestart struct {
	}
	serviceInstall struct {
	}
	serviceUnInstall struct {
	}
	serviceStatus struct {
	}
)

var (
	service    daemon.ServiceIface
	serviceErr error
	once       sync.Once
)

var s = make(chan struct{})

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

func (*serviceStatus) Flags(_ *Subcommand) {
	CheckErr(serviceErr, true)
}

func (*serviceStatus) Run(_ []string) {
	log.Printf("%s: %s\n", service.String(), service.Status())
}

func (*serviceInstall) Flags(_ *Subcommand) {
	CheckErr(serviceErr, true)
}

func (*serviceInstall) Run(_ []string) {
	CheckErr(service.Install(), true)
	CheckErr(service.Start(), true)
}

func (*serviceUnInstall) Flags(_ *Subcommand) {
	CheckErr(serviceErr, true)
}

func (*serviceUnInstall) Run(_ []string) {
	CheckErr(service.Uninstall(), true)
}

func (*serviceStart) Flags(_ *Subcommand) {
	CheckErr(serviceErr, true)
}

func (*serviceStart) Run(_ []string) {
	CheckErr(service.Start(), true)
}

func (*serviceStop) Flags(_ *Subcommand) {
	CheckErr(serviceErr, true)
}

func (*serviceStop) Run(_ []string) {
	CheckErr(service.Stop(), true)
}

func (*serviceRestart) Flags(_ *Subcommand) {
	CheckErr(serviceErr, true)
}

func (*serviceRestart) Run(_ []string) {
	CheckErr(service.Restart(), true)
}

// LaunchServiceRun Launch Service and run
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

// LaunchService Launch Service
func LaunchService(name string, description string, fn func(), config ...*daemon.Config) (daemon.ServiceIface, error) {
	once.Do(func() {
		daemonConfig := &daemon.Config{
			Name:        name,
			Description: description,
			Options: map[string]interface{}{
				"UserService": true,
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

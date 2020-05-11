package zcli

import (
	"log"
	"strings"
	"sync"

	"github.com/sohaha/zlsgo/zarray"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zshell"
	"github.com/sohaha/zlsgo/zutil/daemon"
)

type (
	app struct {
		run func()
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
	service    daemon.ServiceIfe
	serviceErr error
	once       sync.Once
)

var s = make(chan struct{})

func (a *app) Start(daemon.ServiceIfe) error {
	go func() {
		a.run()
		s <- struct{}{}
	}()
	return nil
}

func (*app) Stop(daemon.ServiceIfe) error {
	<-s
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

func LaunchServiceRun(name string, description string, fn func(), config ...*daemon.Config) error {
	_, _ = LaunchService(name, description, fn, config...)
	Parse()
	if serviceErr != nil && (serviceErr != daemon.ErrNoServiceSystemDetected && !daemon.IsPermissionError(serviceErr)) {
		return serviceErr
	}
	if service == nil {
		fn()
		return nil
	}
	return service.Run()
}

func LaunchService(name string, description string, fn func(), config ...*daemon.Config) (daemon.ServiceIfe, error) {

	once.Do(func() {
		var daemonConfig *daemon.Config
		if len(config) > 0 {
			daemonConfig = config[0]
			daemonConfig.Name = name
			daemonConfig.Description = description
		} else {
			daemonConfig = &daemon.Config{
				Name:        name,
				Description: description,
				Option: zarray.DefData{
					// "UserService": true,
				},
			}
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
		Add("install", "Install service", &serviceInstall{})
		Add("uninstall", "Uninstall service", &serviceUnInstall{})
		Add("status", "ServiceIfe status", &serviceStatus{})
		Add("start", "Start service", &serviceStart{})
		Add("stop", "Stop service", &serviceStop{})
		Add("restart", "Restart service", &serviceRestart{})
	})

	return service, serviceErr
}

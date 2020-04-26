package zcli

import (
	"fmt"
	"log"
	"sync"

	"github.com/sohaha/zlsgo/zenv"
	"github.com/sohaha/zlsgo/zfile"
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

func (a *app) Start(daemon.ServiceIfe) error {
	go a.run()
	return nil
}

func (*app) Stop(daemon.ServiceIfe) error {
	return nil
}

func (*serviceStatus) Flags(_ *Subcommand) {
	CheckErr(serviceErr, true)
}

func (*serviceStatus) Run(args []string) {
	log.Printf("%s: %s\n", service.String(), service.Status())
}

func (*serviceInstall) Flags(_ *Subcommand) {
	CheckErr(serviceErr, true)
}

func (*serviceInstall) Run(args []string) {
	CheckErr(service.Install(), true)
	CheckErr(service.Start(), true)
}

func (*serviceUnInstall) Flags(_ *Subcommand) {
	CheckErr(serviceErr, true)
}

func (*serviceUnInstall) Run(args []string) {
	CheckErr(service.Uninstall(), true)
}

func (*serviceStart) Flags(_ *Subcommand) {
	CheckErr(serviceErr, true)
}

func (*serviceStart) Run(args []string) {
	CheckErr(service.Start(), true)
}

func (*serviceStop) Flags(_ *Subcommand) {
	CheckErr(serviceErr, true)
}

func (*serviceStop) Run(args []string) {
	CheckErr(service.Stop(), true)
}

func (*serviceRestart) Flags(_ *Subcommand) {
	CheckErr(serviceErr, true)
}

func (*serviceRestart) Run(args []string) {
	CheckErr(service.Restart(), true)
}

func LaunchService(name string, description string, fn func()) daemon.ServiceIfe {
	// The file path is redirected to the current execution file path
	zfile.ProjectPath = zfile.ProgramPath()
	once.Do(func() {
		service, serviceErr = daemon.New(&app{
			run: fn,
		}, &daemon.Config{
			Name:        name,
			Description: description,
			Option: daemon.OptionData{
				// "UserService": true,
			},
		})
	})

	Add("install", "Install service", &serviceInstall{})
	Add("uninstall", "Uninstall service", &serviceUnInstall{})
	Add("status", "ServiceIfe status", &serviceStatus{})
	Add("start", "Start service", &serviceStart{})
	Add("stop", "Stop service", &serviceStop{})
	Add("restart", "Restart service", &serviceRestart{})

	if serviceErr == daemon.ErrNoServiceSystemDetected {
		CheckErr(fmt.Errorf("%s does not support process daemon\n", zenv.GetOs()), true)
	}

	return service
}

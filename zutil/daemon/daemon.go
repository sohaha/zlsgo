// Package daemon program is installed as a system service to achieve process daemon
package daemon

import (
	"errors"

	"github.com/sohaha/zlsgo/zarray"
)

const (
	optionKeepAlive            = "KeepAlive"
	optionKeepAliveDefault     = true
	optionRunAtLoad            = "RunAtLoad"
	optionRunAtLoadDefault     = true
	optionUserService          = "UserService"
	optionUserServiceDefault   = false
	optionSessionCreate        = "SessionCreate"
	optionSessionCreateDefault = false
	optionRunWait              = "RunWait"
)

type (
	// ServiceIfe represents a service that can be run or controlled
	ServiceIfe interface {
		Run() error
		Start() error
		Stop() error
		Restart() error
		Install() error
		Uninstall() error
		Status() string
		String() string
	}
	Ife interface {
		Start(s ServiceIfe) error
		Stop(s ServiceIfe) error
	}
	SystemIfe interface {
		String() string
		Interactive() bool
		Detect() bool
		New(i Ife, c *Config) (ServiceIfe, error)
	}
	// Config provides the setup for a ServiceIfe. The Name field is required.
	Config struct {
		Name        string
		DisplayName string
		Description string
		UserName    string
		Arguments   []string
		Executable  string
		WorkingDir  string
		RootDir     string
		// System specific options
		//  * OS X
		//    - KeepAlive     bool (true)
		//    - RunAtLoad     bool (true)
		//    - UserService   bool (false) - Install as a current user service.
		//    - SessionCreate bool (false) - Create a full user session.
		//  * POSIX
		//    - RunWait      func() (wait for SIGNAL) - Do not install signal but wait for this function to return.
		//    - ReloadSignal string () [USR1, ...] - Signal to send on reaload.
		//    - PIDFile     string () [/run/prog.pid] - Location of the PID file.
		Option zarray.DefData
	}
)

var (
	system                     SystemIfe
	systemRegistry             []SystemIfe
	ErrNameFieldRequired       = errors.New("config.name field is required")
	ErrNoServiceSystemDetected = errors.New("no service system detected")
	ErrNotAnRootUser           = errors.New("need to execute with sudo permission")
	ErrNotAnAdministrator      = errors.New("please operate with administrator rights")
)

// New creates a new service based on a service interface and configuration
func New(i Ife, c *Config) (ServiceIfe, error) {
	if len(c.Name) == 0 {
		return nil, ErrNameFieldRequired
	}
	if system == nil {
		return nil, ErrNoServiceSystemDetected
	}
	return system.New(i, c)
}

func newSystem() SystemIfe {
	for _, choice := range systemRegistry {
		if !choice.Detect() {
			continue
		}
		return choice
	}
	return nil
}

func chooseSystem(a ...SystemIfe) {
	systemRegistry = a
	system = newSystem()
}

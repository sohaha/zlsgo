// Package daemon program is installed as a system service to achieve process daemon
package daemon

import (
	"errors"
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
	// ServiceIface represents a service that can be run or controlled
	ServiceIface interface {
		Run() error
		Start() error
		Stop() error
		Restart() error
		Install() error
		Uninstall() error
		Status() string
		String() string
	}
	Iface interface {
		Start(s ServiceIface) error
		Stop(s ServiceIface) error
	}
	SystemIface interface {
		String() string
		Interactive() bool
		Detect() bool
		New(i Iface, c *Config) (ServiceIface, error)
	}
	// Config provides the setup for a ServiceIface. The Name field is required.
	Config struct {
		Options     map[string]interface{}
		Name        string
		DisplayName string
		Description string
		UserName    string
		Executable  string
		WorkingDir  string
		RootDir     string
		Arguments   []string
	}
)

var (
	system                     SystemIface
	systemRegistry             []SystemIface
	ErrNameFieldRequired       = errors.New("config.name field is required")
	ErrNoServiceSystemDetected = errors.New("no service system detected")
	ErrNotAnRootUser           = errors.New("need to execute with sudo permission")
	ErrNotAnAdministrator      = errors.New("please operate with administrator rights")
)

// New creates a new service based on a service interface and configuration
func New(i Iface, c *Config) (ServiceIface, error) {
	if len(c.Name) == 0 {
		return nil, ErrNameFieldRequired
	}
	if system == nil {
		return nil, ErrNoServiceSystemDetected
	}
	return system.New(i, c)
}

func newSystem() SystemIface {
	for _, choice := range systemRegistry {
		if !choice.Detect() {
			continue
		}
		return choice
	}
	return nil
}

func chooseSystem(a ...SystemIface) {
	systemRegistry = a
	system = newSystem()
}

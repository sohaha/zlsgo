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

	optionRunWait      = "RunWait"
	optionReloadSignal = "ReloadSignal"
	optionPIDFile      = "PIDFile"
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
		//    - RunAtLoad     bool (false)
		//    - UserService   bool (false) - Install as a current user service.
		//    - SessionCreate bool (false) - Create a full user session.
		//  * POSIX
		//    - RunWait      func() (wait for SIGNAL) - Do not install signal but wait for this function to return.
		//    - ReloadSignal string () [USR1, ...] - Signal to send on reaload.
		//    - PIDFile     string () [/run/prog.pid] - Location of the PID file.
		Option OptionData
	}
	OptionData map[string]interface{}
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

func (kv OptionData) bool(name string, defaultValue bool) bool {
	if v, found := kv[name]; found {
		if castValue, is := v.(bool); is {
			return castValue
		}
	}
	return defaultValue
}

func (kv OptionData) int(name string, defaultValue int) int {
	if v, found := kv[name]; found {
		if castValue, is := v.(int); is {
			return castValue
		}
	}
	return defaultValue
}

func (kv OptionData) string(name string, defaultValue string) string {
	if v, found := kv[name]; found {
		if castValue, is := v.(string); is {
			return castValue
		}
	}
	return defaultValue
}

func (kv OptionData) float64(name string, defaultValue float64) float64 {
	if v, found := kv[name]; found {
		if castValue, is := v.(float64); is {
			return castValue
		}
	}
	return defaultValue
}

func (kv OptionData) funcSingle(name string, defaultValue func()) func() {
	if v, found := kv[name]; found {
		if castValue, is := v.(func()); is {
			return castValue
		}
	}
	return defaultValue
}

func newSystem() SystemIfe {
	for _, choice := range systemRegistry {
		if choice.Detect() == false {
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

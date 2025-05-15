// Package daemon provides functionality for installing and managing programs as system services.
// It supports process daemonization across different operating systems and service managers,
// allowing applications to run in the background as system services.
package daemon

import (
	"context"
	"errors"
)

const (
	// optionKeepAlive is the configuration key for keeping the service alive after exit
	optionKeepAlive        = "KeepAlive"
	optionKeepAliveDefault = true
	// optionRunAtLoad is the configuration key for running the service at system startup
	optionRunAtLoad        = "RunAtLoad"
	optionRunAtLoadDefault = true
	// optionUserService is the configuration key for installing as a user service rather than system service
	optionUserService        = "UserService"
	optionUserServiceDefault = false
	// optionSessionCreate is the configuration key for creating a new session
	optionSessionCreate        = "SessionCreate"
	optionSessionCreateDefault = false
	// optionRunWait is the configuration key for waiting for the service to complete
	optionRunWait = "RunWait"
)

type (
	// ServiceIface represents a service that can be run or controlled.
	// It provides methods for managing the lifecycle of a service, including
	// installation, uninstallation, starting, stopping, and checking status.
	ServiceIface interface {
		// Run starts the service and blocks until the service exits
		Run() error
		// Start starts the service without blocking
		Start() error
		// Stop stops the service
		Stop() error
		// Restart stops and then starts the service
		Restart() error
		// Install installs the service in the system
		Install() error
		// Uninstall removes the service from the system
		Uninstall() error
		// Status returns the current status of the service
		Status() string
		// String returns a string representation of the service
		String() string
	}

	// Iface represents the service implementation that contains the business logic.
	// This interface should be implemented by the application that wants to be run as a service.
	Iface interface {
		// Start is called when the service is started
		Start(s ServiceIface) error
		// Stop is called when the service is stopped
		Stop(s ServiceIface) error
	}

	// SystemIface represents a service management system (like systemd, launchd, etc.)
	// This interface is implemented by platform-specific code to interact with
	// the underlying service management system.
	SystemIface interface {
		// String returns the name of the service system
		String() string
		// Interactive returns whether the system is running in interactive mode
		Interactive() bool
		// Detect checks if this service system is available on the current OS
		Detect() bool
		// New creates a new service for this system
		New(i Iface, c *Config) (ServiceIface, error)
	}

	// Config provides the setup for a ServiceIface. The Name field is required.
	// This structure contains all the configuration needed to install and run a service.
	Config struct {
		// Context is the context for the service
		Context context.Context
		// Options contains platform-specific options
		Options map[string]interface{}
		// Name is the internal name of the service (required)
		Name string
		// DisplayName is the display name of the service shown in service managers
		DisplayName string
		// Description is the description of the service
		Description string
		// UserName is the user to run the service as
		UserName string
		// Executable is the path to the service executable
		Executable string
		// WorkingDir is the working directory for the service
		WorkingDir string
		// RootDir is the root directory for the service
		RootDir string
		// Arguments are the command-line arguments for the service
		Arguments []string
	}
)

var (
	// system is the detected service system for the current platform
	system SystemIface
	// systemRegistry contains all available service systems
	systemRegistry []SystemIface
	// ErrNameFieldRequired is returned when the Name field in Config is empty
	ErrNameFieldRequired = errors.New("config.name field is required")
	// ErrNoServiceSystemDetected is returned when no service system is detected on the current platform
	ErrNoServiceSystemDetected = errors.New("no service system detected")
	// ErrNotAnRootUser is returned when a privileged operation is attempted without root permissions on Unix systems
	ErrNotAnRootUser = errors.New("need to execute with sudo permission")
	// ErrNotAnAdministrator is returned when a privileged operation is attempted without administrator rights on Windows
	ErrNotAnAdministrator = errors.New("please operate with administrator rights")
)

// New creates a new service based on a service interface and configuration.
// It detects the appropriate service system for the current platform and initializes
// a service that can be installed, started, stopped, etc.
func New(i Iface, c *Config) (ServiceIface, error) {
	if len(c.Name) == 0 {
		return nil, ErrNameFieldRequired
	}
	if system == nil {
		return nil, ErrNoServiceSystemDetected
	}
	return system.New(i, c)
}

// newSystem detects and returns the appropriate service system for the current platform.
func newSystem() SystemIface {
	for _, choice := range systemRegistry {
		if !choice.Detect() {
			continue
		}
		return choice
	}
	return nil
}

// chooseSystem registers the provided service systems and selects the appropriate one
// for the current platform. This is called during package initialization to set up
// the available service systems.
func chooseSystem(a ...SystemIface) {
	systemRegistry = a
	system = newSystem()
}

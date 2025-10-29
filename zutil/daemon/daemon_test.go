package daemon

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sohaha/zlsgo"
)

type ss struct {
	I int
}

func (p *ss) Start(s ServiceIface) error {
	p.run()
	return nil
}

func (p *ss) run() {
	fmt.Println("run")
	p.I = p.I + 1
}

func (p *ss) Stop(s ServiceIface) error {
	return nil
}

func TestDaemon(t *testing.T) {
	o := &ss{
		I: 1,
	}
	s, err := New(o, &Config{
		Name:    "zlsgo_daemon_test",
		Options: map[string]interface{}{"UserService": false},
	})
	if err != nil {
		return
	}
	t.Log(o.I)
	t.Log(err)
	_ = s.Install()
	err = s.Start()
	t.Log(err)
	_ = s.Stop()
	_ = s.Restart()
	t.Log(s.Status())
	_ = s.Uninstall()
	t.Log(s.String())
	t.Log(o.I)
}

func TestUtil(t *testing.T) {
	tt := zlsgo.NewTest(t)
	tt.Equal(IsPermissionError(ErrNotAnAdministrator), IsPermissionError(ErrNotAnRootUser))
	_ = isSudo()
}

func TestNew(t *testing.T) {
	tt := zlsgo.NewTest(t)

	o := &ss{I: 1}

	_, err := New(o, &Config{})
	if err == nil {
		t.Error("Expected error for empty name")
	}
	tt.Equal(ErrNameFieldRequired, err)

	_, err = New(o, &Config{Name: "test_service"})
	if err != nil && err != ErrNotAnRootUser {
		tt.NoError(err)
	}
}

func TestNewSystem(t *testing.T) {
	tt := zlsgo.NewTest(t)

	systemRegistry = []SystemIface{}
	sys := newSystem()
	tt.EqualNil(sys)

	mockSystem := &darwinSystem{}
	systemRegistry = []SystemIface{mockSystem}
	sys = newSystem()
	if sys == nil {
		t.Error("Expected system not to be nil")
	}
	tt.Equal(mockSystem, sys)
}

func TestChooseSystem(t *testing.T) {
	tt := zlsgo.NewTest(t)

	originalSystem := system
	originalRegistry := systemRegistry

	defer func() {
		system = originalSystem
		systemRegistry = originalRegistry
	}()

	mockSystem := &darwinSystem{}
	chooseSystem(mockSystem)
	tt.Equal(mockSystem, system)
	tt.Equal([]SystemIface{mockSystem}, systemRegistry)
}

func TestConfig(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tempDir := t.TempDir()
	execPath := filepath.Join(tempDir, "test_exec")

	config := &Config{
		Name:        "test_service",
		DisplayName: "Test Service",
		Description: "Test Description",
		UserName:    "testuser",
		Executable:  execPath,
		WorkingDir:  tempDir,
		RootDir:     "/tmp",
		Arguments:   []string{"arg1", "arg2"},
		Options: map[string]interface{}{
			"KeepAlive":    true,
			"RunAtLoad":    false,
			"UserService":  true,
			"SessionCreate": false,
		},
		Context: context.Background(),
	}

	path := config.execPath()
	tt.Equal(execPath, path)

	config.Executable = ""
	path = config.execPath()
	executablePath, _ := os.Executable()
	tt.Equal(executablePath, path)
}

func TestExecPath(t *testing.T) {
	tt := zlsgo.NewTest(t)

	config := &Config{Executable: "/test/path"}
	path := config.execPath()
	tt.Equal("/test/path", path)

	config = &Config{}
	path = config.execPath()
	expectedPath, _ := os.Executable()
	tt.Equal(expectedPath, path)
}

func TestDarwinSystem(t *testing.T) {
	tt := zlsgo.NewTest(t)

	ds := darwinSystem{}
	tt.Equal("darwin-launchd", ds.String())
	tt.Equal(true, ds.Detect())
	tt.Equal(interactive, ds.Interactive())
}

func TestDarwinLaunchdService(t *testing.T) {
	tt := zlsgo.NewTest(t)

	ds := darwinSystem{}
	o := &ss{I: 1}
	config := &Config{
		Name:    "test_service",
		Options: map[string]interface{}{"UserService": true},
	}

	s, err := ds.New(o, config)
	tt.NoError(err)
	if s == nil {
		t.Error("Expected service not to be nil")
	}

	config.Options = map[string]interface{}{"UserService": false}
	_, err = ds.New(o, config)
	if err == nil {
		t.Error("Expected error for system service without sudo")
	}
}

func TestDarwinLaunchdServiceMethods(t *testing.T) {
	tt := zlsgo.NewTest(t)

	ds := darwinSystem{}
	o := &ss{I: 1}
	config := &Config{
		Name:        "test_service",
		DisplayName: "Test Display Name",
		Options:     map[string]interface{}{"UserService": true},
		Context:     context.Background(),
	}

	s, err := ds.New(o, config)
	tt.NoError(err)

	service := s.(*darwinLaunchdService)

	tt.Equal("Test Display Name", service.String())

	config.DisplayName = ""
	tt.Equal("test_service", service.String())

	homeDir, err := service.getHomeDir()
	tt.NoError(err)
	if homeDir == "" {
		t.Error("Home directory should not be empty")
	}

	servicePath, err := service.getServiceFilePath()
	tt.NoError(err)
	if len(servicePath) == 0 {
		t.Error("Service path should not be empty")
	}
	if !strings.Contains(servicePath, ".plist") {
		t.Error("Service path should contain .plist extension")
	}
	if !strings.Contains(servicePath, "LaunchAgents") {
		t.Error("Service path should contain LaunchAgents for user service")
	}

	service.userService = false
	servicePath, err = service.getServiceFilePath()
	tt.NoError(err)
	if !strings.Contains(servicePath, "LaunchDaemons") {
		t.Error("Service path should contain LaunchDaemons for system service")
	}
}

func TestRunGrep(t *testing.T) {
	tt := zlsgo.NewTest(t)

	_, err := runGrep("world", "echo", "hello world")
	tt.NoError(err)
}

func TestRun(t *testing.T) {
	tt := zlsgo.NewTest(t)

	err := run("echo", "test")
	tt.NoError(err)
}

func TestRuncmd(t *testing.T) {
	tt := zlsgo.NewTest(t)

	commands := []string{"echo", "test"}
	in := bytes.NewReader([]byte(""))
	var out bytes.Buffer
	var outErr bytes.Buffer

	err := runcmd(commands, in, &out, &outErr)
	tt.NoError(err)
	tt.Equal("test\n", out.String())
}

func TestIsServiceRestart(t *testing.T) {
	tt := zlsgo.NewTest(t)

	config := &Config{Options: map[string]interface{}{"RunAtLoad": true}}
	result := isServiceRestart(config)
	tt.Equal(true, result)

	config.Options = map[string]interface{}{"RunAtLoad": false}
	result = isServiceRestart(config)
	tt.Equal(false, result)

	config.Options = map[string]interface{}{}
	result = isServiceRestart(config)
	tt.Equal(true, result)

	config.Options = map[string]interface{}{"RunAtLoad": "true"}
	result = isServiceRestart(config)
	tt.Equal(true, result)
}
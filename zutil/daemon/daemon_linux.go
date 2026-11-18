package daemon

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type (
	linuxSystemService struct {
		detect      func() bool
		interactive func() bool
		new         func(i Iface, c *Config) (ServiceIface, error)
		name        string
	}
	systemd struct {
		i Iface
		*Config
	}
)

const (
	optionReloadSignal = "ReloadSignal"
	optionPIDFile      = "PIDFile"
)

func init() {
	chooseSystem(linuxSystemService{
		name:   "linux-systemd",
		detect: isSystemd,
		interactive: func() bool {
			is, _ := isInteractive()
			return is
		},
		new: newSystemdService,
	})
}

func (sc linuxSystemService) String() string {
	return sc.name
}

func (sc linuxSystemService) Detect() bool {
	return sc.detect()
}

func (sc linuxSystemService) Interactive() bool {
	return sc.interactive()
}

func (sc linuxSystemService) New(i Iface, c *Config) (s ServiceIface, err error) {
	s, err = sc.new(i, c)
	if err != nil {
		return nil, err
	}
	userService := optionUserServiceDefault
	if u, ok := c.Options[optionUserService]; ok {
		userService, _ = u.(bool)
	}
	if !userService {
		if err = isSudo(); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func isInteractive() (bool, error) {
	return os.Getppid() != 1, nil
}

var tf = map[string]interface{}{
	"cmd": func(s string) string {
		return `"` + strings.Replace(s, `"`, `\"`, -1) + `"`
	},
	"cmdEscape": func(s string) string {
		return strings.Replace(s, " ", `\x20`, -1)
	},
}

func isSystemd() bool {
	if _, err := os.Stat("/run/systemd/system"); err == nil {
		return true
	}
	return false
}

func newSystemdService(i Iface, c *Config) (ServiceIface, error) {
	if c.Context == nil {
		c.Context = context.Background()
	}

	s := &systemd{
		i:      i,
		Config: c,
	}

	return s, nil
}

func (s *systemd) String() string {
	if len(s.DisplayName) > 0 {
		return s.DisplayName
	}
	return s.Name
}

func (s *systemd) configPath() (cp string, err error) {
	userService := optionUserServiceDefault
	if u, ok := s.Options[optionUserService]; ok {
		userService, _ = u.(bool)
	}
	if userService {
		dir := os.Getenv("XDG_CONFIG_HOME")
		if dir == "" {
			home, herr := os.UserHomeDir()
			if herr != nil {
				return "", herr
			}
			dir = filepath.Join(home, ".config")
		}
		return filepath.Join(dir, "systemd", "user", s.Config.Name+".service"), nil
	}
	cp = "/etc/systemd/system/" + s.Config.Name + ".service"
	return
}

func (s *systemd) template() *template.Template {
	return template.Must(template.New("").Funcs(tf).Parse(systemdScript))
}

func (s *systemd) Install() error {
	confPath, err := s.configPath()
	if err != nil {
		return err
	}
	_, err = os.Stat(confPath)
	if err == nil {
		return fmt.Errorf("init already exists: %s", confPath)
	}

	userService := optionUserServiceDefault
	if u, ok := s.Options[optionUserService]; ok {
		userService, _ = u.(bool)
	}
	if userService {
		if err = os.MkdirAll(filepath.Dir(confPath), 0700); err != nil {
			return err
		}
	}

	f, err := os.Create(confPath)
	if err != nil {
		return err
	}
	defer f.Close()
	reloadSignal := ""
	if v, ok := s.Options[optionReloadSignal]; ok {
		reloadSignal, _ = v.(string)
	}
	pidFile := ""
	if v, ok := s.Options[optionPIDFile]; ok {
		pidFile, _ = v.(string)
	}
	path := s.execPath()
	to := &struct {
		*Config
		Path         string
		ReloadSignal string
		PIDFile      string
		UserService  bool
		StartDelay   int
	}{
		s.Config,
		path,
		reloadSignal,
		pidFile,
		userService,
		startDelaySeconds(s.Config),
	}

	err = s.template().Execute(f, to)
	if err != nil {
		return err
	}

	if userService {
		if err = run("systemctl", "--user", "enable", s.Name+".service"); err != nil {
			return err
		}
		return run("systemctl", "--user", "daemon-reload")
	}

	err = run("systemctl", "enable", s.Name+".service")
	if err != nil {
		return err
	}
	return run("systemctl", "daemon-reload")
}

func (s *systemd) Uninstall() error {
	userService := optionUserServiceDefault
	if u, ok := s.Options[optionUserService]; ok {
		userService, _ = u.(bool)
	}
	if userService {
		_ = run("systemctl", "--user", "stop", s.Name+".service")
		if err := run("systemctl", "--user", "disable", s.Name+".service"); err != nil {
			return err
		}
	} else {
		_ = run("systemctl", "stop", s.Name+".service")
		if err := run("systemctl", "disable", s.Name+".service"); err != nil {
			return err
		}
	}
	cp, err := s.configPath()
	if err != nil {
		return err
	}
	if err := os.Remove(cp); err != nil {
		return err
	}
	return nil
}

func (s *systemd) Run() (err error) {
	err = s.i.Start(s)
	if err != nil {
		return err
	}

	runWait := func() {
		select {
		case <-SingleKillSignal():
		case <-s.Config.Context.Done():
		}
	}
	if v, ok := s.Options[optionRunWait]; ok {
		runWait, _ = v.(func())
	}

	runWait()

	return s.i.Stop(s)
}

func (s *systemd) Start() error {
	userService := optionUserServiceDefault
	if u, ok := s.Options[optionUserService]; ok {
		userService, _ = u.(bool)
	}
	if userService {
		return run("systemctl", "--user", "start", s.Name+".service")
	}
	if os.Getuid() == 0 {
		return run("systemctl", "start", s.Name+".service")
	}
	return run("sudo", "-n", "systemctl", "start", s.Name+".service")
}

func (s *systemd) Stop() error {
	userService := optionUserServiceDefault
	if u, ok := s.Options[optionUserService]; ok {
		userService, _ = u.(bool)
	}
	if userService {
		return run("systemctl", "--user", "stop", s.Name+".service")
	}
	if os.Getuid() == 0 {
		return run("systemctl", "stop", s.Name+".service")
	}
	return run("sudo", "-n", "systemctl", "stop", s.Name+".service")
}

func (s *systemd) Restart() error {
	userService := optionUserServiceDefault
	if u, ok := s.Options[optionUserService]; ok {
		userService, _ = u.(bool)
	}
	if userService {
		return run("systemctl", "--user", "restart", s.Name+".service")
	}
	if os.Getuid() == 0 {
		return run("systemctl", "restart", s.Name+".service")
	}
	return run("sudo", "-n", "systemctl", "restart", s.Name+".service")
}

func (s *systemd) Status() string {
	userService := optionUserServiceDefault
	if u, ok := s.Options[optionUserService]; ok {
		userService, _ = u.(bool)
	}
	var res string
	if userService {
		res, _ = runGrep("running", "systemctl", "--user", "status", s.Name+".service")
	} else if os.Getuid() == 0 {
		res, _ = runGrep("running", "systemctl", "status", s.Name+".service")
	} else {
		res, _ = runGrep("running", "sudo", "-n", "systemctl", "status", s.Name+".service")
	}
	if res != "" {
		return "Running"
	}
	return "Stop"
}

const systemdScript = `[Unit]
Description={{.Description}}
ConditionFileIsExecutable={{.Path|cmdEscape}}

[Service]
{{if gt .StartDelay 0}}ExecStartPre=/bin/sleep {{.StartDelay}}
{{end}}ExecStart={{.Path|cmdEscape}}{{range .Arguments}} {{.|cmd}}{{end}}
{{if .RootDir}}RootDirectory={{.RootDir|cmd}}{{end}}
{{if .WorkingDir}}WorkingDirectory={{.WorkingDir|cmdEscape}}{{end}}
{{if .UserName}}User={{.UserName}}{{end}}
{{if .ReloadSignal}}ExecReload=/bin/kill -{{.ReloadSignal}} "$MAINPID"{{end}}
{{if .PIDFile}}PIDFile={{.PIDFile|cmd}}{{end}}
Restart=on-failure
RestartSec=1
{{if not .UserService}}EnvironmentFile=-/etc/sysconfig/{{.Name}}
{{end}}KillMode=process

[Install]
WantedBy={{if .UserService}}default.target{{else}}multi-user.target{{end}}
`

package daemon

import (
	"errors"
	"fmt"
	"os"
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

var errNoUserServiceSystemd = errors.New("user services are not supported on systemd")

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
	if err == nil {
		err = isSudo()
	}
	return
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
	if u, ok := s.Option[optionUserService]; ok {
		userService = u.(bool)
	}
	if userService {
		err = errNoUserServiceSystemd
		return
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

	f, err := os.Create(confPath)
	if err != nil {
		return err
	}
	defer f.Close()
	reloadSignal := ""
	if v, ok := s.Option[optionReloadSignal]; ok {
		reloadSignal, _ = v.(string)
	}
	pidFile := ""
	if v, ok := s.Option[optionPIDFile]; ok {
		pidFile, _ = v.(string)
	}
	path := s.execPath()
	var to = &struct {
		*Config
		Path         string
		ReloadSignal string
		PIDFile      string
	}{
		s.Config,
		path,
		reloadSignal,
		pidFile,
	}

	err = s.template().Execute(f, to)
	if err != nil {
		return err
	}

	err = run("systemctl", "enable", s.Name+".service")
	if err != nil {
		return err
	}
	return run("systemctl", "daemon-reload")
}

func (s *systemd) Uninstall() error {
	_ = run("systemctl", "stop", s.Name+".service")
	err := run("systemctl", "disable", s.Name+".service")
	if err != nil {
		return err
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
		<-SingleKillSignal()
	}
	if v, ok := s.Option[optionRunWait]; ok {
		runWait, _ = v.(func())
	}

	runWait()

	return s.i.Stop(s)
}

func (s *systemd) Start() error {
	if os.Getuid() == 0 {
		return run("systemctl", "start", s.Name+".service")
	} else {
		return run("sudo", "-n", "systemctl", "start", s.Name+".service")
	}
}

func (s *systemd) Stop() error {
	if os.Getuid() == 0 {
		return run("systemctl", "stop", s.Name+".service")
	} else {
		return run("sudo", "-n", "systemctl", "stop", s.Name+".service")
	}
}

func (s *systemd) Restart() error {
	if os.Getuid() == 0 {
		return run("systemctl", "restart", s.Name+".service")
	} else {
		return run("sudo", "-n", "systemctl", "restart", s.Name+".service")
	}
}

func (s *systemd) Status() string {
	var res string
	if os.Getuid() == 0 {
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
StartLimitInterval=5
StartLimitBurst=10
ExecStart={{.Path|cmdEscape}}{{range .Arguments}} {{.|cmd}}{{end}}
{{if .RootDir}}RootDirectory={{.RootDir|cmd}}{{end}}
{{if .WorkingDir}}WorkingDirectory={{.WorkingDir|cmdEscape}}{{end}}
{{if .UserName}}User={{.UserName}}{{end}}
{{if .ReloadSignal}}ExecReload=/bin/kill -{{.ReloadSignal}} "$MAINPID"{{end}}
{{if .PIDFile}}PIDFile={{.PIDFile|cmd}}{{end}}
Restart=always
RestartSec=120ms
EnvironmentFile=-/etc/sysconfig/{{.Name}}
KillMode=process

[Install]
WantedBy=multi-user.target
`

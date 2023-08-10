package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

type (
	freebsdRcdService struct {
		i Iface
		*Config
		userService bool
	}
	freebsdSystem struct{}
)

const version = "freebsd-rcd"

var interactive = false

func init() {
	var err error
	chooseSystem(freebsdSystem{})
	interactive, err = isInteractive()
	if err != nil {
		panic(err)
	}
}

func (freebsdSystem) String() string {
	return version
}

func (freebsdSystem) Detect() bool {
	return true
}

func (freebsdSystem) Interactive() bool {
	return interactive
}

func (freebsdSystem) New(i Iface, c *Config) (ServiceIface, error) {
	userService := optionUserServiceDefault
	if s, ok := c.Options[optionUserService]; ok {
		userService, _ = s.(bool)
	}
	s := &freebsdRcdService{
		i:           i,
		Config:      c,
		userService: userService,
	}

	return s, nil
}

func isInteractive() (bool, error) {
	return os.Getppid() != 1, nil
}

func (s *freebsdRcdService) Status() string {
	return "Unknown"
}

func (s *freebsdRcdService) String() string {
	if len(s.DisplayName) > 0 {
		return s.DisplayName
	}
	return s.Name
}

func (s *freebsdRcdService) getServiceFilePath() (string, error) {

	return "/etc/rc.d/" + s.Name, nil
}

func (s *freebsdRcdService) Install() error {
	confPath, err := s.getServiceFilePath()
	if err != nil {
		return err
	}
	_, err = os.Stat(confPath)
	if err == nil {
		return fmt.Errorf("init already exists: %s", confPath)
	}

	if s.userService {
		//  ~/Library/LaunchAgents exists.
		err = os.MkdirAll(filepath.Dir(confPath), 0700)
		if err != nil {
			return err
		}
	}

	f, err := os.Create(confPath)
	if err != nil {
		return err
	}
	defer f.Close()

	keepAlive := optionKeepAliveDefault
	if v, ok := s.Options[optionKeepAlive]; ok {
		keepAlive, _ = v.(bool)
	}
	load := optionRunAtLoadDefault
	if v, ok := s.Options[optionRunAtLoad]; ok {
		load, _ = v.(bool)
	}
	sessionCreate := optionSessionCreateDefault
	if v, ok := s.Options[optionSessionCreate]; ok {
		sessionCreate, _ = v.(bool)
	}

	path := s.execPath()
	to := &struct {
		*Config
		Path string

		KeepAlive, RunAtLoad bool
		SessionCreate        bool
	}{
		Config:        s.Config,
		Path:          path,
		KeepAlive:     keepAlive,
		RunAtLoad:     load,
		SessionCreate: sessionCreate,
	}

	functions := template.FuncMap{
		"bool": func(v bool) string {
			if v {
				return "true"
			}
			return "false"
		},
	}

	rcdScript := ""
	if s.Name == "opsramp-agent" {
		rcdScript = rcdScriptOpsrampAgent
		file, err := os.OpenFile("/etc/rc.conf", os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		defer file.Close()
		data := "opsramp_agent_enable=" + `"` + "YES" + `"`
		_, _ = fmt.Fprintln(file, data)

	} else if s.Name == "opsramp-shield" {
		rcdScript = rcdScriptOpsrampShield
		file, err := os.OpenFile("/etc/rc.conf", os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		defer file.Close()
		data := "opsramp_shield_enable=" + `"` + "YES" + `"`
		_, _ = fmt.Fprintln(file, data)

	} else {
		rcdScript = rcdScriptAgentUninstall
		file, err := os.OpenFile("/etc/rc.conf", os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Printf("failed opening file: %s\n", err)
		}
		defer file.Close()
		data := "agent_uninstall_enable=" + `"` + "YES" + `"`
		_, _ = fmt.Fprintln(file, data)
	}

	t := template.Must(template.New("rcdScript").Funcs(functions).Parse(rcdScript))
	errExecute := t.Execute(f, to)

	serviceName := "/etc/rc.d/" + s.Name
	err = os.Chmod(serviceName, 755)
	if err != nil {
		return err
	}

	return errExecute
}

func (s *freebsdRcdService) Uninstall() error {
	_ = s.Stop()
	if s.Name == "opsramp-agent" {
		_ = run("sed", "-i", "-e", "'/opsramp_agent_enable/d'", "/etc/rc.conf")
	} else if s.Name == "opsramp-shield" {
		_ = run("sed", "-i", "-e", "'/opsramp_shield_enable/d'", "/etc/rc.conf")
	} else {
		_ = run("sed", "-i", "-e", "'/agent_uninstall_enable/d'", "/etc/rc.conf")
	}
	confPath, err := s.getServiceFilePath()
	if err != nil {
		return err
	}
	return os.Remove(confPath)
}

func (s *freebsdRcdService) Start() error {
	return run("service", s.Name, "start")
}
func (s *freebsdRcdService) Stop() error {
	return run("service", s.Name, "stop")

}
func (s *freebsdRcdService) Restart() error {
	return run("service", s.Name, "restart")

}

func (s *freebsdRcdService) Run() error {
	var err error

	err = s.i.Start(s)
	if err != nil {
		return err
	}
	runWait := func() {
		<-SingleKillSignal()
	}
	if v, ok := s.Options[optionRunWait]; ok {
		runWait, _ = v.(func())
	}
	runWait()
	return s.i.Stop(s)
}

const rcdScriptOpsrampAgent = `. /etc/rc.subr

name="opsramp_agent"
rcvar="opsramp_agent_enable"
command="/opt/opsramp/agent/opsramp-agent service"
pidfile="/var/run/${name}.pid"

start_cmd="test_start"
stop_cmd="test_stop"
status_cmd="test_status"

test_start() {
        /usr/sbin/daemon -p ${pidfile} ${command}
}

test_status() {
        if [ -e ${pidfile} ]; then
                echo ${name} is running...
        else
                echo ${name} is not running.
        fi
}

test_stop() {
        if [ -e ${pidfile} ]; then
` +
	" kill `cat ${pidfile}`; " +
	`
        else
                echo ${name} is not running?
        fi
}

load_rc_config $name
run_rc_command "$1"
`

const rcdScriptOpsrampShield = `. /etc/rc.subr

name="opsramp_shield"
rcvar="opsramp_shield_enable"
command="/opt/opsramp/agent/bin/opsramp-shield service"
pidfile="/var/run/${name}.pid"

start_cmd="test_start"
stop_cmd="test_stop"
status_cmd="test_status"

test_start() {
        /usr/sbin/daemon -p ${pidfile} ${command}
}

test_status() {
        if [ -e ${pidfile} ]; then
                echo ${name} is running...
        else
                echo ${name} is not running.
        fi
}

test_stop() {
        if [ -e ${pidfile} ]; then
` +
	" kill `cat ${pidfile}`; " +
	`
        else
                echo ${name} is not running?
        fi
}

load_rc_config $name
run_rc_command "$1"
`
const rcdScriptAgentUninstall = `. /etc/rc.subr

name="agent_uninstall"
rcvar="agent_uninstall_enable"
command="/opt/opsramp/agent/bin/uninstall"
pidfile="/var/run/${name}.pid"

start_cmd="test_start"
stop_cmd="test_stop"
status_cmd="test_status"

test_start() {
        /usr/sbin/daemon -p ${pidfile} ${command}
}

test_status() {
        if [ -e ${pidfile} ]; then
                echo ${name} is running...
        else
                echo ${name} is not running.
        fi
}

test_stop() {
        if [ -e ${pidfile} ]; then
` +
	" kill `cat ${pidfile}`; " +
	`
        else
                echo ${name} is not running?
        fi
}

load_rc_config $name
run_rc_command "$1"
`

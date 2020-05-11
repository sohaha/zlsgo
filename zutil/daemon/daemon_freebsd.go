package daemon

import (
	"fmt"
	"io/ioutil"
	"log/syslog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"text/template"
)

type (
	freebsdRcdService struct {
		i Ife
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

func (freebsdSystem) New(i Ife, c *Config) (ServiceIfe, error) {
	s := &freebsdRcdService{
		i:      i,
		Config: c,

		userService: c.Option.Bool(optionUserService, optionUserServiceDefault),
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

	path := s.execPath()
	to := &struct {
		*Config
		Path string

		KeepAlive, RunAtLoad bool
		SessionCreate        bool
	}{
		Config:        s.Config,
		Path:          path,
		KeepAlive:     s.Option.Bool(optionKeepAlive, optionKeepAliveDefault),
		RunAtLoad:     s.Option.Bool(optionRunAtLoad, optionRunAtLoadDefault),
		SessionCreate: s.Option.Bool(optionSessionCreate, optionSessionCreateDefault),
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
			fmt.Println("failed opening file: %s", err)
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

	s.Option.FuncSingle(optionRunWait, func() {
		var sigChan = make(chan os.Signal, 3)
		signal.Notify(sigChan, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
		<-sigChan
	})()

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

func newSysLogger(name string, errs chan<- error) (Logger, error) {
	w, err := syslog.New(syslog.LOG_INFO, name)
	if err != nil {
		return nil, err
	}
	return sysLogger{w, errs}, nil
}

type sysLogger struct {
	*syslog.Writer
	errs chan<- error
}

func (s sysLogger) send(err error) error {
	if err != nil && s.errs != nil {
		s.errs <- err
	}
	return err
}

func (s sysLogger) Error(v ...interface{}) error {
	return s.send(s.Writer.Err(fmt.Sprint(v...)))
}

func (s sysLogger) Warning(v ...interface{}) error {
	return s.send(s.Writer.Warning(fmt.Sprint(v...)))
}

func (s sysLogger) Info(v ...interface{}) error {
	return s.send(s.Writer.Info(fmt.Sprint(v...)))
}

func (s sysLogger) Errorf(format string, a ...interface{}) error {
	return s.send(s.Writer.Err(fmt.Sprintf(format, a...)))
}

func (s sysLogger) Warningf(format string, a ...interface{}) error {
	return s.send(s.Writer.Warning(fmt.Sprintf(format, a...)))
}

func (s sysLogger) Infof(format string, a ...interface{}) error {
	return s.send(s.Writer.Info(fmt.Sprintf(format, a...)))
}

func run(command string, arguments ...string) error {
	cmd := exec.Command(command, arguments...)
	stderr, err := cmd.StderrPipe()

	if err != nil {
		return fmt.Errorf("%q failed to connect stderr pipe: %v", command, err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("%q failed: %v", command, err)
	}

	if command == "launchctl" {
		slurp, _ := ioutil.ReadAll(stderr)
		if len(slurp) > 0 {
			return fmt.Errorf("%q failed with stderr: %s", command, slurp)
		}
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("%q failed: %v", command, err)
	}

	return nil
}

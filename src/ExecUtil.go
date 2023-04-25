package src

import (
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"time"
)

func convertPath(path string) string {
	index := strings.Index(path, ":")
	if index > 0 {
		s := path[0:index]
		s2 := path[index+1:]
		return "/" + strings.ToLower(s) + strings.ReplaceAll(s2, "\\", "/")
	} else {
		return path
	}
}

func getCmd(pwd string, engine string, command string, logPath string) (*exec.Cmd, error) {
	var cmd *exec.Cmd

	if len(engine) == 0 {
		spaceRe, _ := regexp.Compile(`\s+`)
		split := spaceRe.Split(command, -1)
		cmd = exec.Command(split[0], split[1:]...)
		cmd.Dir = pwd
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		return cmd, nil
	}
	if isCmd(engine) {
		if len(logPath) > 0 {
			command = command + ">" + logPath
		}
		cmd = exec.Command(engine, "/c", command)
		cmd.Dir = pwd
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		return cmd, nil
	}
	if isBash(engine) {
		if len(logPath) > 0 {
			newPath := convertPath(logPath)
			command = command + ">" + newPath
		}

		cmd = exec.Command(engine, "-c", command)
		cmd.Dir = pwd
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		return cmd, nil
	}
	cmd = exec.Command(engine)
	cmd.Dir = pwd
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	pipe, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	commandLine := command + "\nexit\n"
	pipe.Write([]byte(commandLine))
	return cmd, nil
}

func BatchExec(pwd string, engine string, command string, fileLog *TypeFileLog) {
	cmd, err := getCmd(pwd, engine, command, "")
	if err != nil {
		ErrorDialog(err)
		return
	}

	cmd.Wait()

	output, err := cmd.CombinedOutput()
	if err != nil {
		ErrorDialog(err)
		return
	}

	fileLog.mutex.Lock()
	fileLog.file.WriteString(pwd + ">" + command + "\n")
	fileLog.file.Write(output)
	fileLog.mutex.Unlock()
}

func Exec(pwd string, engine string, command string, logPath string) {
	cmd, err := getCmd(pwd, engine, command, logPath)
	if err != nil {
		ErrorDialog(err)
		return
	}
	cmd.Start()
	cmd.Wait()
}

func Launcher(pwd string, engine string, command string) {
	before := time.Now()
	ShowProgress()
	cmd, err := getCmd(pwd, engine, command, "")
	if err != nil {
		HideProgress()
		ErrorDialog(err)
		return
	}

	err = cmd.Start()
	if err != nil {
		HideProgress()
		ErrorDialog(err)
		return
	}
	after := time.Now()
	duration := after.Sub(before)
	t := 500 - duration.Milliseconds()
	if t > 0 {
		time.Sleep(time.Duration(t) * time.Millisecond)
	}
	HideProgress()
}

func isCmd(engine string) bool {
	l := len(engine)
	if strings.Index(engine, "cmd.exe") == l-7 {
		return true
	}
	if strings.Index(engine, "cmd") == l-3 {
		return true
	}
	return false
}

func isBash(engine string) bool {
	l := len(engine)
	if strings.Index(engine, "bash.exe") == l-8 {
		return true
	}
	return false
}

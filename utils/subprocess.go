package utils

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"
	"unilab-backend/logging"
	"unilab-backend/setting"
)

type Response struct {
	StdOut    string
	StdErr    string
	ServerErr string
	ExitCode  int
}

func getErrorExitCode(err error) int {
	// fail, non-zero exit status conditions
	if exitError, ok := err.(*exec.ExitError); ok {
		return exitError.Sys().(syscall.WaitStatus).ExitStatus()
	}
	// fails that do not define an exec.ExitError (e.g. unable to identify executable on system PATH)
	return 1 // assign a default non-zero fail code value of 1
}

func Subprocess(rlimit string, timeout int, executable string, pwd string, args ...string) Response {
	var res Response
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout+5)*time.Second)
	defer cancel()
	var cmdarray []string
	if rlimit != "" {
		cmdarray = append([]string{fmt.Sprintf("%s && %s", rlimit, executable)}, args...)
	} else {
		cmdarray = append([]string{executable}, args...)
	}
	logging.Info("Exec Command: bash -c " + strings.Join(cmdarray, " "))
	cmd := exec.CommandContext(ctx, "bash", "-c", strings.Join(cmdarray, " "))
	if pwd != "" {
		cmd.Dir = pwd
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logging.Info(err)
		res.ServerErr = err.Error()
		res.StdErr = ""
		res.StdOut = ""
		res.ExitCode = getErrorExitCode(err)
		return res
	}
	defer stdout.Close()
	stderr, err := cmd.StderrPipe()
	if err != nil {
		logging.Info(err)
		res.ServerErr = err.Error()
		res.StdErr = ""
		res.StdOut = ""
		res.ExitCode = getErrorExitCode(err)
		return res
	}
	defer stderr.Close()
	outContent := ""
	errContent := ""
	err = cmd.Start()
	if err != nil {
		logging.Info("start command error: ", err)
		res.ServerErr = err.Error()
		res.StdErr = ""
		res.StdOut = ""
		res.ExitCode = 1
		return res
	}
	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)
	if setting.RunMode == "debug" {
		for scanner.Scan() {
			outContent += scanner.Text() + "\n"
		}
		logging.Info("stdout: ", outContent)
	} else {
		linecounter := 0
		for scanner.Scan() && linecounter < 50 {
			outContent += scanner.Text() + "\n"
			linecounter++
		}
	}
	scanner = bufio.NewScanner(stderr)
	scanner.Split(bufio.ScanLines)
	if setting.RunMode == "debug" {
		for scanner.Scan() {
			errContent += scanner.Text() + "\n"
		}
		logging.Info("stderr: ", errContent)
	} else {
		linecounter := 0
		for scanner.Scan() && linecounter < 50 {
			errContent += scanner.Text() + "\n"
			linecounter++
		}
	}
	_stdout := strings.Trim(outContent, " \n")
	_stderr := strings.Trim(errContent, " \n")
	err = cmd.Wait()
	if err != nil {
		logging.Info("server: ", err)
		res.StdOut = _stdout
		res.StdErr = _stderr
		res.ServerErr = err.Error()
		res.ExitCode = getErrorExitCode(err)
		return res
	}
	res.StdOut = _stdout
	res.StdErr = _stderr
	res.ServerErr = ""
	res.ExitCode = cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
	return res
}

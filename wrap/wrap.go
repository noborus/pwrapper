package wrap

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/kballard/go-shellquote"
)

type PWrapper struct {
	WrapCommand string
	Start       string
	End         string
	ExecCommand []string
}

type wCommand struct {
	cmd        *exec.Cmd
	commands   []string
	start      string
	end        string
	stdin      io.WriteCloser
	stdoutChan chan string
	stderr     io.ReadCloser
}

func Open(command string, start string, end string) (*wCommand, error) {
	commands, err := shellquote.Split(command)
	if err != nil {
		return nil, err
	}
	w := &wCommand{
		commands: commands,
		start:    start,
		end:      end,
	}
	if len(commands) == 0 {
		return nil, fmt.Errorf("no commands")
	}
	w.cmd = exec.Command(w.commands[0], w.commands[1:]...)
	stdin, err := w.cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	w.stdin = stdin

	stdout, err := w.cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stdoutChan := make(chan string)
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			stdoutChan <- scanner.Text()
		}
	}()
	w.stdoutChan = stdoutChan

	stderr, err := w.cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	w.stderr = stderr

	if err = w.cmd.Start(); err != nil {
		return nil, err
	}

	if err = w.Write(w.start); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *wCommand) Close() error {
	if err := w.Write(w.end); err != nil {
		return err
	}
	if err := w.stdin.Close(); err != nil {
		return err
	}
	if err := w.cmd.Wait(); err != nil {
		return err
	}
	close(w.stdoutChan)
	return nil
}

func (w *wCommand) Write(str string) error {
	if _, err := io.WriteString(w.stdin, str); err != nil {
		return err
	}
	if _, err := io.WriteString(w.stdin, "\n"); err != nil {
		return err
	}
	return nil
}

func Command(p PWrapper) error {
	w, err := Open(p.WrapCommand, p.Start, p.End)
	if err != nil {
		return err
	}

	go func() {
		_, err := io.Copy(os.Stderr, w.stderr)
		if err != nil {
			fmt.Printf("(stderr) %s\n", err)
		}
	}()

	for _, cmd := range p.ExecCommand {
		stdout := <-w.stdoutChan
		command := strings.Replace(cmd, "$args", stdout, 1)
		commands, err := shellquote.Split(command)
		if err != nil {
			return err
		}
		out, err := exec.Command(commands[0], commands[1:]...).Output()
		if err != nil {
			return fmt.Errorf("exec.Command: %v:%w", commands, err)
		}
		str := string(out)
		for _, r := range strings.Split(str, "\n") {
			if strings.HasPrefix(r, "PWRAPPER:") {
				r = strings.Replace(r, "PWRAPPER:", "", 1)
				if err := w.Write(r); err != nil {
					return err
				}
				fmt.Printf("(pipe) %s\n", r)
			} else {
				fmt.Printf("(stdout) %s\n", r)
			}
		}
	}

	go func() {
		for stdout := range w.stdoutChan {
			fmt.Printf("(end) %s\n", stdout)
		}
	}()
	return w.Close()
}

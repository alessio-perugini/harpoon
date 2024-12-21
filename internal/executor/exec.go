package executor

import (
	"fmt"
	"io"
	"os/exec"
	"sync"
)

func Run(cmd []string, stdoutCB, stderrCB func(r io.Reader)) error {
	command := exec.Command(cmd[0], cmd[1:]...)

	stdout, err := command.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err2 := command.StderrPipe()
	if err2 != nil {
		return err2
	}

	if err := command.Start(); err != nil {
		return err
	}

	wg := &sync.WaitGroup{}
	if stdoutCB != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			stdoutCB(stdout)
		}()

	}
	if stderrCB != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			stderrCB(stderr)
		}()
	}
	defer wg.Wait()

	// Wait return error if the process has terminated with exit code > 0.
	// In those cases we don't want to return an error
	if err := command.Wait(); err != nil && command.ProcessState.ExitCode() == 0 {
		return err
	}
	return nil
}

func Build(packagePath, outputFile string) (string, error) {
	cmd := exec.Command(
		"go",
		"test",
		"-gcflags=-N -l",  // disable optimization
		"-c", packagePath, // build test binary
		"-o", outputFile, // save it in a dedicated directory
	)

	stdout, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to execute build command '%s': %v", cmd.String(), err)
	}

	return string(stdout), nil
}

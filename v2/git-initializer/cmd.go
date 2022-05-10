package main

import (
	"bufio"
	"log"
	"os/exec"
	"sync"

	"github.com/pkg/errors"
)

// execCommand executes a commands and pipes its stdout and stderr out to the
// git-initializer's own logs.
func execCommand(cmd *exec.Cmd) error {
	stdoutReader, err := cmd.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "error obtaining stdout pipe")
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdoutReader)
		for scanner.Scan() {
			log.Println(scanner.Text())
		}
	}()
	stderrReader, err := cmd.StderrPipe()
	if err != nil {
		return errors.Wrap(err, "error obtaining stderr pipe")
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderrReader)
		for scanner.Scan() {
			log.Println(scanner.Text())
		}
	}()
	if err = cmd.Start(); err != nil {
		return errors.Wrapf(err, "error starting command %q", cmd.String())
	}
	if err = cmd.Wait(); err != nil {
		return errors.Wrapf(err, "error waiting for command %q", cmd.String())
	}
	wg.Wait() // Make sure we got all the output before returning
	return nil
}

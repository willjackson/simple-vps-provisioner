package utils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// RunCommand executes a shell command and returns output, error
func RunCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)

	if os.Getenv("DEBUG") == "1" {
		fmt.Printf("[DEBUG] Running: %s %s\n", name, strings.Join(args, " "))
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return stdout.String(), fmt.Errorf("%v: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// RunCommandWithInput executes a command with stdin input
func RunCommandWithInput(input, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Stdin = strings.NewReader(input)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return stdout.String(), fmt.Errorf("%v: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// RunShell executes a shell command via bash -c
func RunShell(command string) (string, error) {
	return RunCommand("bash", "-c", command)
}

// CommandExists checks if a command is available in PATH
func CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// MustRunCommand runs a command and exits the program on error
func MustRunCommand(name string, args ...string) string {
	output, err := RunCommand(name, args...)
	if err != nil {
		Err("Command failed: %s %s", name, strings.Join(args, " "))
		Err("Error: %v", err)
		os.Exit(1)
	}
	return output
}

package command

import (
	"bytes"
	"io"
	"os/exec"
	"path"
	"regexp"
	"runtime"
	"testing"
)

func TestStartWatching(t *testing.T) {
	var output bytes.Buffer
	var input bytes.Buffer

	p := PasswordEnterBot{
		ReadFrom: &input,
		WriteTo:  &output,
		Prompt:   regexp.MustCompile("PASSWORD:"),
		Password: "hunter2",
	}

	p.StartWatching()

	input.WriteString("Welcome to some system\n")
	input.WriteString("This system requires authentication\n")
	input.WriteString("PASSWORD: ")

	// Wait for the password to have been entered
	err := p.Wait()

	if err != nil {
		t.Fatal(err)
	}

	password, err := output.ReadString('\n')

	if err != nil {
		t.Fatal(err)
	}

	if password != (p.Password + "\n") {
		t.Errorf("Password was'%v', expected '%v'", password, p.Password)
	}
}

// Test the password watcher with a real script
func TestStartWatchingReal(t *testing.T) {
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("Skipping because bash does not exist")
	}

	_, filename, _, _ := runtime.Caller(0)
	scriptPath := path.Join(path.Dir(filename), "test/password_prompt.bash")

	cmd := exec.Command("bash", scriptPath)

	var stdinPipe io.WriteCloser
	var stdoutPipe io.ReadCloser
	var stderrPipe io.ReadCloser
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	var err error

	stdinPipe, err = cmd.StdinPipe()

	if err != nil {
		t.Fatal(err)
	}

	stdoutPipe, err = cmd.StdoutPipe()

	if err != nil {
		t.Fatal(err)
	}

	stderrPipe, err = cmd.StderrPipe()

	if err != nil {
		t.Fatal(err)
	}

	p := PasswordEnterBot{
		ReadFrom: stdoutPipe,
		WriteTo:  stdinPipe,
		Prompt:   regexp.MustCompile("Password:"),
		Password: "hunter2",
	}

	go stderr.ReadFrom(stderrPipe)

	// Start watching on the pipes
	p.StartWatching()

	// Start the command
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	// Wait until the password has been entered
	err = p.Wait()

	if err != nil {
		t.Fatal(err)
	}

	// Now start loading the pipes into buffers
	go stdout.ReadFrom(stdoutPipe)

	err = cmd.Wait()

	if err != nil {
		t.Fatalf("%v STDOUT: %v, STDERR: %v", err, stdout.String(), stderr.String())
	}

	if match, _ := regexp.MatchReader(p.Password, &stdout); !match {
		t.Error("Could not find password in stdout")
	}
}

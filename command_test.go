package selenium

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// fakeExecCommand is a replacement for `exec.Command` that we can control
// using the TestHelperProcess function.
//
// For more information, see:
// * https://npf.io/2015/06/testing-exec-command/
// * https://golang.org/src/os/exec/exec_test.go
func fakeExecCommand(command string, args ...string) *exec.Cmd {
	// Use `go test` to run the `TestHelperProcess` test with our arguments.
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestHelperProcess(t *testing.T) {
	// If this function (which masquerades as a test) is run on its own, then
	// just return quietly.
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command\n")
		os.Exit(2)
	}

	cmd, args := args[0], args[1:]
	switch cmd {
	case "echo":
		fmt.Printf("%s\n", strings.Join(args, " "))
		os.Exit(0)
	case "Xvfb":
		// Print out the X11 screen of "1".
		screenNumber := "1"
		file := os.NewFile(uintptr(3), "pipe")
		_, err := file.Write([]byte(screenNumber + "\n"))
		if err != nil {
			panic(err)
		}
		time.Sleep(time.Second * 3)
		file.Close()
		os.Exit(0)
	case "xauth":
		os.Exit(0)
	}

	fmt.Fprintf(os.Stderr, "%s: command not found\n", cmd)
	os.Exit(127)
}

func TestFakeExecCommand(t *testing.T) {
	cmd := fakeExecCommand("echo", "hello", "world")
	outputBytes, err := cmd.Output()
	if err != nil {
		t.Fatalf("Could not get output: %s", err.Error())
	}
	outputString := string(outputBytes)
	if outputString != "hello world\n" {
		t.Fatalf("outputString = %s, want = %s", outputString, "hello world\n")
	}
}

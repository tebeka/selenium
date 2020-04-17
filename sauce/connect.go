package sauce

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"time"
)

// Connect manages an instance of a Sauce Connect Proxy to allow Sauce Labs to
// access HTTP endpoints from the local machine, though a tunnel.
type Connect struct {
	// Path is the path to the Sauce Connect Proxy binary.
	Path string
	// UserName and AccessKey are the credentials used to authenticate with Sauce Labs.
	UserName, AccessKey string
	// LogFile is the location of the log file that the proxy binary should
	// create.
	LogFile string
	// PIDFile is the location of the file that will contain the SauceConnect Proxy process ID. If not specified, one will be generated to avoid collisions between multiple processes.
	PIDFile string
	// SeleniumPort is the port number that the Proxy binary should listen on for
	// new Selenium WebDriver connections.
	SeleniumPort int
	// Verbose and ExtraVerbose control the verbosity level of logging from the
	// Proxy binary.
	Verbose, ExtraVerbose bool
	// Args are additional arguments to provide to the Proxy binary.
	//
	// See the following URL for details about available flags:
	// https://wiki.saucelabs.com/pages/viewpage.action?pageId=48365781
	Args []string

	// If true and the current operating system is Linux, send SIGTERM to the
	// proxy process when this parent process exits.
	QuitProcessUponExit bool

	cmd *exec.Cmd
}

// Start starts the Sauce Connect Proxy.
func (c *Connect) Start() error {
	c.cmd = exec.Command(c.Path, c.Args...)
	if c.UserName != "" {
		c.cmd.Args = append(c.cmd.Args, "--user", c.UserName)
	}
	if c.AccessKey != "" {
		c.cmd.Args = append(c.cmd.Args, "--api-key", c.AccessKey)
	}
	if c.SeleniumPort > 0 {
		c.cmd.Args = append(c.cmd.Args, "--se-port", strconv.Itoa(c.SeleniumPort))
	}
	if c.ExtraVerbose {
		c.cmd.Args = append(c.cmd.Args, "-vv")
	} else if c.Verbose {
		c.cmd.Args = append(c.cmd.Args, "-v")
	}
	if c.ExtraVerbose || c.Verbose {
		c.cmd.Stdout = os.Stdout
		c.cmd.Stderr = os.Stderr
	}
	if c.LogFile != "" {
		c.cmd.Args = append(c.cmd.Args, "--logfile", c.LogFile)
	}
	if c.QuitProcessUponExit && runtime.GOOS == "linux" {
		c.cmd.SysProcAttr = &syscall.SysProcAttr{
			// Deliver SIGTERM to process when we die.
			//
			// TODO(minusnine): Pdeathsig is only supported on Linux. Somehow, make
			// sure process cleanup happens as gracefully as possible.
			//
			// Pdeathsig: syscall.SIGTERM,
		}
	}

	dir, err := ioutil.TempDir("", "selenium-sauce-connect")
	if err != nil {
		return err
	}
	defer func() {
		os.RemoveAll(dir) // ignore error.
	}()

	var pidFilePath string
	if c.PIDFile != "" {
		pidFilePath = c.PIDFile
	} else {
		f, err := ioutil.TempFile("", "selenium-sauce-connect-pid.")
		if err != nil {
			return err
		}
		pidFilePath = f.Name()
		f.Close() // ignore the error.
	}

	// The path specified here will be touched by the proxy process when it is
	// ready to accept connections.
	readyPath := filepath.Join(dir, "ready")
	c.cmd.Args = append(c.cmd.Args, "--readyfile", readyPath)
	c.cmd.Args = append(c.cmd.Args, "--pidfile", pidFilePath)

	if err := c.cmd.Start(); err != nil {
		return err
	}

	// Wait for the Proxy to accept connections.
	var started bool
	for i := 0; i < 60; i++ {
		time.Sleep(time.Second)
		if _, err := os.Stat(readyPath); err == nil {
			started = true
			break
		}
	}
	if !started {
		c.Stop() // ignore error.
		return fmt.Errorf("proxy process did not become ready before the timeout")
	}
	return nil
}

// Addr returns the URL of the WebDriver endpoint to use for driving the
// browser.
func (c *Connect) Addr() string {
	return fmt.Sprintf("http://%s:%s@localhost:%d/wd/hub", c.UserName, c.AccessKey, c.SeleniumPort)
}

// Stop terminates the Proxy process.
func (c *Connect) Stop() error {
	return c.cmd.Process.Kill()
}

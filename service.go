package selenium

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// ServiceOption configures a Service instance.
type ServiceOption func(*Service) error

// Display specifies the value to which set the DISPLAY environment variable,
// as well as the path to the Xauthority file containing credentials needed to
// write to that X server.
func Display(d, xauthPath string) ServiceOption {
	return func(s *Service) error {
		if s.display != "" {
			return fmt.Errorf("service display already set: %v", s.display)
		}
		if s.xauthPath != "" {
			return fmt.Errorf("service xauth path already set: %v", s.xauthPath)
		}
		if !isDisplay(d) {
			return fmt.Errorf("supplied display %q must be of the format 'x' or 'x.y' where x and y are integers", d)
		}
		s.display = d
		s.xauthPath = xauthPath
		return nil
	}
}

// isDisplay validates that the given disp is in the format "x" or "x.y", where
// x and y are both integers.
func isDisplay(disp string) bool {
	ds := strings.Split(disp, ".")
	if len(ds) > 2 {
		return false
	}

	for _, d := range ds {
		if _, err := strconv.Atoi(d); err != nil {
			return false
		}
	}
	return true
}

// StartFrameBuffer causes an X virtual frame buffer to start before the
// WebDriver service. The frame buffer process will be terminated when the
// service itself is stopped.
func StartFrameBuffer() ServiceOption {
	return func(s *Service) error {
		if s.display != "" {
			return fmt.Errorf("service display already set: %v", s.display)
		}
		if s.xauthPath != "" {
			return fmt.Errorf("service xauth path already set: %v", s.xauthPath)
		}
		if s.xvfb != nil {
			return fmt.Errorf("service Xvfb instance already running")
		}
		fb, err := NewFrameBuffer()
		if err != nil {
			return fmt.Errorf("error starting frame buffer: %v", err)
		}
		s.xvfb = fb
		return Display(fb.Display, fb.AuthPath)(s)
	}
}

// Output specifies that the WebDriver service should log to the provided
// writer.
func Output(w io.Writer) ServiceOption {
	return func(s *Service) error {
		s.output = w
		return nil
	}
}

// GeckoDriver sets the path to the geckodriver binary for the Selenium Server.
// Unlike other drivers, Selenium Server does not support specifying the
// geckodriver path at runtime. This ServiceOption is only useful when calling
// NewSeleniumService.
func GeckoDriver(path string) ServiceOption {
	return func(s *Service) error {
		s.geckoDriverPath = path
		return nil
	}
}

// JavaPath specifies the path to the JRE for a Selenium service.
func JavaPath(path string) ServiceOption {
	return func(s *Service) error {
		s.javaPath = path
		return nil
	}
}

// Service controls a locally-running Selenium subprocess.
type Service struct {
	port            int
	addr            string
	cmd             *exec.Cmd
	shutdownURLPath string

	display, xauthPath string
	xvfb               *FrameBuffer

	geckoDriverPath, javaPath string

	output io.Writer
}

// NewSeleniumService starts a Selenium instance in the background.
func NewSeleniumService(jarPath string, port int, opts ...ServiceOption) (*Service, error) {
	cmd := exec.Command("java", "-jar", jarPath, "-port", strconv.Itoa(port))
	s, err := newService(cmd, port, opts...)
	if err != nil {
		return nil, err
	}
	if s.javaPath != "" {
		s.cmd.Path = s.javaPath
	}
	if s.geckoDriverPath != "" {
		s.cmd.Args = append([]string{"java", "-Dwebdriver.gecko.driver=" + s.geckoDriverPath}, cmd.Args[1:]...)
	}
	if err := s.start(port); err != nil {
		return nil, err
	}
	return s, nil
}

// NewChromeDriverService starts a ChromeDriver instance in the background.
func NewChromeDriverService(path string, port int, opts ...ServiceOption) (*Service, error) {
	cmd := exec.Command(path, "--port="+strconv.Itoa(port), "--url-base=wd/hub", "--verbose")
	s, err := newService(cmd, port, opts...)
	if err != nil {
		return nil, err
	}
	s.shutdownURLPath = "/wd/hub/shutdown"
	if err := s.start(port); err != nil {
		return nil, err
	}
	return s, nil
}

// NewGeckoDriverService starts a GeckoDriver instance in the background.
func NewGeckoDriverService(path string, port int, opts ...ServiceOption) (*Service, error) {
	cmd := exec.Command(path, "--port", strconv.Itoa(port))
	s, err := newService(cmd, port, opts...)
	if err != nil {
		return nil, err
	}
	if err := s.start(port); err != nil {
		return nil, err
	}
	return s, nil
}

func newService(cmd *exec.Cmd, port int, opts ...ServiceOption) (*Service, error) {
	s := &Service{
		port: port,
		addr: fmt.Sprintf("http://localhost:%d", port),
	}
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}
	cmd.Stderr = s.output
	cmd.Stdout = s.output
	cmd.Env = os.Environ()
	// TODO(minusnine): Pdeathsig is only supported on Linux. Somehow, make sure
	// process cleanup happens as gracefully as possible.
	if s.display != "" {
		cmd.Env = append(cmd.Env, "DISPLAY=:"+s.display)
	}
	if s.xauthPath != "" {
		cmd.Env = append(cmd.Env, "XAUTHORITY="+s.xauthPath)
	}
	s.cmd = cmd
	return s, nil
}

func (s *Service) start(port int) error {
	if err := s.cmd.Start(); err != nil {
		return err
	}

	for i := 0; i < 30; i++ {
		time.Sleep(time.Second)
		resp, err := http.Get(s.addr + "/status")
		if err == nil {
			resp.Body.Close()
			switch resp.StatusCode {
			// Selenium <3 returned Forbidden and BadRequest. ChromeDriver and
			// Selenium 3 return OK.
			case http.StatusForbidden, http.StatusBadRequest, http.StatusOK:
				return nil
			}
		}
	}
	return fmt.Errorf("server did not respond on port %d", port)
}

// Stop shuts down the WebDriver service, and the X virtual frame buffer
// if one was started.
func (s *Service) Stop() error {
	// Selenium 3 stopped supporting the shutdown URL by default.
	// https://github.com/SeleniumHQ/selenium/issues/2852
	if s.shutdownURLPath == "" {
		if err := s.cmd.Process.Kill(); err != nil {
			return err
		}
	} else {
		resp, err := http.Get(s.addr + s.shutdownURLPath)
		if err != nil {
			return err
		}
		resp.Body.Close()
	}
	if err := s.cmd.Wait(); err != nil && err.Error() != "signal: killed" {
		return err
	}
	if s.xvfb != nil {
		return s.xvfb.Stop()
	}
	return nil
}

// FrameBuffer controls an X virtual frame buffer running as a background
// process.
type FrameBuffer struct {
	// Display is the X11 display number that the Xvfb process is hosting
	// (without the preceding colon).
	Display string
	// AuthPath is the path to the X11 authorization file that permits X clients
	// to use the X server. This is typically provided to the client via the
	// XAUTHORITY environment variable.
	AuthPath string

	cmd *exec.Cmd
}

// NewFrameBuffer starts an X virtual frame buffer running in the background.
func NewFrameBuffer() (*FrameBuffer, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	defer r.Close()

	auth, err := ioutil.TempFile("", "selenium-xvfb")
	if err != nil {
		return nil, err
	}
	authPath := auth.Name()
	if err := auth.Close(); err != nil {
		return nil, err
	}

	// Xvfb will print the display on which it is listening to file descriptor 3,
	// for which we provide a pipe.
	xvfb := exec.Command("Xvfb", "-displayfd", "3", "-nolisten", "tcp")
	xvfb.ExtraFiles = []*os.File{w}

	// TODO(minusnine): plumb a way to set xvfb.Std{err,out} conditionally.
	// TODO(minusnine): Pdeathsig is only supported on Linux. Somehow, make sure
	// process cleanup happens as gracefully as possible.
	xvfb.Env = []string{"XAUTHORITY=" + authPath}
	if err := xvfb.Start(); err != nil {
		return nil, err
	}
	w.Close()

	type resp struct {
		display string
		err     error
	}
	ch := make(chan resp)
	go func() {
		bufr := bufio.NewReader(r)
		s, err := bufr.ReadString('\n')
		ch <- resp{s, err}
	}()

	var display string
	select {
	case resp := <-ch:
		if resp.err != nil {
			return nil, resp.err
		}
		display = strings.TrimSpace(resp.display)
		if _, err := strconv.Atoi(display); err != nil {
			return nil, errors.New("Xvfb did not print the display number")
		}
	case <-time.After(3 * time.Second):
		return nil, errors.New("timeout waiting for Xvfb")
	}

	xauth := exec.Command("xauth", "generate", ":"+display, ".", "trusted")
	xauth.Stderr = os.Stderr
	xauth.Stdout = os.Stdout
	xauth.Env = []string{"XAUTHORITY=" + authPath}

	if err := xauth.Run(); err != nil {
		return nil, err
	}

	return &FrameBuffer{display, authPath, xvfb}, nil
}

// Stop kills the background frame buffer process and removes the X
// authorization file.
func (f FrameBuffer) Stop() error {
	if err := f.cmd.Process.Kill(); err != nil {
		return err
	}
	os.Remove(f.AuthPath) // best effort removal; ignore error
	if err := f.cmd.Wait(); err != nil && err.Error() != "signal: killed" {
		return err
	}
	return nil
}

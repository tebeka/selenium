// Package firefox provides Firefox-specific types for WebDriver.
package firefox

import (
	"bytes"
	"encoding/base64"

	"github.com/tebeka/selenium/internal/zip"
)

// CapabilitiesKey is the name of the Firefox-specific key in the WebDriver
// capabilities object.
const CapabilitiesKey = "moz:firefoxOptions"

// Capabilities provides Firefox-specific options to WebDriver.
type Capabilities struct {
	// Binary is the absolute path of the Firefox binary, e.g. /usr/bin/firefox
	// or /Applications/Firefox.app/Contents/MacOS/firefox, to select which
	// custom browser binary to use. If left undefined, geckodriver will attempt
	// to deduce the default location of Firefox on the current system.
	Binary string `json:"binary,omitempty"`
	// Args are the command line arguments to pass to the Firefox binary. These
	// must include the leading -- where required e.g. ["--devtools"].
	Args []string `json:"args,omitempty"`
	// Profile is the Base64-encoded zip file of a profile directory to use as
	// the profile for the Firefox instance. This may be used to e.g.
	// install extensions or custom certificates. Use the SetProfile method
	// to load an existing profile from a file system.
	Profile string `json:"profile,omitempty"`
	// Log specifies the logging options for Gecko.
	Log *Log `json:"log,omitempty"`
	// Map of preference name to preference value, which can be a string, a
	// boolean or an integer.
	Prefs map[string]interface{} `json:"prefs,omitempty"`
}

// SetProfile sets the Profile datum with a Base64-encoded zip file of a
// profile directory that is specified by basePath. This directory should
// directly contain the profile's files, e.g. "user.js".
//
// Note that a zip file will be created in memory and then the zip file
// will be base64-encoded. This will require memory at least 2x the size
// of the data.
func (c *Capabilities) SetProfile(basePath string) error {
	buf, err := zip.New(basePath)
	if err != nil {
		return err
	}
	encoded := new(bytes.Buffer)
	encoded.Grow(buf.Len())
	encoder := base64.NewEncoder(base64.StdEncoding, encoded)
	if _, err := buf.WriteTo(encoder); err != nil {
		return err
	}
	encoder.Close()

	c.Profile = encoded.String()

	return nil
}

// LogLevel is an enum that defines logging levels for Firefox.
type LogLevel string

// Levels of logging that can be specified in the Log structure.
const (
	Trace  LogLevel = "trace"
	Debug           = "debug"
	Config          = "config"
	Info            = "info"
	Warn            = "warn"
	Error           = "error"
	Fatal           = "fatal"
)

// Log specifies how Firefox should log debug data.
type Log struct {
	// Level is the verbosity level of logs that Firefox should output.
	Level LogLevel `json:"level"`
}

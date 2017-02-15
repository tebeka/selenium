// Package firefox provides Firefox-specific types for WebDriver.
package firefox

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
	// Profile is the Base64-encoded zip of a profile directory to use as the
	// profile for the Firefox instance. This may be used to e.g. install
	// extensions or custom certificates.
	Profile string `json:"profile,omitempty"`
	// Log specifies the logging options for Gecko.
	Log *Log `json:"log,omitempty"`
	// Map of preference name to preference value, which can be a string, a
	// boolean or an integer.
	Prefs map[string]interface{} `json:"prefs,omitempty"`
}

// LogLevel is an enum that defines logging levels for Firefox.
type LogLevel string

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

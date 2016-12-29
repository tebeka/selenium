// Package chrome provides Chrome-specific options for WebDriver.
package chrome

const CapabilitiesKey = "chromeOptions"

// Capabilities defines the Chrome-specific desired capabilities when using
// ChromeDriver. An instance of this struct can be stored in the Capabilities
// map with a key of `chromeOptions`.  See
// https://sites.google.com/a/chromium.org/chromedriver/capabilities
type Capabilities struct {
	// Path is the file path to the Chrome binary to use.
	Path string `json:"binary,omitempty"`
	// Args are the command-line arguments to pass to the Chrome binary, in
	// addition to the ChromeDriver-supplied ones.
	Args []string `json:"args,omitempty"`
	// ExcludeSwitches are the command line flags that should be removed from
	// the ChromeDriver-supplied default flags. The strings included here should
	// not include a preceding '--'.
	ExcludeSwitches []string `json:"excludeSwitches,omitempty"`
	// Extensions are the list of extentions to install at startup. The
	// elements of this list should be the base-64, padded contents of a Chrome
	// extension file (.crx). Use the AddExtension method to add a local file.
	Extensions []string `json:"extensions,omitempty"`
	// LocalState are key/value pairs that are applied to the Local State file
	// in the user data folder.
	LocalState map[string]interface{} `json:"localState,omitempty"`
	// Prefs are the key/value pairs that are applied to the preferences of the
	// user profile in use.
	Prefs map[string]interface{} `json:"prefs,omitempty"`
	// Detatch, if true, will cause the browser to not be killed when
	// ChromeDriver quits if the session was not terminated.
	Detach *bool `json:"detach,omitempty"`
	// DebuggerAddr is the TCP/IP address of a Chrome debugger server to connect
	// to.
	DebuggerAddr string `json:"debuggerAddress,omitempty"`
	// MinidumpPath specifies the directory in which to store Chrome minidumps.
	// (This is only available on Linux).
	MinidumpPath string `json:"minidumpPath,omitempty"`
	// MobileEmulation provides options for mobile emulation.
	MobileEmulation *MobileEmulation `json:"mobileEmulation,omitempty"`
	// PerfLoggingPrefs specifies options for performance logging.
	PerfLoggingPrefs *PerfLoggingPreferences `json:"perfLoggingPrefs,omitempty"`
	// WindowTypes is a list of window types that will appear in the list of
	// window handles. For access to <webview> elements, include "webview" in
	// this list.
	WindowTypes []string `json:"windowTypes,omitempty"`
}

// MobileEmulation provides options for mobile emulation. Only
// DeviceName or both of DeviceMetrics and UserAgent may be set at once.
type MobileEmulation struct {
	// DeviceName is the name of the device to emulate, e.g. "Google Nexus 5".
	// It should not be set if DeviceMetrics and UserAgent are set.
	DeviceName string `json:"deviceName,omitempty"`
	// DeviceMetrics provides specifications of an device to emulate. It should
	// not be set if DeviceName is set.
	DeviceMetrics DeviceMetrics `json:"deviceMetrics,omitempty"`
	// UserAgent specifies the user agent string to send to the remote web
	// server.
	UserAgent string `json:"userAgent,omitempty"`
}

// DeviceMetrics specifies device attributes for emulation.
type DeviceMetrics struct {
	// Width is the width of the screen.
	Width uint `json:"width"`
	// Height is the height of the screen.
	Height uint `json:"height"`
	// PixelRatio is the pixel ratio of the screen.
	PixelRatio float64 `json:"pixelRatio"`
	// Touch indicates whether to emulate touch events. The default is true, if
	// unset.
	Touch *bool `json:"touch,omitempty"`
}

// PerfLoggingPreferences specifies configuration options for performance
// logging.
type PerfLoggingPreferences struct {
	// EnableNetwork specifies whether of not to collect events from the Network
	// domain. The default is true.
	EnableNetwork *bool `json:"enableNetwork,omitempty"`
	// EnablePage specifies whether or not to collect events from the Page
	// domain. The default is true.
	EnablePage *bool `json:"enablePage,omitempty"`
	// EnableTimeline specifies whether or not to collect events from the
	// Timeline domain. When tracing is enabled, Timeline domain is implicitly
	// disabled, unless enableTimeline is explicitly set to true.
	EnableTimeline *bool `json:"enableTimeline,omitempty"`
	// TracingCategories is a comma-separated string of Chrome tracing categories
	// for which trace events should be collected. An unspecified or empty string
	// disables tracing.
	TracingCategories string `json:"tracingCategories,omitempty"`
	// BufferUsageReportingIntervalMillis is the requested number of milliseconds
	// between DevTools trace buffer usage events. For example, if 1000, then
	// once per second, DevTools will report how full the trace buffer is. If a
	// report indicates the buffer usage is 100%, a warning will be issued.
	BufferUsageReportingIntervalMillis uint `json:"bufferUsageReportingInterval,omitempty"`
}

// TODO(minusnine): Add a method to add an extension given a path to a .crx file.

// Package chrome provides Chrome-specific options for WebDriver.
package chrome

import (
	"bufio"
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"io"
	"os"

	"github.com/tebeka/selenium/internal/zip"
)

// CapabilitiesKey is the key in the top-level Capabilities map under which
// ChromeDriver expects the Chrome-specific options to be set.
const CapabilitiesKey = "goog:chromeOptions"

// DeprecatedCapabilitiesKey is the legacy version of CapabilitiesKey.
const DeprecatedCapabilitiesKey = "chromeOptions"

// Capabilities defines the Chrome-specific desired capabilities when using
// ChromeDriver. An instance of this struct can be stored in the Capabilities
// map with a key of CapabilitiesKey ("goog:chromeOptions").  See
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
	// Extensions are the list of extensions to install at startup. The
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
	// Android Chrome WebDriver path "com.android.chrome"
	AndroidPackage string `json:"androidPackage,omitempty"`
	// Use W3C mode, if true.
	W3C bool `json:"w3c"`
}

// TODO(minusnine): https://bugs.chromium.org/p/chromedriver/issues/detail?id=1625
// mentions "experimental options". Implement that.

// MobileEmulation provides options for mobile emulation. Only
// DeviceName or both of DeviceMetrics and UserAgent may be set at once.
type MobileEmulation struct {
	// DeviceName is the name of the device to emulate, e.g. "Google Nexus 5".
	// It should not be set if DeviceMetrics and UserAgent are set.
	DeviceName string `json:"deviceName,omitempty"`
	// DeviceMetrics provides specifications of an device to emulate. It should
	// not be set if DeviceName is set.
	DeviceMetrics *DeviceMetrics `json:"deviceMetrics,omitempty"`
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
	// TraceCategories is a comma-separated string of Chrome tracing categories
	// for which trace events should be collected. An unspecified or empty string
	// disables tracing.
	TraceCategories string `json:"traceCategories,omitempty"`
	// BufferUsageReportingIntervalMillis is the requested number of milliseconds
	// between DevTools trace buffer usage events. For example, if 1000, then
	// once per second, DevTools will report how full the trace buffer is. If a
	// report indicates the buffer usage is 100%, a warning will be issued.
	BufferUsageReportingIntervalMillis uint `json:"bufferUsageReportingInterval,omitempty"`
}

// AddExtension adds an extension for the browser to load at startup. The path
// parameter should be a path to an extension file (which typically has a
// `.crx` file extension. Note that the contents of the file will be loaded
// into memory, as required by the protocol.
func (c *Capabilities) AddExtension(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return c.addExtension(f)
}

// addExtension reads a Chrome extension's data from r, base64-encodes it, and
// attaches it to the Capabilities instance.
func (c *Capabilities) addExtension(r io.Reader) error {
	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	if _, err := io.Copy(encoder, bufio.NewReader(r)); err != nil {
		return err
	}
	encoder.Close()
	c.Extensions = append(c.Extensions, buf.String())
	return nil
}

// AddUnpackedExtension creates a packaged Chrome extension with the files
// below the provided directory path and causes the browser to load that
// extension at startup.
func (c *Capabilities) AddUnpackedExtension(basePath string) error {
	buf, _, err := NewExtension(basePath)
	if err != nil {
		return err
	}
	return c.addExtension(bytes.NewBuffer(buf))
}

// NewExtension creates the payload of a Chrome extension file which is signed
// using the returned private key.
func NewExtension(basePath string) ([]byte, *rsa.PrivateKey, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	data, err := NewExtensionWithKey(basePath, key)
	if err != nil {
		return nil, nil, err
	}
	return data, key, nil
}

// NewExtensionWithKey creates the payload of a Chrome extension file which is
// signed by the provided private key.
func NewExtensionWithKey(basePath string, key *rsa.PrivateKey) ([]byte, error) {
	zip, err := zip.New(basePath)
	if err != nil {
		return nil, err
	}

	h := sha1.New()
	if _, err := io.Copy(h, bytes.NewReader(zip.Bytes())); err != nil {
		return nil, err
	}
	hashed := h.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA1, hashed[:])
	if err != nil {
		return nil, err
	}

	pubKey, err := x509.MarshalPKIXPublicKey(key.Public())
	if err != nil {
		return nil, err
	}

	// This format is documented at https://developer.chrome.com/extensions/crx .
	buf := new(bytes.Buffer)
	if _, err := buf.Write([]byte("Cr24")); err != nil { // Magic number.
		return nil, err
	}

	// Version.
	if err := binary.Write(buf, binary.LittleEndian, uint32(2)); err != nil {
		return nil, err
	}

	// Public key length.
	if err := binary.Write(buf, binary.LittleEndian, uint32(len(pubKey))); err != nil {
		return nil, err
	}
	// Signature length.
	if err := binary.Write(buf, binary.LittleEndian, uint32(len(signature))); err != nil {
		return nil, err
	}

	// Public key payload.
	if err := binary.Write(buf, binary.LittleEndian, pubKey); err != nil {
		return nil, err
	}

	// Signature payload.
	if err := binary.Write(buf, binary.LittleEndian, signature); err != nil {
		return nil, err
	}

	// Zipped extension directory payload.
	if err := binary.Write(buf, binary.LittleEndian, zip.Bytes()); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

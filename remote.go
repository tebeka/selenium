// Remote Selenium client implementation.
// See https://www.w3.org/TR/webdriver for the protocol.

package selenium

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/tebeka/selenium/firefox"
	"github.com/tebeka/selenium/log"
)

// Errors returned by Selenium server.
var remoteErrors = map[int]string{
	6:  "invalid session ID",
	7:  "no such element",
	8:  "no such frame",
	9:  "unknown command",
	10: "stale element reference",
	11: "element not visible",
	12: "invalid element state",
	13: "unknown error",
	15: "element is not selectable",
	17: "javascript error",
	19: "xpath lookup error",
	21: "timeout",
	23: "no such window",
	24: "invalid cookie domain",
	25: "unable to set cookie",
	26: "unexpected alert open",
	27: "no alert open",
	28: "script timeout",
	29: "invalid element coordinates",
	32: "invalid selector",
}

type remoteWD struct {
	id, urlPrefix string
	capabilities  Capabilities

	w3cCompatible  bool
	browser        string
	browserVersion semver.Version
}

// HTTPClient is the default client to use to communicate with the WebDriver
// server.
var HTTPClient = http.DefaultClient

// jsonContentType is JSON content type.
const jsonContentType = "application/json"

func newRequest(method string, url string, data []byte) (*http.Request, error) {
	request, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	request.Header.Add("Accept", jsonContentType)

	return request, nil
}

func (wd *remoteWD) requestURL(template string, args ...interface{}) string {
	return wd.urlPrefix + fmt.Sprintf(template, args...)
}

// TODO(minusnine): provide a "sessionURL" function that prepends the
// /session/<id> URL prefix and replace most requestURL (and voidCommand) calls
// with it.

type serverReply struct {
	SessionID *string // SessionID can be nil.
	Value     json.RawMessage

	// The following fields were used prior to Selenium 3.0 for error state and
	// in ChromeDriver for additional information.
	Status int
	State  string

	Error
}

// Error contains information about a failure of a command. See the table of
// these strings at https://www.w3.org/TR/webdriver/#handling-errors .
//
// This error type is only returned by servers that implement the W3C
// specification.
type Error struct {
	// Err contains a general error string provided by the server.
	Err string `json:"error"`
	// Message is a detailed, human-readable message specific to the failure.
	Message string `json:"message"`
	// Stacktrace may contain the server-side stacktrace where the error occurred.
	Stacktrace string `json:"stacktrace"`
	// HTTPCode is the HTTP status code returned by the server.
	HTTPCode int
	// LegacyCode is the "Response Status Code" defined in the legacy Selenium
	// WebDriver JSON wire protocol. This code is only produced by older
	// Selenium WebDriver versions, Chromedriver, and InternetExplorerDriver.
	LegacyCode int
}

// TODO(minusnine): Make Stacktrace more descriptive. Selenium emits a list of
// objects that enumerate various fields. This is not standard, though.

// Error implements the error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Err, e.Message)
}

// execute performs an HTTP request and inspects the returned data for an error
// encoded by the remote end in a JSON structure. If no error is present, the
// entire, raw request payload is returned.
func (wd *remoteWD) execute(method, url string, data []byte) (json.RawMessage, error) {
	return executeCommand(method, url, data)
}

func executeCommand(method, url string, data []byte) (json.RawMessage, error) {
	debugLog("-> %s %s\n%s", method, filteredURL(url), data)
	request, err := newRequest(method, url, data)
	if err != nil {
		return nil, err
	}

	response, err := HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}

	buf, err := ioutil.ReadAll(response.Body)
	if debugFlag {
		if err == nil {
			// Pretty print the JSON response
			var prettyBuf bytes.Buffer
			if err = json.Indent(&prettyBuf, buf, "", "    "); err == nil && prettyBuf.Len() > 0 {
				buf = prettyBuf.Bytes()
			}
		}
		debugLog("<- %s [%s]\n%s", response.Status, response.Header["Content-Type"], buf)
	}
	if err != nil {
		return nil, errors.New(response.Status)
	}

	fullCType := response.Header.Get("Content-Type")
	cType, _, err := mime.ParseMediaType(fullCType)
	if err != nil {
		return nil, fmt.Errorf("got content type header %q, expected %q", fullCType, jsonContentType)
	}
	if cType != jsonContentType {
		return nil, fmt.Errorf("got content type %q, expected %q", cType, jsonContentType)
	}

	reply := new(serverReply)
	if err := json.Unmarshal(buf, reply); err != nil {
		if response.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("bad server reply status: %s", response.Status)
		}
		return nil, err
	}
	if reply.Err != "" {
		return nil, &reply.Error
	}

	// Handle the W3C-compliant error format. In the W3C spec, the error is
	// embedded in the 'value' field.
	if len(reply.Value) > 0 {
		respErr := new(Error)
		if err := json.Unmarshal(reply.Value, respErr); err == nil && respErr.Err != "" {
			respErr.HTTPCode = response.StatusCode
			return nil, respErr
		}
	}

	// Handle the legacy error format.
	const success = 0
	if reply.Status != success {
		shortMsg, ok := remoteErrors[reply.Status]
		if !ok {
			shortMsg = fmt.Sprintf("unknown error - %d", reply.Status)
		}

		longMsg := new(struct {
			Message string
		})
		if err := json.Unmarshal(reply.Value, longMsg); err != nil {
			return nil, errors.New(shortMsg)
		}
		return nil, &Error{
			Err:        shortMsg,
			Message:    longMsg.Message,
			HTTPCode:   response.StatusCode,
			LegacyCode: reply.Status,
		}
	}

	return buf, nil
}

// DefaultURLPrefix is the default HTTP endpoint that offers the WebDriver API.
const DefaultURLPrefix = "http://127.0.0.1:4444/wd/hub"

// NewRemote creates new remote client, this will also start a new session.
// capabilities provides the desired capabilities. urlPrefix is the URL to the
// Selenium server, must be prefixed with protocol (http, https, ...).
//
// Providing an empty string for urlPrefix causes the DefaultURLPrefix to be
// used.
func NewRemote(capabilities Capabilities, urlPrefix string) (WebDriver, error) {
	if urlPrefix == "" {
		urlPrefix = DefaultURLPrefix
	}

	wd := &remoteWD{
		urlPrefix:    urlPrefix,
		capabilities: capabilities,
	}
	if b := capabilities["browserName"]; b != nil {
		wd.browser = b.(string)
	}
	if _, err := wd.NewSession(); err != nil {
		return nil, err
	}
	return wd, nil
}

// DeleteSession deletes an existing session at the WebDriver instance
// specified by the urlPrefix and the session ID.
func DeleteSession(urlPrefix, id string) error {
	u, err := url.Parse(urlPrefix)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, "session", id)
	return voidCommand("DELETE", u.String(), nil)
}

func (wd *remoteWD) stringCommand(urlTemplate string) (string, error) {
	url := wd.requestURL(urlTemplate, wd.id)
	response, err := wd.execute("GET", url, nil)
	if err != nil {
		return "", err
	}

	reply := new(struct{ Value *string })
	if err := json.Unmarshal(response, reply); err != nil {
		return "", err
	}

	if reply.Value == nil {
		return "", fmt.Errorf("nil return value")
	}

	return *reply.Value, nil
}

func voidCommand(method, url string, params interface{}) error {
	if params == nil {
		params = make(map[string]interface{})
	}
	data, err := json.Marshal(params)
	if err != nil {
		return err
	}
	_, err = executeCommand(method, url, data)
	return err
}

func (wd *remoteWD) voidCommand(urlTemplate string, params interface{}) error {
	return voidCommand("POST", wd.requestURL(urlTemplate, wd.id), params)
}

func (wd remoteWD) stringsCommand(urlTemplate string) ([]string, error) {
	url := wd.requestURL(urlTemplate, wd.id)
	response, err := wd.execute("GET", url, nil)
	if err != nil {
		return nil, err
	}

	reply := new(struct{ Value []string })
	if err := json.Unmarshal(response, reply); err != nil {
		return nil, err
	}

	return reply.Value, nil
}

func (wd *remoteWD) boolCommand(urlTemplate string) (bool, error) {
	url := wd.requestURL(urlTemplate, wd.id)
	response, err := wd.execute("GET", url, nil)
	if err != nil {
		return false, err
	}

	reply := new(struct{ Value bool })
	if err := json.Unmarshal(response, reply); err != nil {
		return false, err
	}

	return reply.Value, nil
}

func (wd *remoteWD) Status() (*Status, error) {
	url := wd.requestURL("/status")
	reply, err := wd.execute("GET", url, nil)
	if err != nil {
		return nil, err
	}

	status := new(struct{ Value Status })
	if err := json.Unmarshal(reply, status); err != nil {
		return nil, err
	}

	return &status.Value, nil
}

// parseVersion sanitizes the browser version enough for semver.ParseTolerant
// to parse it.
func parseVersion(v string) (semver.Version, error) {
	parts := strings.Split(v, ".")
	var err error
	for i := len(parts); i > 0; i-- {
		var ver semver.Version
		ver, err = semver.ParseTolerant(strings.Join(parts[:i], "."))
		if err == nil {
			return ver, nil
		}
	}
	return semver.Version{}, err
}

// The list of valid, top-level capability names, according to the W3C
// specification.
//
// This must be kept in sync with the specification:
// https://www.w3.org/TR/webdriver/#capabilities
var w3cCapabilityNames = []string{
	"acceptInsecureCerts",
	"browserName",
	"browserVersion",
	"platformName",
	"pageLoadStrategy",
	"proxy",
	"setWindowRect",
	"timeouts",
	"unhandledPromptBehavior",
}

var chromeCapabilityNames = []string{
	// This is not a standardized top-level capability name, but Chromedriver
	// expects this capability here.
	// https://cs.chromium.org/chromium/src/chrome/test/chromedriver/capabilities.cc?rcl=0754b5d0aad903439a628618f0e41845f1988f0c&l=759
	"loggingPrefs",
}

// Create a W3C-compatible capabilities instance.
func newW3CCapabilities(caps Capabilities) Capabilities {
	isValidW3CCapability := map[string]bool{}
	for _, name := range w3cCapabilityNames {
		isValidW3CCapability[name] = true
	}
	if b, ok := caps["browserName"]; ok && b == "chrome" {
		for _, name := range chromeCapabilityNames {
			isValidW3CCapability[name] = true
		}
	}

	alwaysMatch := make(Capabilities)
	for name, value := range caps {
		if isValidW3CCapability[name] || strings.Contains(name, ":") {
			alwaysMatch[name] = value
		}
	}

	// Move the Firefox profile setting from the old location to the new
	// location.
	if prof, ok := caps["firefox_profile"]; ok {
		if c, ok := alwaysMatch[firefox.CapabilitiesKey]; ok {
			firefoxCaps := c.(firefox.Capabilities)
			if firefoxCaps.Profile == "" {
				firefoxCaps.Profile = prof.(string)
			}
		} else {
			alwaysMatch[firefox.CapabilitiesKey] = firefox.Capabilities{
				Profile: prof.(string),
			}
		}
	}

	return Capabilities{
		"alwaysMatch": alwaysMatch,
	}
}

func (wd *remoteWD) NewSession() (string, error) {
	// Detect whether the remote end complies with the W3C specification:
	// non-compliant implementations use the top-level 'desiredCapabilities' JSON
	// key, whereas the specification mandates the 'capabilities' key.
	//
	// However, Selenium 3 currently does not implement this part of the specification.
	// https://github.com/SeleniumHQ/selenium/issues/2827
	//
	// TODO(minusnine): audit which ones of these are still relevant. The W3C
	// standard switched to the "alwaysMatch" version in February 2017.
	attempts := []struct {
		params map[string]interface{}
	}{
		{map[string]interface{}{
			"capabilities":        newW3CCapabilities(wd.capabilities),
			"desiredCapabilities": wd.capabilities,
		}},
		{map[string]interface{}{
			"capabilities": map[string]interface{}{
				"desiredCapabilities": wd.capabilities,
			},
		}},
		{map[string]interface{}{
			"desiredCapabilities": wd.capabilities,
		}}}

	for i, s := range attempts {
		data, err := json.Marshal(s.params)
		if err != nil {
			return "", err
		}

		response, err := wd.execute("POST", wd.requestURL("/session"), data)
		if err != nil {
			return "", err
		}

		reply := new(serverReply)
		if err := json.Unmarshal(response, reply); err != nil {
			if i < len(attempts) {
				continue
			}
			return "", err
		}
		if reply.Status != 0 && i < len(attempts) {
			continue
		}
		if reply.SessionID != nil {
			wd.id = *reply.SessionID
		}

		if len(reply.Value) > 0 {
			type returnedCapabilities struct {
				// firefox via geckodriver: 55.0a1
				BrowserVersion string
				// chrome via chromedriver: 61.0.3116.0
				// firefox via selenium 2: 45.9.0
				// htmlunit: 9.4.3.v20170317
				Version          string
				PageLoadStrategy string
				Proxy            Proxy
				Timeouts         struct {
					Implicit       float32
					PageLoadLegacy float32 `json:"page load"`
					PageLoad       float32
					Script         float32
				}
			}

			value := struct {
				SessionID string

				// The W3C specification moved most of the returned data into the
				// "capabilities" field.
				Capabilities *returnedCapabilities

				// Legacy implementations returned most data directly in the "values"
				// key.
				returnedCapabilities
			}{}

			if err := json.Unmarshal(reply.Value, &value); err != nil {
				return "", fmt.Errorf("error unmarshalling value: %v", err)
			}
			if value.SessionID != "" && wd.id == "" {
				wd.id = value.SessionID
			}
			var caps returnedCapabilities
			if value.Capabilities != nil {
				caps = *value.Capabilities
				wd.w3cCompatible = true
			} else {
				caps = value.returnedCapabilities
			}

			for _, s := range []string{caps.Version, caps.BrowserVersion} {
				if s == "" {
					continue
				}
				v, err := parseVersion(s)
				if err != nil {
					debugLog("error parsing version: %v\n", err)
					continue
				}
				wd.browserVersion = v
			}
		}

		return wd.id, nil
	}
	panic("unreachable")
}

// SessionId returns the current session ID
//
// Deprecated: This identifier is not Go-style correct. Use SessionID instead.
func (wd *remoteWD) SessionId() string {
	return wd.SessionID()
}

// SessionID returns the current session ID
func (wd *remoteWD) SessionID() string {
	return wd.id
}

func (wd *remoteWD) SwitchSession(sessionID string) error {
	wd.id = sessionID
	return nil
}

func (wd *remoteWD) Capabilities() (Capabilities, error) {
	url := wd.requestURL("/session/%s", wd.id)
	response, err := wd.execute("GET", url, nil)
	if err != nil {
		return nil, err
	}

	c := new(struct{ Value Capabilities })
	if err := json.Unmarshal(response, c); err != nil {
		return nil, err
	}

	return c.Value, nil
}

func (wd *remoteWD) SetAsyncScriptTimeout(timeout time.Duration) error {
	if !wd.w3cCompatible {
		return wd.voidCommand("/session/%s/timeouts/async_script", map[string]uint{
			"ms": uint(timeout / time.Millisecond),
		})
	}
	return wd.voidCommand("/session/%s/timeouts", map[string]uint{
		"script": uint(timeout / time.Millisecond),
	})
}

func (wd *remoteWD) SetImplicitWaitTimeout(timeout time.Duration) error {
	if !wd.w3cCompatible {
		return wd.voidCommand("/session/%s/timeouts/implicit_wait", map[string]uint{
			"ms": uint(timeout / time.Millisecond),
		})
	}
	return wd.voidCommand("/session/%s/timeouts", map[string]uint{
		"implicit": uint(timeout / time.Millisecond),
	})
}

func (wd *remoteWD) SetPageLoadTimeout(timeout time.Duration) error {
	if !wd.w3cCompatible {
		return wd.voidCommand("/session/%s/timeouts", map[string]interface{}{
			"ms":   uint(timeout / time.Millisecond),
			"type": "page load",
		})
	}
	return wd.voidCommand("/session/%s/timeouts", map[string]uint{
		"pageLoad": uint(timeout / time.Millisecond),
	})
}

func (wd *remoteWD) Quit() error {
	if wd.id == "" {
		return nil
	}
	_, err := wd.execute("DELETE", wd.requestURL("/session/%s", wd.id), nil)
	if err == nil {
		wd.id = ""
	}
	return err
}

func (wd *remoteWD) CurrentWindowHandle() (string, error) {
	if !wd.w3cCompatible {
		return wd.stringCommand("/session/%s/window_handle")
	}
	return wd.stringCommand("/session/%s/window")
}

func (wd *remoteWD) WindowHandles() ([]string, error) {
	if !wd.w3cCompatible {
		return wd.stringsCommand("/session/%s/window_handles")
	}
	return wd.stringsCommand("/session/%s/window/handles")
}

func (wd *remoteWD) CurrentURL() (string, error) {
	url := wd.requestURL("/session/%s/url", wd.id)
	response, err := wd.execute("GET", url, nil)
	if err != nil {
		return "", err
	}
	reply := new(struct{ Value *string })
	if err := json.Unmarshal(response, reply); err != nil {
		return "", err
	}

	return *reply.Value, nil
}

func (wd *remoteWD) Get(url string) error {
	requestURL := wd.requestURL("/session/%s/url", wd.id)
	params := map[string]string{
		"url": url,
	}
	data, err := json.Marshal(params)
	if err != nil {
		return err
	}
	_, err = wd.execute("POST", requestURL, data)
	return err
}

func (wd *remoteWD) Forward() error {
	return wd.voidCommand("/session/%s/forward", nil)
}

func (wd *remoteWD) Back() error {
	return wd.voidCommand("/session/%s/back", nil)
}

func (wd *remoteWD) Refresh() error {
	return wd.voidCommand("/session/%s/refresh", nil)
}

func (wd *remoteWD) Title() (string, error) {
	return wd.stringCommand("/session/%s/title")
}

func (wd *remoteWD) PageSource() (string, error) {
	return wd.stringCommand("/session/%s/source")
}

func (wd *remoteWD) find(by, value, suffix, url string) ([]byte, error) {
	// The W3C specification removed the specific ID and Name locator strategies,
	// instead only providing a CSS-based strategy. Emulate the old behavior to
	// maintain API compatibility.
	if wd.w3cCompatible {
		switch by {
		case ByID:
			by = ByCSSSelector
			value = "#" + value
		case ByName:
			by = ByCSSSelector
			value = fmt.Sprintf("input[name=%q]", value)
		}
	}

	params := map[string]string{
		"using": by,
		"value": value,
	}
	data, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	if len(url) == 0 {
		url = "/session/%s/element"
	}

	return wd.execute("POST", wd.requestURL(url+suffix, wd.id), data)
}

func (wd *remoteWD) DecodeElement(data []byte) (WebElement, error) {
	reply := new(struct{ Value map[string]string })
	if err := json.Unmarshal(data, &reply); err != nil {
		return nil, err
	}

	id := elementIDFromValue(reply.Value)
	if id == "" {
		return nil, fmt.Errorf("invalid element returned: %+v", reply)
	}
	return &remoteWE{
		parent: wd,
		id:     id,
	}, nil
}

const (
	// legacyWebElementIdentifier is the string constant used in the old
	// WebDriver JSON protocol that is the key for the map that contains an
	// unique element identifier.
	legacyWebElementIdentifier = "ELEMENT"

	// webElementIdentifier is the string constant defined by the W3C
	// specification that is the key for the map that contains a unique element identifier.
	webElementIdentifier = "element-6066-11e4-a52e-4f735466cecf"
)

func elementIDFromValue(v map[string]string) string {
	for _, key := range []string{webElementIdentifier, legacyWebElementIdentifier} {
		v, ok := v[key]
		if !ok || v == "" {
			continue
		}
		return v
	}
	return ""
}

func (wd *remoteWD) DecodeElements(data []byte) ([]WebElement, error) {
	reply := new(struct{ Value []map[string]string })
	if err := json.Unmarshal(data, reply); err != nil {
		return nil, err
	}

	elems := make([]WebElement, len(reply.Value))
	for i, elem := range reply.Value {
		id := elementIDFromValue(elem)
		if id == "" {
			return nil, fmt.Errorf("invalid element returned: %+v", reply)
		}
		elems[i] = &remoteWE{
			parent: wd,
			id:     id,
		}
	}

	return elems, nil
}

func (wd *remoteWD) FindElement(by, value string) (WebElement, error) {
	response, err := wd.find(by, value, "", "")
	if err != nil {
		return nil, err
	}
	return wd.DecodeElement(response)
}

func (wd *remoteWD) FindElements(by, value string) ([]WebElement, error) {
	response, err := wd.find(by, value, "s", "")
	if err != nil {
		return nil, err
	}

	return wd.DecodeElements(response)
}

func (wd *remoteWD) Close() error {
	url := wd.requestURL("/session/%s/window", wd.id)
	_, err := wd.execute("DELETE", url, nil)
	return err
}

func (wd *remoteWD) SwitchWindow(name string) error {
	params := make(map[string]string)
	if !wd.w3cCompatible {
		params["name"] = name
	} else {
		params["handle"] = name
	}
	return wd.voidCommand("/session/%s/window", params)
}

func (wd *remoteWD) CloseWindow(name string) error {
	return wd.modifyWindow(name, "DELETE", "", nil)
}

func (wd *remoteWD) MaximizeWindow(name string) error {
	if !wd.w3cCompatible {
		if name != "" {
			var err error
			name, err = wd.CurrentWindowHandle()
			if err != nil {
				return err
			}
		}
		url := wd.requestURL("/session/%s/window/%s/maximize", wd.id, name)
		_, err := wd.execute("POST", url, nil)
		return err
	}
	return wd.modifyWindow(name, "POST", "maximize", map[string]string{})
}

func (wd *remoteWD) MinimizeWindow(name string) error {
	return wd.modifyWindow(name, "POST", "minimize", map[string]string{})
}

func (wd *remoteWD) modifyWindow(name, verb, command string, params interface{}) error {
	// The original protocol allowed for maximizing any named window. The W3C
	// specification only allows the current window be be modified. Emulate the
	// previous behavior by switching to the target window, maximizing the
	// current window, and switching back to the original window.
	var startWindow string
	if name != "" && wd.w3cCompatible {
		var err error
		startWindow, err = wd.CurrentWindowHandle()
		if err != nil {
			return err
		}
		if name != startWindow {
			if err := wd.SwitchWindow(name); err != nil {
				return err
			}
		}
	}

	url := wd.requestURL("/session/%s/window", wd.id)
	if command != "" {
		if wd.w3cCompatible {
			url = wd.requestURL("/session/%s/window/%s", wd.id, command)
		} else {
			url = wd.requestURL("/session/%s/window/%s/%s", wd.id, name, command)
		}
	}

	var data []byte
	if params != nil {
		var err error
		if data, err = json.Marshal(params); err != nil {
			return err
		}
	}

	if _, err := wd.execute(verb, url, data); err != nil {
		return err
	}

	// TODO(minusnine): add a test for switching back to the original window.
	if name != startWindow && wd.w3cCompatible {
		if err := wd.SwitchWindow(startWindow); err != nil {
			return err
		}
	}

	return nil
}

func (wd *remoteWD) ResizeWindow(name string, width, height int) error {
	if !wd.w3cCompatible {
		return wd.modifyWindow(name, "POST", "size", map[string]int{
			"width":  width,
			"height": height,
		})
	}
	return wd.modifyWindow(name, "POST", "rect", map[string]float64{
		"width":  float64(width),
		"height": float64(height),
	})
}

func (wd *remoteWD) SwitchFrame(frame interface{}) error {
	params := map[string]interface{}{}
	switch f := frame.(type) {
	case WebElement, int, nil:
		params["id"] = f
	case string:
		if f == "" {
			params["id"] = nil
		} else if wd.w3cCompatible {
			e, err := wd.FindElement(ByID, f)
			if err != nil {
				return err
			}
			params["id"] = e
		} else { // Legacy, non W3C-spec behavior.
			params["id"] = f
		}
	default:
		return fmt.Errorf("invalid type %T", frame)
	}
	return wd.voidCommand("/session/%s/frame", params)
}

func (wd *remoteWD) ActiveElement() (WebElement, error) {
	verb := "GET"
	if wd.browser == "firefox" && wd.browserVersion.Major < 47 {
		verb = "POST"
	}
	url := wd.requestURL("/session/%s/element/active", wd.id)
	response, err := wd.execute(verb, url, nil)
	if err != nil {
		return nil, err
	}
	return wd.DecodeElement(response)
}

// ChromeDriver returns the expiration date as a float. Handle both formats
// via a type switch.
type cookie struct {
	Name   string      `json:"name"`
	Value  string      `json:"value"`
	Path   string      `json:"path"`
	Domain string      `json:"domain"`
	Secure bool        `json:"secure"`
	Expiry interface{} `json:"expiry"`
}

func (c cookie) sanitize() Cookie {
	sanitized := Cookie{
		Name:   c.Name,
		Value:  c.Value,
		Path:   c.Path,
		Domain: c.Domain,
		Secure: c.Secure,
	}
	switch expiry := c.Expiry.(type) {
	case int:
		if expiry > 0 {
			sanitized.Expiry = uint(expiry)
		}
	case float64:
		sanitized.Expiry = uint(expiry)
	}
	return sanitized
}

func (wd *remoteWD) GetCookie(name string) (Cookie, error) {
	if wd.browser == "chrome" {
		cs, err := wd.GetCookies()
		if err != nil {
			return Cookie{}, err
		}
		for _, c := range cs {
			if c.Name == name {
				return c, nil
			}
		}
		return Cookie{}, errors.New("cookie not found")
	}
	url := wd.requestURL("/session/%s/cookie/%s", wd.id, name)
	data, err := wd.execute("GET", url, nil)
	if err != nil {
		return Cookie{}, err
	}

	// GeckoDriver returns a list of cookies for this method. Try both a single
	// cookie and a list.
	//
	// https://github.com/mozilla/geckodriver/issues/761
	reply := new(struct{ Value cookie })
	if err := json.Unmarshal(data, reply); err == nil {
		return reply.Value.sanitize(), nil
	}
	listReply := new(struct{ Value []cookie })
	if err := json.Unmarshal(data, listReply); err != nil {
		return Cookie{}, err
	}
	if len(listReply.Value) == 0 {
		return Cookie{}, errors.New("no cookies returned")
	}
	return listReply.Value[0].sanitize(), nil
}

func (wd *remoteWD) GetCookies() ([]Cookie, error) {
	url := wd.requestURL("/session/%s/cookie", wd.id)
	data, err := wd.execute("GET", url, nil)
	if err != nil {
		return nil, err
	}

	reply := new(struct{ Value []cookie })
	if err := json.Unmarshal(data, reply); err != nil {
		return nil, err
	}

	cookies := make([]Cookie, len(reply.Value))
	for i, c := range reply.Value {
		sanitized := Cookie{
			Name:   c.Name,
			Value:  c.Value,
			Path:   c.Path,
			Domain: c.Domain,
			Secure: c.Secure,
		}
		switch expiry := c.Expiry.(type) {
		case int:
			if expiry > 0 {
				sanitized.Expiry = uint(expiry)
			}
		case float64:
			sanitized.Expiry = uint(expiry)
		}
		cookies[i] = sanitized
	}

	return cookies, nil
}

func (wd *remoteWD) AddCookie(cookie *Cookie) error {
	return wd.voidCommand("/session/%s/cookie", map[string]*Cookie{
		"cookie": cookie,
	})
}

func (wd *remoteWD) DeleteAllCookies() error {
	url := wd.requestURL("/session/%s/cookie", wd.id)
	_, err := wd.execute("DELETE", url, nil)
	return err
}

func (wd *remoteWD) DeleteCookie(name string) error {
	url := wd.requestURL("/session/%s/cookie/%s", wd.id, name)
	_, err := wd.execute("DELETE", url, nil)
	return err
}

// TODO(minusnine): add a test for Click.
func (wd *remoteWD) Click(button int) error {
	return wd.voidCommand("/session/%s/click", map[string]int{
		"button": button,
	})
}

// TODO(minusnine): add a test for DoubleClick.
func (wd *remoteWD) DoubleClick() error {
	return wd.voidCommand("/session/%s/doubleclick", nil)
}

// TODO(minusnine): add a test for ButtonDown.
func (wd *remoteWD) ButtonDown() error {
	return wd.voidCommand("/session/%s/buttondown", nil)
}

// TODO(minusnine): add a test for ButtonUp.
func (wd *remoteWD) ButtonUp() error {
	return wd.voidCommand("/session/%s/buttonup", nil)
}

func (wd *remoteWD) SendModifier(modifier string, isDown bool) error {
	if isDown {
		return wd.KeyDown(modifier)
	}
	return wd.KeyUp(modifier)
}

func (wd *remoteWD) keyAction(action, keys string) error {
	type keyAction struct {
		Type string `json:"type"`
		Key  string `json:"value"`
	}
	actions := make([]keyAction, 0, len(keys))
	for _, key := range keys {
		actions = append(actions, keyAction{
			Type: action,
			Key:  string(key),
		})
	}
	return wd.voidCommand("/session/%s/actions", map[string]interface{}{
		"actions": []interface{}{
			map[string]interface{}{
				"type":    "key",
				"id":      "default keyboard",
				"actions": actions,
			}},
	})
}

func (wd *remoteWD) KeyDown(keys string) error {
	// Selenium implemented the actions API but has not yet updated its new
	// session response.
	if !wd.w3cCompatible && !(wd.browser == "firefox" && wd.browserVersion.Major > 47) {
		return wd.voidCommand("/session/%s/keys", wd.processKeyString(keys))
	}
	return wd.keyAction("keyDown", keys)
}

func (wd *remoteWD) KeyUp(keys string) error {
	// Selenium implemented the actions API but has not yet updated its new
	// session response.
	if !wd.w3cCompatible && !(wd.browser == "firefox" && wd.browserVersion.Major > 47) {
		return wd.KeyDown(keys)
	}
	return wd.keyAction("keyUp", keys)
}

// TODO(minusnine): Implement PerformActions and ReleaseActions, for more
// direct access to the W3C specification.
func (wd *remoteWD) DismissAlert() error {
	return wd.voidCommand("/session/%s/alert/dismiss", nil)
}

func (wd *remoteWD) AcceptAlert() error {
	return wd.voidCommand("/session/%s/alert/accept", nil)
}

func (wd *remoteWD) AlertText() (string, error) {
	return wd.stringCommand("/session/%s/alert/text")
}

func (wd *remoteWD) SetAlertText(text string) error {
	data, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return err
	}

	return wd.voidCommand("/session/%s/alert/text", data)
}

func (wd *remoteWD) execScriptRaw(script string, args []interface{}, suffix string) ([]byte, error) {
	if args == nil {
		args = make([]interface{}, 0)
	}

	data, err := json.Marshal(map[string]interface{}{
		"script": script,
		"args":   args,
	})
	if err != nil {
		return nil, err
	}

	return wd.execute("POST", wd.requestURL("/session/%s/execute"+suffix, wd.id), data)
}

func (wd *remoteWD) execScript(script string, args []interface{}, suffix string) (interface{}, error) {
	response, err := wd.execScriptRaw(script, args, suffix)
	if err != nil {
		return nil, err
	}

	reply := new(struct{ Value interface{} })
	if err = json.Unmarshal(response, reply); err != nil {
		return nil, err
	}

	return reply.Value, nil
}

func (wd *remoteWD) ExecuteScript(script string, args []interface{}) (interface{}, error) {
	if !wd.w3cCompatible {
		return wd.execScript(script, args, "")
	}
	return wd.execScript(script, args, "/sync")
}

func (wd *remoteWD) ExecuteScriptAsync(script string, args []interface{}) (interface{}, error) {
	if !wd.w3cCompatible {
		return wd.execScript(script, args, "_async")
	}
	return wd.execScript(script, args, "/async")
}

func (wd *remoteWD) ExecuteScriptRaw(script string, args []interface{}) ([]byte, error) {
	if !wd.w3cCompatible {
		return wd.execScriptRaw(script, args, "")
	}
	return wd.execScriptRaw(script, args, "/sync")
}

func (wd *remoteWD) ExecuteScriptAsyncRaw(script string, args []interface{}) ([]byte, error) {
	if !wd.w3cCompatible {
		return wd.execScriptRaw(script, args, "_async")
	}
	return wd.execScriptRaw(script, args, "/async")
}

func (wd *remoteWD) Screenshot() ([]byte, error) {
	data, err := wd.stringCommand("/session/%s/screenshot")
	if err != nil {
		return nil, err
	}

	// Selenium returns a base64 encoded image.
	buf := []byte(data)
	decoder := base64.NewDecoder(base64.StdEncoding, bytes.NewBuffer(buf))
	return ioutil.ReadAll(decoder)
}

// Condition is an alias for a type that is passed as an argument
// for selenium.Wait(cond Condition) (error) function.
type Condition func(wd WebDriver) (bool, error)

const (
	// DefaultWaitInterval is the default polling interval for selenium.Wait
	// function.
	DefaultWaitInterval = 100 * time.Millisecond

	// DefaultWaitTimeout is the default timeout for selenium.Wait function.
	DefaultWaitTimeout = 60 * time.Second
)

func (wd *remoteWD) WaitWithTimeoutAndInterval(condition Condition, timeout, interval time.Duration) error {
	startTime := time.Now()

	for {
		done, err := condition(wd)
		if err != nil {
			return err
		}
		if done {
			return nil
		}

		if elapsed := time.Since(startTime); elapsed > timeout {
			return fmt.Errorf("timeout after %v", elapsed)
		}
		time.Sleep(interval)
	}
}

func (wd *remoteWD) WaitWithTimeout(condition Condition, timeout time.Duration) error {
	return wd.WaitWithTimeoutAndInterval(condition, timeout, DefaultWaitInterval)
}

func (wd *remoteWD) Wait(condition Condition) error {
	return wd.WaitWithTimeoutAndInterval(condition, DefaultWaitTimeout, DefaultWaitInterval)
}

func (wd *remoteWD) Log(typ log.Type) ([]log.Message, error) {
	url := wd.requestURL("/session/%s/log", wd.id)
	params := map[string]log.Type{
		"type": typ,
	}
	data, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	response, err := wd.execute("POST", url, data)
	if err != nil {
		return nil, err
	}

	c := new(struct {
		Value []struct {
			Timestamp int64
			Level     string
			Message   string
		}
	})
	if err = json.Unmarshal(response, c); err != nil {
		return nil, err
	}

	val := make([]log.Message, len(c.Value))
	for i, v := range c.Value {
		val[i] = log.Message{
			// n.b.: Chrome, which is the only browser that supports this API,
			// supplies timestamps in milliseconds since the Epoch.
			Timestamp: time.Unix(0, v.Timestamp*int64(time.Millisecond)),
			Level:     log.Level(v.Level),
			Message:   v.Message,
		}
	}

	return val, nil
}

type remoteWE struct {
	parent *remoteWD
	// Prior to the W3C specification, elements would be returned as a map with
	// the literal key "ELEMENT" and a value of a UUID. The W3C specification
	// specifies that this key has changed to an UUID-based string constant and
	// that the value is called a "reference". For ease of transition, we store
	// the "reference" in this now misnamed field.
	id string
}

func (elem *remoteWE) Click() error {
	urlTemplate := fmt.Sprintf("/session/%%s/element/%s/click", elem.id)
	return elem.parent.voidCommand(urlTemplate, nil)
}

func (elem *remoteWE) SendKeys(keys string) error {
	urlTemplate := fmt.Sprintf("/session/%%s/element/%s/value", elem.id)
	return elem.parent.voidCommand(urlTemplate, elem.parent.processKeyString(keys))
}

func (wd *remoteWD) processKeyString(keys string) interface{} {
	if !wd.w3cCompatible {
		chars := make([]string, len(keys))
		for i, c := range keys {
			chars[i] = string(c)
		}
		return map[string][]string{"value": chars}
	}
	return map[string]string{"text": keys}
}

func (elem *remoteWE) TagName() (string, error) {
	urlTemplate := fmt.Sprintf("/session/%%s/element/%s/name", elem.id)
	return elem.parent.stringCommand(urlTemplate)
}

func (elem *remoteWE) Text() (string, error) {
	urlTemplate := fmt.Sprintf("/session/%%s/element/%s/text", elem.id)
	return elem.parent.stringCommand(urlTemplate)
}

func (elem *remoteWE) Submit() error {
	urlTemplate := fmt.Sprintf("/session/%%s/element/%s/submit", elem.id)
	return elem.parent.voidCommand(urlTemplate, nil)
}

func (elem *remoteWE) Clear() error {
	urlTemplate := fmt.Sprintf("/session/%%s/element/%s/clear", elem.id)
	return elem.parent.voidCommand(urlTemplate, nil)
}

func (elem *remoteWE) MoveTo(xOffset, yOffset int) error {
	return elem.parent.voidCommand("/session/%s/moveto", map[string]interface{}{
		"element": elem.id,
		"xoffset": xOffset,
		"yoffset": yOffset,
	})
}

func (elem *remoteWE) FindElement(by, value string) (WebElement, error) {
	url := fmt.Sprintf("/session/%%s/element/%s/element", elem.id)
	response, err := elem.parent.find(by, value, "", url)
	if err != nil {
		return nil, err
	}

	return elem.parent.DecodeElement(response)
}

func (elem *remoteWE) FindElements(by, value string) ([]WebElement, error) {
	url := fmt.Sprintf("/session/%%s/element/%s/element", elem.id)
	response, err := elem.parent.find(by, value, "s", url)
	if err != nil {
		return nil, err
	}

	return elem.parent.DecodeElements(response)
}

func (elem *remoteWE) boolQuery(urlTemplate string) (bool, error) {
	return elem.parent.boolCommand(fmt.Sprintf(urlTemplate, elem.id))
}

func (elem *remoteWE) IsSelected() (bool, error) {
	return elem.boolQuery("/session/%%s/element/%s/selected")
}

func (elem *remoteWE) IsEnabled() (bool, error) {
	return elem.boolQuery("/session/%%s/element/%s/enabled")
}

func (elem *remoteWE) IsDisplayed() (bool, error) {
	return elem.boolQuery("/session/%%s/element/%s/displayed")
}

// TODO(minusnine): Add Property(name string) (string, error).

func (elem *remoteWE) GetAttribute(name string) (string, error) {
	template := "/session/%%s/element/%s/attribute/%s"
	urlTemplate := fmt.Sprintf(template, elem.id, name)

	return elem.parent.stringCommand(urlTemplate)
}

func round(f float64) int {
	if f < -0.5 {
		return int(f - 0.5)
	}
	if f > 0.5 {
		return int(f + 0.5)
	}
	return 0
}

func (elem *remoteWE) location(suffix string) (*Point, error) {
	if !elem.parent.w3cCompatible {
		wd := elem.parent
		path := "/session/%s/element/%s/location" + suffix
		url := wd.requestURL(path, wd.id, elem.id)
		response, err := wd.execute("GET", url, nil)
		if err != nil {
			return nil, err
		}
		reply := new(struct{ Value rect })
		if err := json.Unmarshal(response, reply); err != nil {
			return nil, err
		}
		return &Point{round(reply.Value.X), round(reply.Value.Y)}, nil
	}

	rect, err := elem.rect()
	if err != nil {
		return nil, err
	}
	return &Point{round(rect.X), round(rect.Y)}, nil
}

func (elem *remoteWE) Location() (*Point, error) {
	return elem.location("")
}

func (elem *remoteWE) LocationInView() (*Point, error) {
	return elem.location("_in_view")
}

func (elem *remoteWE) Size() (*Size, error) {
	if !elem.parent.w3cCompatible {
		wd := elem.parent
		url := wd.requestURL("/session/%s/element/%s/size", wd.id, elem.id)
		response, err := wd.execute("GET", url, nil)
		if err != nil {
			return nil, err
		}
		reply := new(struct{ Value rect })
		if err := json.Unmarshal(response, reply); err != nil {
			return nil, err
		}
		return &Size{round(reply.Value.Width), round(reply.Value.Height)}, nil
	}

	rect, err := elem.rect()
	if err != nil {
		return nil, err
	}

	return &Size{round(rect.Width), round(rect.Height)}, nil
}

type rect struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// rect implements the "Get Element Rect" method of the W3C standard.
func (elem *remoteWE) rect() (*rect, error) {
	wd := elem.parent
	url := wd.requestURL("/session/%s/element/%s/rect", wd.id, elem.id)
	response, err := wd.execute("GET", url, nil)
	if err != nil {
		return nil, err
	}
	r := new(struct{ Value rect })
	if err := json.Unmarshal(response, r); err != nil {
		return nil, err
	}
	return &r.Value, nil
}

func (elem *remoteWE) CSSProperty(name string) (string, error) {
	wd := elem.parent
	return wd.stringCommand(fmt.Sprintf("/session/%%s/element/%s/css/%s", elem.id, name))
}

func (elem *remoteWE) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		"ELEMENT":            elem.id,
		webElementIdentifier: elem.id,
	})
}

func (elem *remoteWE) Screenshot(scroll bool) ([]byte, error) {
	data, err := elem.parent.stringCommand(fmt.Sprintf("/session/%%s/element/%s/screenshot", elem.id))
	if err != nil {
		return nil, err
	}

	// Selenium returns a base64 encoded image.
	buf := []byte(data)
	decoder := base64.NewDecoder(base64.StdEncoding, bytes.NewBuffer(buf))
	return ioutil.ReadAll(decoder)
}

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
	"time"
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

const (
	// Success is status code that indicates the method was successful.
	Success = 0
	// DefaultURLPrefix is the default HTTP endpoint that offers the WebDriver
	// API.
	DefaultURLPrefix = "http://127.0.0.1:4444/wd/hub"
	// JSONType is JSON content type.
	JSONType = "application/json"
	// MaxRedirects is the maximum number of redirects to follow.
	MaxRedirects = 10
)

type remoteWD struct {
	id, urlPrefix string
	capabilities  Capabilities

	w3cCompatible bool
	browser       string
}

var httpClient *http.Client

// GetHTTPClient returns the default HTTP client.
func GetHTTPClient() *http.Client {
	return httpClient
}

func newRequest(method string, url string, data []byte) (*http.Request, error) {
	request, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	request.Header.Add("Accept", JSONType)

	return request, nil
}

func isRedirect(response *http.Response) bool {
	switch response.StatusCode {
	case 301, 302, 303, 307:
		return true
	}
	return false
}

func normalizeURL(n string, base string) (string, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("Failed to parse base URL %s with error %s", base, err)
	}
	nURL, err := baseURL.Parse(n)
	if err != nil {
		return "", fmt.Errorf("Failed to parse new URL %s with error %s", n, err)
	}
	return nURL.String(), nil
}

func (wd *remoteWD) requestURL(template string, args ...interface{}) string {
	return wd.urlPrefix + fmt.Sprintf(template, args...)
}

type serverReply struct {
	SessionID *string // SessionID can be nil.
	Value     json.RawMessage

	// The following fields were used prior to Selenium 3.0 for error state and
	// in ChromeDriver for additional information.
	Status int
	State  string

	Error
}

// Error contains information about a failure of a command.
type Error struct {
	Err        string `json:"error"`
	Message    string `json:"message"`
	Stacktrace string `json:"stacktrace"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Err, e.Message)
}

func (wd *remoteWD) execute(method, url string, data []byte) ([]byte, error) {
	debugLog("-> %s %s\n%s", method, filteredURL(url), data)
	request, err := newRequest(method, url, data)
	if err != nil {
		return nil, err
	}

	response, err := httpClient.Do(request)
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
		return nil, fmt.Errorf("got content type header %q, expected %q", fullCType, JSONType)
	}
	if cType != JSONType {
		return nil, fmt.Errorf("got content type %q, expected %q", cType, JSONType)
	}

	reply := new(serverReply)
	if err := json.Unmarshal(buf, reply); err != nil {
		if response.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("bad server reply status: %s", response.Status)
		}
		return nil, err
	}

	// Handle the W3C-compliant error format.
	if reply.Err != "" {
		return nil, &reply.Error
	}

	if reply.Status != Success {
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
		return nil, fmt.Errorf("%s: %s", shortMsg, longMsg.Message)
	}

	return buf, nil
}

// NewRemote creates new remote client, this will also start a new session.
// capabilities provides the desired capabilities. urlPrefix is the URL to the
// Selenium server, must be prefixed with protocol (http, https, ...).
//
// Providing an empty string for urlPrefix causes the DefaultURLPrefix to be
// used.
func NewRemote(capabilities Capabilities, urlPrefix string) (WebDriver, error) {
	if len(urlPrefix) == 0 {
		urlPrefix = DefaultURLPrefix
	}

	wd := &remoteWD{urlPrefix: urlPrefix, capabilities: capabilities}
	if _, err := wd.NewSession(); err != nil {
		return nil, err
	}
	return wd, nil
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

func (wd *remoteWD) voidCommand(urlTemplate string, params interface{}) error {
	var data []byte
	if params != nil {
		var err error
		data, err = json.Marshal(params)
		if err != nil {
			return err
		}
	}
	_, err := wd.execute("POST", wd.requestURL(urlTemplate, wd.id), data)
	return err
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

func (wd *remoteWD) NewSession() (string, error) {
	// Detect whether the remote end complies with the W3C specification:
	// non-compliant implementations use the top-level 'desiredCapabilities' JSON
	// key, whereas the specification mandates the 'capabilities' key.
	//
	// However, Selenium 3 currently does not implement this part of the specification.
	// https://github.com/SeleniumHQ/selenium/issues/2827
	attempts := []struct {
		params map[string]interface{}
	}{
		{map[string]interface{}{
			"capabilities": map[string]interface{}{
				"desiredCapabilities": wd.capabilities,
			},
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

		if reply.SessionID != nil {
			wd.id = *reply.SessionID
		} else if len(reply.Value) > 0 {
			value := new(struct {
				SessionID string
			})
			if err := json.Unmarshal(reply.Value, value); err != nil {
				return "", fmt.Errorf("error unmarshalling value: %v", err)
			}
			wd.id = value.SessionID
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
	return wd.voidCommand("/session/%s/timeouts/async_script", map[string]uint{
		"ms": uint(timeout / time.Millisecond),
	})
}

func (wd *remoteWD) SetImplicitWaitTimeout(timeout time.Duration) error {
	return wd.voidCommand("/session/%s/timeouts/implicit_wait", map[string]uint{
		"ms": uint(timeout / time.Millisecond),
	})
}

func (wd *remoteWD) SetPageLoadTimeout(timeout time.Duration) error {
	return wd.voidCommand("/session/%s/timeouts", map[string]interface{}{
		"ms":   uint(timeout / time.Millisecond),
		"type": "page load",
	})
}

func (wd *remoteWD) AvailableEngines() ([]string, error) {
	return wd.stringsCommand("/session/%s/ime/available_engines")
}

func (wd *remoteWD) ActiveEngine() (string, error) {
	return wd.stringCommand("/session/%s/ime/active_engine")
}

func (wd *remoteWD) IsEngineActivated() (bool, error) {
	return wd.boolCommand("/session/%s/ime/activated")
}

func (wd *remoteWD) DeactivateEngine() error {
	return wd.voidCommand("session/%s/ime/deactivate", nil)
}

func (wd *remoteWD) ActivateEngine(engine string) error {
	return wd.voidCommand("/session/%s/ime/activate", map[string]string{
		"engine": engine,
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
	return wd.stringCommand("/session/%s/window_handle")
}

func (wd *remoteWD) WindowHandles() ([]string, error) {
	return wd.stringsCommand("/session/%s/window_handles")
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

type element struct {
	Element string `json:"ELEMENT"`
}

func (wd *remoteWD) DecodeElement(data []byte) (WebElement, error) {
	reply := new(struct{ Value element })
	if err := json.Unmarshal(data, reply); err != nil {
		return nil, err
	}
	return &remoteWE{wd, reply.Value.Element}, nil
}

func (wd *remoteWD) FindElement(by, value string) (WebElement, error) {
	response, err := wd.find(by, value, "", "")
	if err != nil {
		return nil, err
	}
	return wd.DecodeElement(response)
}

func (wd *remoteWD) DecodeElements(data []byte) ([]WebElement, error) {
	reply := new(struct{ Value []element })
	if err := json.Unmarshal(data, reply); err != nil {
		return nil, err
	}

	elems := make([]WebElement, len(reply.Value))
	for i, elem := range reply.Value {
		elems[i] = &remoteWE{wd, elem.Element}
	}

	return elems, nil
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
	return wd.voidCommand("/session/%s/window", map[string]string{
		"name": name,
	})
}

func (wd *remoteWD) CloseWindow(name string) error {
	url := wd.requestURL("/session/%s/window", wd.id)
	_, err := wd.execute("DELETE", url, nil)
	return err
}

func (wd *remoteWD) MaximizeWindow(name string) error {
	if len(name) == 0 {
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

func (wd *remoteWD) ResizeWindow(name string, width, height int) error {
	if len(name) == 0 {
		var err error
		name, err = wd.CurrentWindowHandle()
		if err != nil {
			return err
		}
	}

	params := struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	}{
		width,
		height,
	}
	data, err := json.Marshal(params)
	if err != nil {
		return err
	}

	url := wd.requestURL("/session/%s/window/%s/size", wd.id, name)
	_, err = wd.execute("POST", url, data)
	return err
}

func (wd *remoteWD) SwitchFrame(frame interface{}) error {
	params := map[string]interface{}{}
	switch f := frame.(type) {
	case WebElement, int, nil:
		params["id"] = f
	case string:
		if f != "" {
			params["id"] = f
		} else {
			params["id"] = nil
		}
	default:
		return fmt.Errorf("invalid type %T", frame)
	}
	return wd.voidCommand("/session/%s/frame", params)
}

func (wd *remoteWD) ActiveElement() (WebElement, error) {
	url := wd.requestURL("/session/%s/element/active", wd.id)
	response, err := wd.execute("GET", url, nil)
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
	url := wd.requestURL("/session/%s/cookie/%s", wd.id, name)
	data, err := wd.execute("GET", url, nil)
	if err != nil {
		return Cookie{}, err
	}

	reply := new(struct{ Value cookie })
	if err := json.Unmarshal(data, reply); err != nil {
		return Cookie{}, err
	}
	return reply.Value.sanitize(), nil
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

func (wd *remoteWD) Click(button int) error {
	return wd.voidCommand("/session/%s/click", map[string]int{
		"button": button,
	})
}

func (wd *remoteWD) DoubleClick() error {
	return wd.voidCommand("/session/%s/doubleclick", nil)
}

func (wd *remoteWD) ButtonDown() error {
	return wd.voidCommand("/session/%s/buttondown", nil)
}

func (wd *remoteWD) ButtonUp() error {
	return wd.voidCommand("/session/%s/buttonup", nil)
}

func (wd *remoteWD) SendModifier(modifier string, isDown bool) error {
	return wd.voidCommand("/session/%s/modifier", map[string]interface{}{
		"value":  modifier,
		"isdown": isDown,
	})
}

func (wd *remoteWD) KeyDown(keys string) error {
	// TODO(minusnine): The W3C specification has changed this protocol, but
	// neither Firefox nor Chrome have implemented it yet.
	//
	// Marionette: https://bugzilla.mozilla.org/show_bug.cgi?id=1292178
	// GeckoDriver: https://github.com/mozilla/geckodriver/issues/159
	// Selenium 3: https://github.com/SeleniumHQ/selenium/issues/2285
	return wd.voidCommand("/session/%s/keys", processKeyString(keys))
}

func (wd *remoteWD) KeyUp(keys string) error {
	return wd.KeyDown(keys)
}

func (wd *remoteWD) DismissAlert() error {
	return wd.voidCommand("/session/%s/dismiss_alert", nil)
}

func (wd *remoteWD) AcceptAlert() error {
	return wd.voidCommand("/session/%s/accept_alert", nil)
}

func (wd *remoteWD) AlertText() (string, error) {
	return wd.stringCommand("/session/%s/alert_text")
}

func (wd *remoteWD) SetAlertText(text string) error {
	data, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return err
	}

	return wd.voidCommand("/session/%s/alert_text", data)
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
	return wd.execScript(script, args, "")
}

func (wd *remoteWD) ExecuteScriptAsync(script string, args []interface{}) (interface{}, error) {
	return wd.execScript(script, args, "_async")
}

func (wd *remoteWD) ExecuteScriptRaw(script string, args []interface{}) ([]byte, error) {
	return wd.execScriptRaw(script, args, "")
}

func (wd *remoteWD) ExecuteScriptAsyncRaw(script string, args []interface{}) ([]byte, error) {
	return wd.execScriptRaw(script, args, "_async")
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

func (wd *remoteWD) Log(typ LogType) ([]LogMessage, error) {
	url := wd.requestURL("/session/%s/log", wd.id)
	params := map[string]LogType{
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

	c := new(struct{ Value []LogMessage })
	if err = json.Unmarshal(response, c); err != nil {
		return nil, err
	}

	return c.Value, nil
}

type remoteWE struct {
	parent *remoteWD
	id     string
}

func (elem *remoteWE) Click() error {
	urlTemplate := fmt.Sprintf("/session/%%s/element/%s/click", elem.id)
	return elem.parent.voidCommand(urlTemplate, nil)
}

func (elem *remoteWE) SendKeys(keys string) error {
	urlTemplate := fmt.Sprintf("/session/%%s/element/%s/value", elem.id)
	return elem.parent.voidCommand(urlTemplate, processKeyString(keys))
}

func processKeyString(keys string) interface{} {
	chars := make([]string, len(keys))
	for i, c := range keys {
		chars[i] = string(c)
	}
	return map[string][]string{"value": chars}
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

func (elem *remoteWE) GetAttribute(name string) (string, error) {
	template := "/session/%%s/element/%s/attribute/%s"
	urlTemplate := fmt.Sprintf(template, elem.id, name)

	return elem.parent.stringCommand(urlTemplate)
}

func (elem *remoteWE) location(suffix string) (*Point, error) {
	wd := elem.parent
	path := "/session/%s/element/%s/location" + suffix
	url := wd.requestURL(path, wd.id, elem.id)
	response, err := wd.execute("GET", url, nil)
	if err != nil {
		return nil, err
	}
	reply := new(struct{ Value Point })
	if err := json.Unmarshal(response, reply); err != nil {
		return nil, err
	}

	return &reply.Value, nil
}

func (elem *remoteWE) Location() (*Point, error) {
	return elem.location("")
}

func (elem *remoteWE) LocationInView() (*Point, error) {
	return elem.location("_in_view")
}

func (elem *remoteWE) Size() (*Size, error) {
	wd := elem.parent
	url := wd.requestURL("/session/%s/element/%s/size", wd.id, elem.id)
	response, err := wd.execute("GET", url, nil)
	if err != nil {
		return nil, err
	}
	reply := new(struct{ Value Size })
	if err := json.Unmarshal(response, reply); err != nil {
		return nil, err
	}

	return &reply.Value, nil
}

func (elem *remoteWE) CSSProperty(name string) (string, error) {
	wd := elem.parent
	return wd.stringCommand(fmt.Sprintf("/session/%%s/element/%s/css/%s", elem.id, name))
}

func (elem *remoteWE) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		"ELEMENT": elem.id,
		"element-6066-11e4-a52e-4f735466cecf": elem.id,
	})
}

func init() {
	// http.Client doesn't copy request headers, and selenium requires that
	httpClient = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > MaxRedirects {
				return fmt.Errorf("too many redirects (%d)", len(via))
			}

			req.Header.Add("Accept", JSONType)
			return nil
		},
	}
}

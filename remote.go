/* Remote Selenium client implementation.

See http://code.google.com/p/selenium/wiki/JsonWireProtocol for wire protocol.
*/

package selenium

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

/* Errors returned by Selenium server. */
var remoteErrors = map[int]string{
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
	// Success of method
	Success          = 0
	DEFAULT_EXECUTOR = "http://127.0.0.1:4444/wd/hub"
	JSON_TYPE        = "application/json"
	MAX_REDIRECTS    = 10
)

type remoteWD struct {
	id, executor string
	capabilities Capabilities
	// FIXME
	// profile             BrowserProfile
}

/* Server reply */
type serverReply struct {
	SessionId *string // sessionId can be null
	Status    int
}

/* Various reply types, we use them to json.Unmarshal replies */
type statusReply struct {
	Value Status
}
type stringReply struct {
	Value *string
}
type stringsReply struct {
	Value []string
}
type boolReply struct {
	Value bool
}
type element struct {
	ELEMENT string
}
type elementReply struct {
	Value element
}
type elementsReply struct {
	Value []element
}
type cookiesReply struct {
	Value []Cookie
}
type locationReply struct {
	Value Point
}
type sizeReply struct {
	Value Size
}
type anyReply struct {
	Value interface{}
}
type capabilitiesReply struct {
	Value Capabilities
}
type logReply struct {
	Value []LogMessage
}

var httpClient *http.Client

func GetHTTPClient() *http.Client {
	return httpClient
}

func isMimeType(response *http.Response, mtype string) bool {
	if ctype, ok := response.Header["Content-Type"]; ok {
		return strings.HasPrefix(ctype[0], mtype)
	}

	return false
}

func newRequest(method string, url string, data []byte) (*http.Request, error) {
	request, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	request.Header.Add("Accept", JSON_TYPE)

	return request, nil
}

func cleanNils(buf []byte) {
	for i, b := range buf {
		if b == 0 {
			buf[i] = ' '
		}
	}
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
		return "", fmt.Errorf(
			"Failed to parse base URL %s with error %s", base, err)
	}
	nURL, err := baseURL.Parse(n)
	if err != nil {
		return "", fmt.Errorf("Failed to parse new URL %s with error %s", n, err)
	}
	return nURL.String(), nil
}

func (wd *remoteWD) requestURL(template string, args ...interface{}) string {
	path := fmt.Sprintf(template, args...)
	return wd.executor + path
}

func (wd *remoteWD) execute(method, url string, data []byte) ([]byte, error) {
	debugLog("-> %s %s\n%s", method, url, data)
	request, err := newRequest(method, url, data)
	if err != nil {
		return nil, err
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	buf, err := ioutil.ReadAll(response.Body)
	debugLog("<- %s [%s]\n%s",
		response.Status, response.Header["Content-Type"], buf)
	if err != nil {
		buf = []byte(response.Status)
	}

	if err != nil {
		return nil, errors.New(string(buf))
	}

	cleanNils(buf)
	if response.StatusCode >= 400 {
		reply := new(serverReply)
		err := json.Unmarshal(buf, reply)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Bad server reply status: %s", response.Status))
		}
		message, ok := remoteErrors[reply.Status]
		if !ok {
			message = fmt.Sprintf("unknown error - %d", reply.Status)
		}

		return nil, errors.New(message)
	}

	/* Some bug(?) in Selenium gets us nil values in output, json.Unmarshal is
	* not happy about that.
	 */
	if isMimeType(response, JSON_TYPE) {
		reply := new(serverReply)
		err := json.Unmarshal(buf, reply)
		if err != nil {
			return nil, err
		}

		if reply.Status != Success {
			message, ok := remoteErrors[reply.Status]
			if !ok {
				message = fmt.Sprintf("unknown error - %d", reply.Status)
			}

			return nil, errors.New(message)
		}
		return buf, err
	}

	// Nothing was returned, this is OK for some commands
	return buf, nil
}

/* Create new remote client, this will also start a new session.
   capabilities - the desired capabilities, see http://goo.gl/SNlAk
   executor - the URL to the Selenim server, *must* be prefixed with protocol (http,https...).
              Empty string means DEFAULT_EXECUTOR
*/
func NewRemote(capabilities Capabilities, executor string) (WebDriver, error) {

	if len(executor) == 0 {
		executor = DEFAULT_EXECUTOR
	}

	wd := &remoteWD{executor: executor, capabilities: capabilities}
	// FIXME: Handle profile

	_, err := wd.NewSession()
	if err != nil {
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

	reply := new(stringReply)
	err = json.Unmarshal(response, reply)
	if err != nil {
		return "", err
	}

	if reply.Value == nil {
		return "", fmt.Errorf("nil return value")
	}

	return *reply.Value, nil
}

func (wd *remoteWD) voidCommand(urlTemplate string, data []byte) error {
	url := wd.requestURL(urlTemplate, wd.id)
	_, err := wd.execute("POST", url, data)
	return err

}

func (wd remoteWD) stringsCommand(urlTemplate string) ([]string, error) {
	url := wd.requestURL(urlTemplate, wd.id)
	response, err := wd.execute("GET", url, nil)
	if err != nil {
		return nil, err
	}
	reply := new(stringsReply)
	err = json.Unmarshal(response, reply)
	if err != nil {
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

	reply := new(boolReply)
	err = json.Unmarshal(response, reply)
	if err != nil {
		return false, err
	}

	return reply.Value, nil
}

// WebDriver interface implementation

func (wd *remoteWD) Status() (*Status, error) {
	url := wd.requestURL("/status")
	reply, err := wd.execute("GET", url, nil)
	if err != nil {
		return nil, err
	}

	status := new(statusReply)
	err = json.Unmarshal(reply, status)
	if err != nil {
		return nil, err
	}

	return &status.Value, nil
}

func (wd *remoteWD) NewSession() (string, error) {
	message := map[string]interface{}{
		"sessionId":           nil,
		"desiredCapabilities": wd.capabilities,
	}
	data, err := json.Marshal(message)
	if err != nil {
		return "", nil
	}

	url := wd.requestURL("/session")
	response, err := wd.execute("POST", url, data)
	if err != nil {
		return "", err
	}

	reply := new(serverReply)
	json.Unmarshal(response, reply)

	wd.id = *reply.SessionId

	return wd.id, nil
}

func (wd *remoteWD) SessionId() string {
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

	c := new(capabilitiesReply)
	err = json.Unmarshal(response, c)
	if err != nil {
		return nil, err
	}

	return c.Value, nil
}

func (wd *remoteWD) SetAsyncScriptTimeout(timeout time.Duration) error {
	params := map[string]uint{
		"ms": uint(timeout / time.Millisecond),
	}

	data, err := json.Marshal(params)
	if err != nil {
		return err
	}

	return wd.voidCommand("/session/%s/timeouts/async_script", data)
}

func (wd *remoteWD) SetImplicitWaitTimeout(timeout time.Duration) error {
	params := map[string]uint{
		"ms": uint(timeout / time.Millisecond),
	}

	data, err := json.Marshal(params)
	if err != nil {
		return err
	}

	return wd.voidCommand("/session/%s/timeouts/implicit_wait", data)
}

func (wd *remoteWD) SetPageLoadTimeout(timeout time.Duration) error {
	params := map[string]interface{}{
		"ms":   uint(timeout / time.Millisecond),
		"type": "page load",
	}

	data, err := json.Marshal(params)
	if err != nil {
		return err
	}

	return wd.voidCommand("/session/%s/timeouts", data)
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
	params := map[string]string{
		"engine": engine,
	}

	data, err := json.Marshal(params)
	if err != nil {
		return err
	}

	return wd.voidCommand("/session/%s/ime/activate", data)
}

func (wd *remoteWD) Quit() error {
	url := wd.requestURL("/session/%s", wd.id)
	_, err := wd.execute("DELETE", url, nil)
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
	reply := new(stringReply)
	json.Unmarshal(response, reply)

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

	urlTemplate := url + suffix
	url = wd.requestURL(urlTemplate, wd.id)
	return wd.execute("POST", url, data)
}

func (wd *remoteWD) DecodeElement(data []byte) (WebElement, error) {
	reply := new(elementReply)
	err := json.Unmarshal(data, reply)
	if err != nil {
		return nil, err
	}

	elem := &remoteWE{wd, reply.Value.ELEMENT}
	return elem, nil
}

func (wd *remoteWD) FindElement(by, value string) (WebElement, error) {
	response, err := wd.find(by, value, "", "")
	if err != nil {
		return nil, err
	}

	return wd.DecodeElement(response)
}

func (wd *remoteWD) DecodeElements(data []byte) ([]WebElement, error) {
	reply := new(elementsReply)
	err := json.Unmarshal(data, reply)
	if err != nil {
		return nil, err
	}

	elems := make([]WebElement, len(reply.Value))
	for i, elem := range reply.Value {
		elems[i] = &remoteWE{wd, elem.ELEMENT}
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
	params := map[string]string{
		"name": name,
	}
	data, err := json.Marshal(params)
	if err != nil {
		return err
	}

	return wd.voidCommand("/session/%s/window", data)
}

func (wd *remoteWD) CloseWindow(name string) error {
	url := wd.requestURL("/session/%s/window", wd.id)
	_, err := wd.execute("DELETE", url, nil)
	return err
}

func (wd *remoteWD) MaximizeWindow(name string) error {
	var err error
	if len(name) == 0 {
		name, err = wd.CurrentWindowHandle()
		if err != nil {
			return err
		}
	}

	url := wd.requestURL("/session/%s/window/%s/maximize", wd.id, name)
	_, err = wd.execute("POST", url, nil)
	return err
}

func (wd *remoteWD) ResizeWindow(name string, width, height int) error {
	var err error
	if len(name) == 0 {
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

func (wd *remoteWD) SwitchFrame(frame string) error {
	params := map[string]string{
		"id": frame,
	}
	data, err := json.Marshal(params)
	if err != nil {
		return err
	}
	return wd.voidCommand("/session/%s/frame", data)
}

func (wd *remoteWD) ActiveElement() (WebElement, error) {
	url := wd.requestURL("/session/%s/element/active", wd.id)
	response, err := wd.execute("GET", url, nil)
	if err != nil {
		return nil, err
	}

	return wd.DecodeElement(response)
}

func (wd *remoteWD) GetCookies() ([]Cookie, error) {
	url := wd.requestURL("/session/%s/cookie", wd.id)
	data, err := wd.execute("GET", url, nil)
	if err != nil {
		return nil, err
	}

	reply := new(cookiesReply)
	err = json.Unmarshal(data, reply)
	if err != nil {
		return nil, err
	}

	return reply.Value, nil
}

func (wd *remoteWD) AddCookie(cookie *Cookie) error {
	params := map[string]*Cookie{
		"cookie": cookie,
	}
	data, err := json.Marshal(params)
	if err != nil {
		return err
	}

	return wd.voidCommand("/session/%s/cookie", data)
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
	params := map[string]int{
		"button": button,
	}
	data, err := json.Marshal(params)
	if err != nil {
		return err
	}
	return wd.voidCommand("/session/%s/click", data)
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
	params := map[string]interface{}{
		"value":  modifier,
		"isdown": isDown,
	}

	data, err := json.Marshal(params)
	if err != nil {
		return err
	}

	return wd.voidCommand("/session/%s/modifier", data)
}

func (wd *remoteWD) KeyDown(keys string) error {
	data, err := processKeyString(keys)
	if err != nil {
		return err
	}

	return wd.voidCommand("/session/%s/keys", data)
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
	params := map[string]string{
		"text": text,
	}
	data, err := json.Marshal(params)
	if err != nil {
		return err
	}

	return wd.voidCommand("/session/%s/alert_text", data)
}

func (wd *remoteWD) execScriptRaw(script string, args []interface{}, suffix string) ([]byte, error) {
	params := map[string]interface{}{
		"script": script,
		"args":   args,
	}

	data, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	template := "/session/%s/execute" + suffix
	url := wd.requestURL(template, wd.id)
	return wd.execute("POST", url, data)
}

func (wd *remoteWD) execScript(script string, args []interface{}, suffix string) (interface{}, error) {
	response, err := wd.execScriptRaw(script, args, suffix)
	if err != nil {
		return nil, err
	}

	reply := new(anyReply)
	err = json.Unmarshal(response, reply)
	if err != nil {
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

	// Selenium returns base64 encoded image
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

	c := new(logReply)
	err = json.Unmarshal(response, c)
	if err != nil {
		return nil, err
	}

	return c.Value, nil
}

// WebElement interface implementation

type remoteWE struct {
	parent *remoteWD
	id     string
}

func (elem *remoteWE) Click() error {
	urlTemplate := fmt.Sprintf("/session/%%s/element/%s/click", elem.id)
	return elem.parent.voidCommand(urlTemplate, nil)
}

func (elem *remoteWE) SendKeys(keys string) error {
	data, err := processKeyString(keys)
	if err != nil {
		return err
	}

	urlTemplate := fmt.Sprintf("/session/%%s/element/%s/value", elem.id)
	return elem.parent.voidCommand(urlTemplate, data)
}

func processKeyString(keys string) ([]byte, error) {
	chars := make([]string, len(keys))
	for i, c := range keys {
		chars[i] = string(c)
	}
	params := map[string][]string{
		"value": chars,
	}

	data, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	return data, nil
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
	params := map[string]interface{}{
		"element": elem.id,
		"xoffset": xOffset,
		"yoffset": yOffset,
	}
	data, err := json.Marshal(params)
	if err != nil {
		return err
	}
	return elem.parent.voidCommand("/session/%s/moveto", data)
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
	url := fmt.Sprintf(urlTemplate, elem.id)
	return elem.parent.boolCommand(url)
}

// Porperties
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
	reply := new(locationReply)
	err = json.Unmarshal(response, reply)
	if err != nil {
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
	reply := new(sizeReply)
	err = json.Unmarshal(response, reply)
	if err != nil {
		return nil, err
	}

	return &reply.Value, nil
}

func (elem *remoteWE) CSSProperty(name string) (string, error) {
	wd := elem.parent
	urlTemplate := fmt.Sprintf("/session/%s/element/%s/css/%s", wd.id, elem.id, name)
	return elem.parent.stringCommand(urlTemplate)
}

func init() {
	// http.Client doesn't copy request headers, and selenium requires that
	httpClient = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > MAX_REDIRECTS {
				return fmt.Errorf("too many redirects (%d)", len(via))
			}

			req.Header.Add("Accept", JSON_TYPE)
			return nil
		},
	}
}

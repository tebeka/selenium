/* Remote Selenium client implementation.

See http://code.google.com/p/selenium/wiki/JsonWireProtocol for wire protocol.
*/

package selenium

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"http"
	"strings"
	"os"
	"io/ioutil"
	"json"
)

/* Errors returned by Selenium server. */
var errors = map[int]string{
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
	SUCCESS          = 0
	DEFAULT_EXECUTOR = "http://127.0.0.1:4444/wd/hub"
	JSON_TYPE        = "application/json"
)

type remoteWD struct {
	id, executor string
	capabilities *Capabilities
	// FIXME
	// profile             BrowserProfile
}

type serverReply struct {
	SessionId *string
	Status    int
}

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

func isMimeType(response *http.Response, mtype string) bool {
	if ctype, ok := response.Header["Content-Type"]; ok {
		return strings.HasPrefix(ctype[0], mtype)
	}

	return false
}

func newRequest(method string, url string, data []byte) (*http.Request, os.Error) {
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

func (wd *remoteWD) requestURL(template string, args ...interface{}) string {
	path := fmt.Sprintf(template, args...)
	return wd.executor + path
}


func (wd *remoteWD) execute(method, url string, data []byte) ([]byte, os.Error) {
	request, err := newRequest(method, url, data)
	if err != nil {
		return nil, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	// http.Client don't follow POST redirects ....
	if (method == "POST") && isRedirect(response) {
		url := response.Header["Location"][0]
		request, _ = newRequest("GET", url, nil)
		response, err = http.DefaultClient.Do(request)
		if err != nil {
			return nil, err
		}
	}

	buf, err := ioutil.ReadAll(response.Body)
	if err != nil {
		buf = []byte(response.Status)
	}

	if (err != nil) || (response.StatusCode >= 400) {
		return nil, os.NewError(string(buf))
	}

	cleanNils(buf)

	reply := new(serverReply)
	if isMimeType(response, JSON_TYPE) {
		err := json.Unmarshal(buf, reply)
		if err != nil {
			return nil, err
		}

		if reply.Status != SUCCESS {
			message, ok := errors[reply.Status]
			if !ok {
				message = fmt.Sprintf("unknown error - %d", reply.Status)
			}

			return nil, os.NewError(message)
		}
		return buf, err
	}

	ctype, ok := response.Header["Content-Type"]
	if ok {
		err := os.NewError(fmt.Sprintf("unknown reply content type: %v", ctype))
		return nil, err
	}

	// Nothing was returned, this is OK for some commands
	return nil, nil
}

func NewRemote(capabilities *Capabilities, executor string, profileDir string) (WebDriver, os.Error) {

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

func (wd *remoteWD) Status() (*Status, os.Error) {
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

func (wd *remoteWD) NewSession() (string, os.Error) {
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

func (wd *remoteWD) Quit() os.Error {
	url := wd.requestURL("/session/%s", wd.id)
	_, err := wd.execute("DELETE", url, nil)
	if err == nil {
		wd.id = ""
	}

	return err
}

func (wd *remoteWD) stringCommand(urlTemplate string) (string, os.Error) {
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

	return *reply.Value, nil
}

func (wd *remoteWD) CurrentWindowHandle() (string, os.Error) {
	return wd.stringCommand("/session/%s/window_handle")
}

func (wd *remoteWD) WindowHandles() ([]string, os.Error) {
	url := wd.requestURL("/session/%s/window_handles", wd.id)
	response, err := wd.execute("GET", url, nil)
	if err != nil {
		return nil, err
	}
	reply := new(stringsReply)
	json.Unmarshal(response, reply)

	return reply.Value, nil
}

func (wd *remoteWD) CurrentURL() (string, os.Error) {
	url := wd.requestURL("/session/%s/url", wd.id)
	response, err := wd.execute("GET", url, nil)
	if err != nil {
		return "", err
	}
	reply := new(stringReply)
	json.Unmarshal(response, reply)

	return *reply.Value, nil

}

func (wd *remoteWD) Get(url string) os.Error {
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

func (wd *remoteWD) voidCommand(urlTemplate string) os.Error {
	url := wd.requestURL(urlTemplate, wd.id)
	_, err := wd.execute("POST", url, nil)
	return err

}

func (wd *remoteWD) Forward() os.Error {
	return wd.voidCommand("/session/%s/forward")
}

func (wd *remoteWD) Back() os.Error {
	return wd.voidCommand("/session/%s/back")
}

func (wd *remoteWD) Refresh() os.Error {
	return wd.voidCommand("/session/%s/refresh")
}

func (wd *remoteWD) Title() (string, os.Error) {
	return wd.stringCommand("/session/%s/title")
}

func (wd *remoteWD) PageSource() (string, os.Error) {
	return wd.stringCommand("/session/%s/source")
}

func (wd *remoteWD) find(by, value, suffix, url string) ([]byte, os.Error) {
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

func decodeElement(wd *remoteWD, data []byte) (WebElement, os.Error) {
	reply := new(elementReply)
	err := json.Unmarshal(data, reply)
	if err != nil {
		return nil, err
	}

	elem := &remoteWE{wd, reply.Value.ELEMENT}
	return elem, nil
}

func (wd *remoteWD) FindElement(by, value string) (WebElement, os.Error) {
	response, err := wd.find(by, value, "", "")
	if err != nil {
		return nil, err
	}

	return decodeElement(wd, response)
}

func decodeElements(wd *remoteWD, data []byte) ([]WebElement, os.Error) {
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

func (wd *remoteWD) FindElements(by, value string) ([]WebElement, os.Error) {
	response, err := wd.find(by, value, "s", "")
	if err != nil {
		return nil, err
	}

	return decodeElements(wd, response)
}

func (wd *remoteWD) Close() os.Error {
	url := wd.requestURL("/session/%s/window", wd.id)
	_, err := wd.execute("DELETE", url, nil)
	return err
}

func (wd *remoteWD) SwitchWindow(name string) os.Error {
	params := map[string]string{
		"name": name,
	}
	data, err := json.Marshal(params)
	if err != nil {
		return err
	}
	url := wd.requestURL("/session/%s/window", wd.id)
	_, err = wd.execute("POST", url, data)
	return err
}

func (wd *remoteWD) SwitchFrame(frame string) os.Error {
	params := map[string]string{
		"id": frame,
	}
	data, err := json.Marshal(params)
	if err != nil {
		return err
	}
	url := wd.requestURL("/session/%s/frame", wd.id)
	_, err = wd.execute("POST", url, data)
	return err
}

func (wd *remoteWD) ActiveElement() (WebElement, os.Error) {
	url := wd.requestURL("/session/%s/element/active", wd.id)
	response, err := wd.execute("GET", url, nil)

	reply := new(elementReply)
	err = json.Unmarshal(response, reply)
	if err != nil {
		return nil, err
	}

	elem := &remoteWE{wd, reply.Value.ELEMENT}
	return elem, nil
}

func (wd *remoteWD) GetCookies() ([]Cookie, os.Error) {
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

func (wd *remoteWD) AddCookie(cookie *Cookie) os.Error {
	params := map[string]*Cookie{
		"cookie": cookie,
	}
	data, err := json.Marshal(params)
	if err != nil {
		return err
	}

	url := wd.requestURL("/session/%s/cookie", wd.id)
	_, err = wd.execute("POST", url, data)
	return err
}

func (wd *remoteWD) DeleteAllCookies() os.Error {
	url := wd.requestURL("/session/%s/cookie", wd.id)
	_, err := wd.execute("DELETE", url, nil)
	return err
}

func (wd *remoteWD) DeleteCookie(name string) os.Error {
	url := wd.requestURL("/session/%s/cookie/%s", wd.id, name)
	_, err := wd.execute("DELETE", url, nil)
	return err
}

func (wd *remoteWD) Click(button int) os.Error {
	params := map[string]int{
		"button": button,
	}
	data, err := json.Marshal(params)
	if err != nil {
		return err
	}
	url := wd.requestURL("/session/%s/click", wd.id)
	_, err = wd.execute("POST", url, data)
	return err
}

func (wd *remoteWD) DoubleClick() os.Error {
	return wd.voidCommand("/session/%s/doubleclick")
}

func (wd *remoteWD) ButtonDown() os.Error {
	return wd.voidCommand("/session/%s/buttondown")
}

func (wd *remoteWD) ButtonUp() os.Error {
	return wd.voidCommand("/session/%s/buttonup")
}

func (wd *remoteWD) SendModifier(modifier string, isDown bool) os.Error {
	params := map[string]interface{}{
		"value":  modifier,
		"isdown": isDown,
	}

	data, err := json.Marshal(params)
	if err != nil {
		return err
	}

	url := wd.requestURL("/session/%s/modifier", wd.id)
	_, err = wd.execute("POST", url, data)
	return err
}

func (wd *remoteWD) DismissAlert() os.Error {
	return wd.voidCommand("/session/%s/dismiss_alert")
}


func (wd *remoteWD) AcceptAlert() os.Error {
	return wd.voidCommand("/session/%s/accept_alert")
}


func (wd *remoteWD) AlertText() (string, os.Error) {
	return wd.stringCommand("/session/%s/alert_text")
}

func (wd *remoteWD) SetAlertText(text string) os.Error {
	params := map[string]string{
		"text": text,
	}
	data, err := json.Marshal(params)
	if err != nil {
		return err
	}

	url := wd.requestURL("/session/%s/alert_text", wd.id)
	_, err = wd.execute("POST", url, data)
	return err
}

func (wd *remoteWD) execScript(script string, args []interface{}, suffix string) (interface{}, os.Error) {
	params := map[string]interface{} {
		"script": script,
		"args": args,
	}

	data, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	template := "/session/%s/execute" + suffix
	url := wd.requestURL(template, wd.id)
	response, err := wd.execute("POST", url, data)
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

func (wd *remoteWD) ExecuteScript(script string, args []interface{}) (interface{}, os.Error) {
	return wd.execScript(script, args, "")
}

func (wd *remoteWD) ExecuteScriptAsync(script string, args []interface{}) (interface{}, os.Error) {
	return wd.execScript(script, args, "_async")
}

func (wd *remoteWD) Screenshot() ([]byte, os.Error) {
	data, err := wd.stringCommand("/session/%s/screenshot")
	if err != nil {
		return nil, err
	}

	// Selenium returns base64 encoded image
	buf := []byte(data)
	decoder := base64.NewDecoder(base64.StdEncoding, bytes.NewBuffer(buf))
	return ioutil.ReadAll(decoder)
}

// Elements

type remoteWE struct {
	parent *remoteWD
	id     string
}

func (elem *remoteWE) Click() os.Error {
	urlTemplate := fmt.Sprintf("/session/%%s/element/%s/click", elem.id)
	return elem.parent.voidCommand(urlTemplate)
}

func (elem *remoteWE) SendKeys(keys string) os.Error {
	sid := elem.parent.id
	urlTemplate := fmt.Sprintf("/session/%s/element/%s/value", sid, elem.id)
	url := elem.parent.requestURL(urlTemplate)

	chars := make([]string, len(keys))
	for i, c := range keys {
		chars[i] = string(c)
	}
	params := map[string][]string{
		"value": chars,
	}

	data, err := json.Marshal(params)
	if err != nil {
		return err
	}

	_, err = elem.parent.execute("POST", url, data)
	return err
}


func (elem *remoteWE) TagName() (string, os.Error) {
	urlTemplate := fmt.Sprintf("/session/%%s/element/%s/name", elem.id)
	return elem.parent.stringCommand(urlTemplate)
}

func (elem *remoteWE) Text() (string, os.Error) {
	urlTemplate := fmt.Sprintf("/session/%%s/element/%s/text", elem.id)
	return elem.parent.stringCommand(urlTemplate)
}

func (elem *remoteWE) Submit() os.Error {
	urlTemplate := fmt.Sprintf("/session/%%s/element/%s/submit", elem.id)
	return elem.parent.voidCommand(urlTemplate)
}

func (elem *remoteWE) Clear() os.Error {
	urlTemplate := fmt.Sprintf("/session/%%s/element/%s/clear", elem.id)
	return elem.parent.voidCommand(urlTemplate)
}

func (elem *remoteWE) MoveTo(xOffset, yOffset int) os.Error {
	params := map[string]interface{}{
		"element": elem.id,
		"xoffset": xOffset,
		"yoffset": yOffset,
	}
	data, err := json.Marshal(params)
	if err != nil {
		return err
	}
	url := elem.parent.requestURL("/session/%s/moveto", elem.parent.id)
	_, err = elem.parent.execute("POST", url, data)
	return err
}

func (elem *remoteWE) FindElement(by, value string) (WebElement, os.Error) {
	url := fmt.Sprintf("/session/%%s/element/%s/element", elem.id)
	response, err := elem.parent.find(by, value, "", url)
	if err != nil {
		return nil, err
	}

	return decodeElement(elem.parent, response)
}

func (elem *remoteWE) FindElements(by, value string) ([]WebElement, os.Error) {
	url := fmt.Sprintf("/session/%%s/element/%s/element", elem.id)
	response, err := elem.parent.find(by, value, "s", url)
	if err != nil {
		return nil, err
	}

	return decodeElements(elem.parent, response)
}

func (elem *remoteWE) boolQuery(urlTemplate string) (bool, os.Error) {
	wd := elem.parent
	url := wd.requestURL(urlTemplate, wd.id, elem.id)
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

// Porperties
func (elem *remoteWE) IsSelected() (bool, os.Error) {
	return elem.boolQuery("/session/%s/element/%s/selected")
}

func (elem *remoteWE) IsEnabled() (bool, os.Error) {
	return elem.boolQuery("/session/%s/element/%s/enabled")
}

func (elem *remoteWE) IsDiaplayed() (bool, os.Error) {
	return elem.boolQuery("/session/%s/element/%s/displayed")
}

func (elem *remoteWE) GetAttribute(name string) (string, os.Error) {
	wd := elem.parent
	url := wd.requestURL("/session/%s/element/%s/attribute/%s", wd.id, elem.id, name)
	response, err := wd.execute("GET", url, nil)
	if err != nil {
		return "", err
	}
	reply := new(stringReply)
	err = json.Unmarshal(response, reply)
	if err != nil {
		return "", err
	}

	return *reply.Value, nil
}

func (elem *remoteWE) location(suffix string) (*Point, os.Error) {
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

func (elem *remoteWE) Location() (*Point, os.Error) {
	return elem.location("")
}

func (elem *remoteWE) LocationInView() (*Point, os.Error) {
	return elem.location("_in_view")
}

func (elem *remoteWE) Size() (*Size, os.Error) {
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

func (elem *remoteWE) CSSProperty(name string) (string, os.Error) {
	wd := elem.parent
	urlTemplate := fmt.Sprintf("/session/%s/element/%s/css/%s", wd.id, elem.id, name)
	return elem.parent.stringCommand(urlTemplate)
}

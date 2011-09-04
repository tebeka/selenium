/* Remote Selenium client */
package selenium

import (
	"bytes"
	"fmt"
	"http"
	"strings"
	"os"
	"io/ioutil"
	"json"
)

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
	SessionId, Executor string
	Capabilities        *Capabilities
	profile             BrowserProfile
}

type serverReply struct {
	SessionId *string
	Status    int
	//Value string
}

type statusReply struct {
	Value Status
}
type stringReply struct {
	Value string
}
type stringsReply struct {
	Value []string
}
type element struct {
	ELEMENT string
}
type elementReply struct {
	Value element
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
	return wd.Executor + path
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

		if reply.Status != 0 {
			message, ok := errors[reply.Status]
			if !ok {
				message = fmt.Sprintf("unknown error - %d", reply.Status)
			}

			return nil, os.NewError(message)
		}
		return buf, err
	} else if isMimeType(response, "image/png") {
		// FIXME: Handle images
	}

	ctype, ok := response.Header["Content-Type"]
	if ok {
		err := os.NewError(fmt.Sprintf("unknown reply content type: %v", ctype))
		return nil, err
	}

	// Nothing was returned, this is OK for some commands
	return nil, nil

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
		"desiredCapabilities": wd.Capabilities,
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

	wd.SessionId = *reply.SessionId

	return wd.SessionId, nil
}

func (wd *remoteWD) Quit() os.Error {
	url := wd.requestURL("/session/%s", wd.SessionId)
	_, err := wd.execute("DELETE", url, nil)
	if err == nil {
		wd.SessionId = ""
	}

	return err
}

func (wd *remoteWD) stringCommand(urlTemplate string) (string, os.Error) {
	url := wd.requestURL(urlTemplate, wd.SessionId)
	response, err := wd.execute("GET", url, nil)
	if err != nil {
		return "", err
	}

	reply := new(stringReply)
	err = json.Unmarshal(response, reply)
	if err != nil {
		return "", err
	}

	return reply.Value, nil
}

func (wd *remoteWD) CurrentWindowHandle() (string, os.Error) {
	return wd.stringCommand("/session/%s/window_handle")
}

func (wd *remoteWD) WindowHandles() ([]string, os.Error) {
	url := wd.requestURL("/session/%s/window_handles", wd.SessionId)
	response, err := wd.execute("GET", url, nil)
	if err != nil {
		return nil, err
	}
	reply := new(stringsReply)
	json.Unmarshal(response, reply)

	return reply.Value, nil
}

func (wd *remoteWD) CurrentURL() (string, os.Error) {
	url := wd.requestURL("/session/%s/url", wd.SessionId)
	response, err := wd.execute("GET", url, nil)
	if err != nil {
		return "", err
	}
	reply := new(stringReply)
	json.Unmarshal(response, reply)

	return reply.Value, nil

}

func (wd *remoteWD) Get(url string) os.Error {
	requestURL := wd.requestURL("/session/%s/url", wd.SessionId)
	params := map[string]string {
		"url" : url,
	}
	data, err := json.Marshal(params)
	if err != nil {
		return err
	}
	_, err = wd.execute("POST", requestURL, data)

	return err
}

func (wd *remoteWD) voidCommand(urlTemplate string) os.Error {
	url := wd.requestURL(urlTemplate, wd.SessionId)
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

func (wd *remoteWD) FindElement(by, value string) (WebElement, os.Error) {
	params := map[string]string {
		"using" : by,
		"value" : value,
	}
	data, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	url := wd.requestURL("/session/%s/element", wd.SessionId)
	response, err := wd.execute("POST", url, data)
	if err != nil {
		return nil, err
	}

	reply := new(elementReply)
	err = json.Unmarshal(response, reply)
	if err != nil {
		return nil, err
	}

	elem := &remoteWE{wd, reply.Value.ELEMENT}
	return elem, nil
}

func NewRemote(capabilities *Capabilities, executor string,
		 profile BrowserProfile) (WebDriver, os.Error) {

	if len(executor) == 0 {
		executor = DEFAULT_EXECUTOR
	}

	// FIXME: Handle profile
	wd := &remoteWD{Executor: executor,
		Capabilities: capabilities,
		profile:      profile}

	wd.NewSession()


	return wd, nil
}

type remoteWE struct {
	parent *remoteWD
	id string
}

func (elem *remoteWE) Click() os.Error {
	urlTemplate := fmt.Sprintf("/session/%%s/element/%s/click", elem.id)
	return elem.parent.voidCommand(urlTemplate)
}

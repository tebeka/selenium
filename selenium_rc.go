package selenium

import (
	"fmt"
	"http"
	"io/ioutil"
	"os"
	"url"
)

type rcClient struct {
	url                      string
	startCommand, browserUrl string
	sessionId                string
}

func NewSeleniumRC(host string, port int, startCommand, browserUrl string) SeleniumRC {
	url := fmt.Sprintf("http://%s:%d/selenium-server/driver/", host, port)
	return &rcClient{url, startCommand, browserUrl, ""}
}

func (rc *rcClient) do(command string, args ...string) (string, os.Error) {
	values := url.Values{}
	values.Add("cmd", command)
	for i, arg := range args {
		values.Add(fmt.Sprintf("%d", i+1), arg)
	}

	debugLog("-> %s %s", rc.url, values.Encode())
	response, err := http.DefaultClient.PostForm(rc.url, values)
	if err != nil {
		return "", err
	}

	buf, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	msg := string(buf)
	debugLog("<- %s", msg)
	if msg[:2] != "OK" {
		return "", os.NewError(msg)
	}

	return msg[3:], nil
}

func (rc *rcClient) void(command string, args ...string) os.Error {
	_, err := rc.do(command, args...)
	return err
}

func (rc *rcClient) getBool(command string, args ...string) (bool, os.Error) {
	reply, err := rc.do(command, args...)
	if err != nil {
		return false, err
	}
	switch reply {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, os.NewError(fmt.Sprintf("bad bool reply - %s", reply))
	}

	// Make the compiler happy
	return false, nil
}

func locStr(x, y int) string {
	return fmt.Sprintf("%d,%d", x, y)
}

func (rc *rcClient) doAt(command string, x, y int) os.Error {
	return rc.void(command, locStr(x, y))
}

// SeleniumRC interface

func (rc *rcClient) Start() (string, os.Error) {
	if rc.sessionId != "" {
		return "", os.NewError("Already started")
	}
	id, err := rc.do("getNewBrowserSession", rc.startCommand, rc.browserUrl)
	rc.sessionId = id
	return id, err
}

func (rc *rcClient) Stop() os.Error {
	_, err := rc.do("testComplete")
	if err != nil {
		return err
	}
	rc.sessionId = ""
	return nil
}


func (rc *rcClient) Click(locator string) os.Error {
	_, err := rc.do("click", locator)
	return err
}

func (rc *rcClient) DoubleClick(locator string) os.Error {
	_, err := rc.do("doubleClick", locator)
	return err
}

func (rc *rcClient) ContextMenu(locator string) os.Error {
	_, err := rc.do("contextMenu", locator)
	return err
}

func (rc *rcClient) ClickAt(locator string, x, y int) os.Error {
	return rc.doAt("clickAt", x, y)
}

func (rc *rcClient) DoubleClickAt(locator string, x, y int) os.Error {
	return rc.doAt("doubleClickAt", x, y)
}

func (rc *rcClient) ContextMenuAt(locator string, x, y int) os.Error {
	return rc.doAt("contextMenuAt", x, y)
}

func (rc *rcClient) FireEvent(locator, event string) os.Error {
	return rc.void("fireEvent", locator, event)
}

func (rc *rcClient) Focus(locator string) os.Error {
	return rc.void("focus", locator)
}

func (rc *rcClient) KeyPress(locator, keys string) os.Error {
	return rc.void("keyPress", locator, keys)
}

func (rc *rcClient) KeyDown(locator, keys string) os.Error {
	return rc.void("keyDown", locator, keys)
}

func (rc *rcClient) KeyUp(locator, keys string) os.Error {
	return rc.void("keyUp", locator, keys)
}

func (rc *rcClient) MouseOver(locator string) os.Error {
	return rc.void("mouseOver", locator)
}

func (rc *rcClient) MouseOut(locator string) os.Error {
	return rc.void("mouseOut", locator)
}

func (rc *rcClient) MouseDown(locator string) os.Error {
	return rc.void("mouseDown", locator)
}

func (rc *rcClient) MouseUp(locator string) os.Error {
	return rc.void("mouseUp", locator)
}

func (rc *rcClient) MouseDownRight(locator string) os.Error {
	return rc.void("mouseDownRight", locator)
}

func (rc *rcClient) MouseDownAt(locator string, x, y int) os.Error {
	_, err := rc.do("mouseDownAt", locator, locStr(x, y))
	return err
}

func (rc *rcClient) MouseDownRightAt(locator string, x, y int) os.Error {
	_, err := rc.do("mouseDownRightAt", locator, locStr(x, y))
	return err
}

func (rc *rcClient) Type(locator, value string) os.Error {
	return rc.void("type", locator, value)
}

func (rc *rcClient) TypeKeys(locator, value string) os.Error {
	return rc.void("typeKeys", locator, value)
}

func (rc *rcClient) Check(locator string) os.Error {
	return rc.void("check", locator)
}

func (rc *rcClient) Uncheck(locator string) os.Error {
	return rc.void("uncheck", locator)
}

func (rc *rcClient) Select(locator, option string) os.Error {
	return rc.void("select", locator, option)
}

func (rc *rcClient) Submit(form string) os.Error {
	return rc.void("submit", form)
}

func (rc *rcClient) Open(url string, ignoreResponse bool) os.Error {
	return rc.void("open", url, fmt.Sprintf("%v", ignoreResponse))
}

func (rc *rcClient) Back() os.Error {
	return rc.void("goBack")
}

func (rc *rcClient) Refresh() os.Error {
	return rc.void("refresh")
}

func (rc *rcClient) Close() os.Error {
	return rc.void("close")
}

func (rc *rcClient) Location() (string, os.Error) {
	return rc.do("getLocation")
}

func (rc *rcClient) Title() (string, os.Error) {
	return rc.do("getTitle")
}

func (rc *rcClient) Body() (string, os.Error) {
	return rc.do("getBodyText")
}

func (rc *rcClient) Value(locator string) (string, os.Error) {
	return rc.do("getValue", locator)
}

func (rc *rcClient) Text(locator string) (string, os.Error) {
	return rc.do("getText", locator)
}

func (rc *rcClient) Highlight(locator string) os.Error {
	return rc.void("highlight", locator)
}

func (rc *rcClient) IsChecked(locator string) (bool, os.Error) {
	return rc.getBool("isChecked", locator)
}

func (rc *rcClient) Attribute(locator string) (string, os.Error) {
	return rc.do("getAttribute", locator)
}

func (rc *rcClient) IsPresent(locator string) (bool, os.Error) {
	return rc.getBool("isElementPresent", locator)
}

func (rc *rcClient) IsVisible(locator string) (bool, os.Error) {
	return rc.getBool("isVisible", locator)
}

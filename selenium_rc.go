package selenium

import (
	"fmt"
	"http"
	"io/ioutil"
	"os"
	"url"
)

type SeleniumRC interface {
	Start() (string, os.Error)
}

type rcClient struct {
	url                      string
	startCommand, browserUrl string
	sessionId                string
}

func NewSeleniumRC(host string, port int, startCommand, browserUrl string) SeleniumRC {
	url := fmt.Sprintf("%s:%d/selenium-server/driver/", host, port)
	return &rcClient{url, startCommand, browserUrl, ""}
}

func (rc *rcClient) do(command string, args ...string) (string, os.Error) {
	values := url.Values{}
	values.Add("cmd", "command")
	for i, arg := range args {
		values.Add(fmt.Sprintf("%d", i+1), arg)
	}

	response, err := http.DefaultClient.PostForm(rc.url, values)
	if err != nil {
		return "", err
	}

	buf, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	msg := string(buf)
	if msg[:2] != "OK" {
		return "", os.NewError(msg)
	}

	return msg[3:], nil
}

func (rc *rcClient) Start() (string, os.Error) {
	id, err := rc.do("getNewBrowserSession")
	rc.sessionId = id
	return id, err
}

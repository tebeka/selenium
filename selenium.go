/* Selenium client */
package selenium

import (
	"os"
)

const (
    ById = "id"
    ByXPATH = "xpath"
    ByLinkText = "link text"
    ByPartialLinkText = "partial link text"
    ByName = "name"
    ByTagName = "tag name"
    ByClassName = "class name"
    ByCSSSelector = "css selector"
)

type Capabilities map[string]interface{}
type BrowserProfile *map[string]string

type Build struct {
	Version, Revision, Time string
}

type OS struct {
	Arch, Name, Version string
}

type Status struct {
	Build
	OS
}

type Cookie struct {
	Name string `json:"name"`
	Value string `json:"value"`
	Path string `json:"path"`
	Domain string `json:"domain"`
	Secure bool `json:"secure"`
	Expiry uint `json:"expiry"`
}

type WebDriver interface {
	Status() (*Status, os.Error)
	NewSession() (string, os.Error)
	Quit() os.Error
	CurrentWindowHandle() (string, os.Error)
	WindowHandles() ([]string, os.Error)
	Get(url string) os.Error
	CurrentURL() (string, os.Error)
	Forward() os.Error
	Back() os.Error
	Refresh() os.Error
	Title() (string, os.Error)
	PageSource() (string, os.Error)
	Close() os.Error
	FindElement(by, value string) (WebElement, os.Error)
	FindElements(by, value string) ([]WebElement, os.Error)
	ActiveElement() (WebElement, os.Error)
	SwitchWindow(name string) os.Error
	SwitchFrame(frame string) os.Error
	GetCookies() ([]Cookie, os.Error)
	AddCookie(cookie *Cookie) os.Error
	DeleteAllCookies() os.Error
	DeleteCookie(name string) os.Error
}

type WebElement interface {
	Click() os.Error
	SendKeys(keys string) os.Error
	TagName() (string, os.Error)
	Text() (string, os.Error)
	Submit() os.Error
	Clear() os.Error
//	GetAttribute() (interface{}, os.Error)

}

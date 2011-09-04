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
	FindElement(by, value string) (WebElement, os.Error)
	FindElements(by, value string) ([]WebElement, os.Error)
}

type WebElement interface {
	Click() os.Error
}

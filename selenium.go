/* Selenium client */
package selenium

import (
	"os"
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
}


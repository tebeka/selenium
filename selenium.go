/* Selenium client.

Version: 0.1.1
*/
package selenium

import (
	"os"
)

/* Element finding options */
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

/* Mouse buttons */
const (
	LeftButton = iota
	MiddleButton
	RightButton
)

/* FIXME: Add rest of 
http://code.google.com/p/selenium/wiki/JsonWireProtocol#/session/:sessionId/element/:id/value
*/
const (
	ShiftKey = string('\ue008')
	ControlKey = string('\ue009')
	AltKey = string('\ue00a')
	MetaKey = string('\ue03d')
)

/* Browser capabilities, see
  http://code.google.com/p/selenium/wiki/JsonWireProtocol#Capabilities_JSON_Object
*/
type Capabilities map[string]interface{}

/* Build object, part of Status return. */
type Build struct {
	Version, Revision, Time string
}
/* OS object, part of Status return. */
type OS struct {
	Arch, Name, Version string
}

/* Information retured by Status method. */
type Status struct {
	Build
	OS
}

/* Point */
type Point struct {
	X, Y int
}

/* Size */
type Size struct {
	Width, Height int
}

/* Cookie */
type Cookie struct {
	Name string `json:"name"`
	Value string `json:"value"`
	Path string `json:"path"`
	Domain string `json:"domain"`
	Secure bool `json:"secure"`
	Expiry uint `json:"expiry"`
}

type WebDriver interface {
	/* Status (info) on server */
	Status() (*Status, os.Error)

	/* Start a new session, return session id */
	NewSession() (string, os.Error)
	/* Quit (end) current session */
	Quit() os.Error

	// Page information and manipulation
	/* Return id of current window handle. */
	CurrentWindowHandle() (string, os.Error)
	/* Return ids of current open windows. */
	WindowHandles() ([]string, os.Error)
	/* Current url. */
	CurrentURL() (string, os.Error)
	/* Page title. */
	Title() (string, os.Error)
	/* Get page source. */
	PageSource() (string, os.Error)
	/* Close current window. */
	Close() os.Error
	/* Swtich to window. */
	SwitchWindow(name string) os.Error
	/* Switch to frame, frame parameter can be name or id. */
	SwitchFrame(frame string) os.Error

	// Navigation
	/* Open url. */
	Get(url string) os.Error
	/* Move forward in history. */
	Forward() os.Error
	/* Move backward in history. */
	Back() os.Error
	/* Refresh page. */
	Refresh() os.Error

	// Finding element(s)
	/* Find, return one element. */
	FindElement(by, value string) (WebElement, os.Error)
	/* Find, return list of elements. */
	FindElements(by, value string) ([]WebElement, os.Error)
	/* Current active element. */
	ActiveElement() (WebElement, os.Error)

	// Cookies
	/* Get all cookies */
	GetCookies() ([]Cookie, os.Error)
	/* Add a cookies */
	AddCookie(cookie *Cookie) os.Error
	/* Delete all cookies */
	DeleteAllCookies() os.Error
	/* Delete a cookie */
	DeleteCookie(name string) os.Error

	// Mouse
	/* Click mouse button, button should be on of RightButton, MiddleButton or
	  LeftButton.
	*/
	Click(button int) os.Error
	/* Dobule click */
	DoubleClick() os.Error
	/* Mouse button down */
	ButtonDown() os.Error
	/* Mouse button up */
	ButtonUp() os.Error

	// Misc
	/* Send modifier key to active element.
		modifier can be one of ShiftKey, ControlKey, AltKey, MetaKey.
	*/
	SendModifier(modifier string, isDown bool) os.Error
	Screenshot() ([]byte, os.Error)

	// Alerts
	/* Dismiss current alert. */
	DismissAlert() os.Error
	/* Accept current alert. */
	AcceptAlert() os.Error
	/* Current alert text. */
	AlertText() (string, os.Error)
	/* Set current alert text. */
	SetAlertText(text string) os.Error

	// Scripts
	/* Execute a script. */
	ExecuteScript(script string, args []interface{}) (interface{}, os.Error)
	/* Execute a script async. */
	ExecuteScriptAsync(script string, args []interface{}) (interface{}, os.Error)
}

type WebElement interface {
	// Manipulation

	/* Click on element */
	Click() os.Error
	/* Send keys (type) into element */
	SendKeys(keys string) os.Error
	/* Submit */
	Submit() os.Error
	/* Clear */
	Clear() os.Error
	/* Move mouse to relative coordinates */
	MoveTo(xOffset, yOffset int) os.Error

	// Finding

	/* Find children, return one element. */
	FindElement(by, value string) (WebElement, os.Error)
	/* Find children, return list of elements. */
	FindElements(by, value string) ([]WebElement, os.Error)

	// Porperties

	/* Element name */
	TagName() (string, os.Error)
	/* Text of element */
	Text() (string, os.Error)
	/* Check if element is selected. */
	IsSelected() (bool, os.Error)
	/* Check if element is enabled. */
	IsEnabled() (bool, os.Error)
	/* Check if element is displayed. */
	IsDiaplayed() (bool, os.Error)
	/* Get element attribute. */
	GetAttribute(name string) (string, os.Error)
	/* Element location. */
	Location() (*Point, os.Error)
	/* Element location once it has been scrolled into view. */
	LocationInView() (*Point, os.Error)
	/* Element size */
	Size() (*Size, os.Error)
	/* Get element CSS property value. */
	CSSProperty(name string) (string, os.Error)
}

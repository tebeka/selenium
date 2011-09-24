/* Selenium client.

Currently provides on WebDriver remote client.

Version: 0.2.0
*/
package selenium

import (
	"os"
)

/* Element finding options */
const (
	ById              = "id"
	ByXPATH           = "xpath"
	ByLinkText        = "link text"
	ByPartialLinkText = "partial link text"
	ByName            = "name"
	ByTagName         = "tag name"
	ByClassName       = "class name"
	ByCSSSelector     = "css selector"
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
	ShiftKey   = string('\ue008')
	ControlKey = string('\ue009')
	AltKey     = string('\ue00a')
	MetaKey    = string('\ue03d')
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
	Name   string `json:"name"`
	Value  string `json:"value"`
	Path   string `json:"path"`
	Domain string `json:"domain"`
	Secure bool   `json:"secure"`
	Expiry uint   `json:"expiry"`
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

type SeleniumRC interface {
	// Start new session
	Start() (string, os.Error)
	// Stop (close) current session
	Stop() os.Error
	// Clicks on a link, button, checkbox or radio button
	Click(locator string) os.Error
	// Double clicks on a link, button, checkbox or radio button
	DoubleClick(locator string) os.Error
	// Simulates opening the context menu for the specified element
	ContextMenu(locator string) os.Error
	// Clicks on a link, button, checkbox or radio button at specific location
	ClickAt(locator string, x, y int) os.Error
	// Double clicks on a link, button, checkbox or radio button at specific location
	DoubleClickAt(locator string, x, y int) os.Error
	// Simulates opening the context menu for the specified element at specific location
	ContextMenuAt(locator string, x, y int) os.Error
	// Explicitly simulate an event, to trigger the corresponding "on\ *event*" handler
	FireEvent(locator, event string) os.Error
	// Move the focus to the specified element
	Focus(locator string) os.Error
	// Simulates a user pressing and releasing a key.
	KeyPress(locator, keys string) os.Error
	// Simulates a user pressing a key (without releasing it yet)
	KeyDown(locator, keys string) os.Error
	// Simulates a user releasing a key
	KeyUp(locator, keys string) os.Error
	// Simulates a user hovering a mouse over the specified element
	MouseOver(locator string) os.Error
    // Simulates a user moving the mouse pointer away from the specified element
	MouseOut(locator string) os.Error
    // Simulates a user pressing the left mouse button
	MouseDown(locator string) os.Error
    // Simulates a user pressing the right mouse button
	MouseDownRight(locator string) os.Error
	// Simulates the event that occurs when the user releases the mouse button
	MouseUp(locator string) os.Error
	// Sets the value of an input field, as though you typed it in
	Type(locator, value string) os.Error
	// Simulates keystroke events on the specified element, as typed key-by-key
	TypeKeys(locator, value string) os.Error
	// Check a toggle-button (checkbox/radio)
	Check(locator string) os.Error
	// Uncheck a toggle-button (checkbox/radio)
	Uncheck(locator string) os.Error
	// Select an option from a drop-down using an option locator
	Select(location, option string) os.Error
	// Submit the specified form
	Submit(form string) os.Error
	// Opens an URL in the test frame
	Open(url string, ignoreResponse bool) os.Error
	// Simulates the user clicking the "back" button on their browser
	Back() os.Error
	// Simulates the user clicking the "Refresh" button on their browser
	Refresh() os.Error
	// Simulates the user clicking the "close" button in the titlebar of a popup window or tab
	Close() os.Error
	// Gets the absolute URL of the current page
	Location() (string, os.Error)
    // Gets the title of the current page
	Title() (string, os.Error)
	// Gets the entire text of the page
	Body() (string, os.Error)
	// Gets the (whitespace-trimmed) value of an input field
	Value(locator string) (string, os.Error)
	// Gets the text of an element
	Text(locator string) (string, os.Error)
	// Briefly changes the backgroundColor of the specified element yellow (useful for debugging)
	Highlight(locator string) os.Error
    // Gets whether a toggle-button (checkbox/radio) is checked
	IsChecked(location string) (bool, os.Error)
    // Gets the value of an element attribute
	Attribute(locator string) (string, os.Error)
	// Verifies that the specified element is somewhere on the page
	IsPresent(locator string) (bool, os.Error)
	// Determines if the specified element is visible
	IsVisible(locator string) (bool, os.Error)

}


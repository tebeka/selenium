/* Selenium client.

Currently provides on WebDriver remote client.

Version: 0.2.2
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

/* Keys */
const (
	NullKey       = string('\ue000')
	CancelKey     = string('\ue001')
	HelpKey       = string('\ue002')
	BackspaceKey  = string('\ue003')
	TabKey        = string('\ue004')
	ClearKey      = string('\ue005')
	ReturnKey     = string('\ue006')
	EnterKey      = string('\ue007')
	ShiftKey      = string('\ue008')
	ControlKey    = string('\ue009')
	AltKey        = string('\ue00a')
	PauseKey      = string('\ue00b')
	EscapeKey     = string('\ue00c')
	SpaceKey      = string('\ue00d')
	PageUpKey     = string('\ue00e')
	PageDownKey   = string('\ue00f')
	EndKey        = string('\ue010')
	HomeKey       = string('\ue011')
	LeftArrowKey  = string('\ue012')
	UpArrowKey    = string('\ue013')
	RightArrowKey = string('\ue014')
	DownArrowKey  = string('\ue015')
	InsertKey     = string('\ue016')
	DeleteKey     = string('\ue017')
	SemicolonKey  = string('\ue018')
	EqualsKey     = string('\ue019')
	Numpad0Key    = string('\ue01a')
	Numpad1Key    = string('\ue01b')
	Numpad2Key    = string('\ue01c')
	Numpad3Key    = string('\ue01d')
	Numpad4Key    = string('\ue01e')
	Numpad5Key    = string('\ue01f')
	Numpad6Key    = string('\ue020')
	Numpad7Key    = string('\ue021')
	Numpad8Key    = string('\ue022')
	Numpad9Key    = string('\ue023')
	MultiplyKey   = string('\ue024')
	AddKey        = string('\ue025')
	SeparatorKey  = string('\ue026')
	SubstractKey  = string('\ue027')
	DecimalKey    = string('\ue028')
	DivideKey     = string('\ue029')
	F1Key         = string('\ue031')
	F2Key         = string('\ue032')
	F3Key         = string('\ue033')
	F4Key         = string('\ue034')
	F5Key         = string('\ue035')
	F6Key         = string('\ue036')
	F7Key         = string('\ue037')
	F8Key         = string('\ue038')
	F9Key         = string('\ue039')
	F10Key        = string('\ue03a')
	F11Key        = string('\ue03b')
	F12Key        = string('\ue03c')
	MetaKey       = string('\ue03d')
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

	/* Current session capabilities */
	Capabilities() (Capabilities, os.Error)
	/* Set the amount of time, in milliseconds, that asynchronous scripts are permitted to run before they are aborted. */
	SetAsyncScriptTimeout(ms uint) os.Error
	/* Set the amount of time, in milliseconds, the driver should wait when searching for elements. */
	SetImplicitWaitTimeout(ms uint) os.Error

	// IME
	/* List all available engines on the machine. */
	AvailableEngines() ([]string, os.Error)
	/* Get the name of the active IME engine. */
	ActiveEngine() (string, os.Error)
	/* Indicates whether IME input is active at the moment. */
	IsEngineActivated() (bool, os.Error)
	/* De-activates the currently-active IME engine. */
	DeactivateEngine() os.Error
	/* Make an engines active */
	ActivateEngine(engine string) os.Error

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
	/* Switch to frame, frame parameter can be name or id. */
	SwitchFrame(frame string) os.Error
	/* Swtich to window. */
	SwitchWindow(name string) os.Error
	/* Close window. */
	CloseWindow(name string) os.Error

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

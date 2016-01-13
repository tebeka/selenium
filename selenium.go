package selenium

import (
	"time"
)

const (
	Version = "0.8.5" // Driver version
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

// Proxy  object, see
// https://code.google.com/p/selenium/wiki/JsonWireProtocol#Proxy_JSON_Object
type Proxy map[string]string

/* Build object, part of Status return. */
type Build struct {
	Version, Revision, Time string
}

/* OS object, part of Status return. */
type OS struct {
	Arch, Name, Version string
}

type Java struct {
	Version string
}

/* Information retured by Status method. */
type Status struct {
	Java  Java
	Build Build
	OS    OS
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
	Status() (*Status, error)

	/* Start a new session, return session id */
	NewSession() (string, error)

	/* Current session id (empty string on none) */
	SessionId() string

	/* Current session capabilities */
	Capabilities() (Capabilities, error)
	/* Set the amount of time, in microseconds, that asynchronous scripts are permitted to run before they are aborted.

	Note that Selenium/WebDriver timeouts are in milliseconds, timeout will be rounded to nearest millisecond.
	*/
	SetAsyncScriptTimeout(timeout time.Duration) error
	/* Set the amount of time, in milliseconds, the driver should wait when searching for elements.

	Note that Selenium/WebDriver timeouts are in milliseconds, timeout will be rounded to nearest millisecond.
	*/
	SetImplicitWaitTimeout(timeout time.Duration) error

	// IME
	/* List all available engines on the machine. */
	AvailableEngines() ([]string, error)
	/* Get the name of the active IME engine. */
	ActiveEngine() (string, error)
	/* Indicates whether IME input is active at the moment. */
	IsEngineActivated() (bool, error)
	/* De-activates the currently-active IME engine. */
	DeactivateEngine() error
	/* Make an engines active */
	ActivateEngine(engine string) error

	/* Quit (end) current session */
	Quit() error

	// Page information and manipulation
	/* Return id of current window handle. */
	CurrentWindowHandle() (string, error)
	/* Return ids of current open windows. */
	WindowHandles() ([]string, error)
	/* Current url. */
	CurrentURL() (string, error)
	/* Page title. */
	Title() (string, error)
	/* Get page source. */
	PageSource() (string, error)
	/* Close current window. */
	Close() error
	/* Switch to frame, frame parameter can be name or id. */
	SwitchFrame(frame string) error
	/* Swtich to window. */
	SwitchWindow(name string) error
	/* Close window. */
	CloseWindow(name string) error
	/* Maximize window, if name is empty - will use current */
	MaximizeWindow(name string) error
	/* Resize window, if name is empty - will use current */
	ResizeWindow(name string, width, height int) error

	// Navigation
	/* Open url. */
	Get(url string) error
	/* Move forward in history. */
	Forward() error
	/* Move backward in history. */
	Back() error
	/* Refresh page. */
	Refresh() error

	// Finding element(s)
	/* Find, return one element. */
	FindElement(by, value string) (WebElement, error)
	/* Find, return list of elements. */
	FindElements(by, value string) ([]WebElement, error)
	/* Current active element. */
	ActiveElement() (WebElement, error)

	// Decoding element(s)
	/* Decode a single element response. */
	DecodeElement([]byte) (WebElement, error)
	/* Decode a multi element response. */
	DecodeElements([]byte) ([]WebElement, error)

	// Cookies
	/* Get all cookies */
	GetCookies() ([]Cookie, error)
	/* Add a cookies */
	AddCookie(cookie *Cookie) error
	/* Delete all cookies */
	DeleteAllCookies() error
	/* Delete a cookie */
	DeleteCookie(name string) error

	// Mouse
	/* Click mouse button, button should be on of RightButton, MiddleButton or
	LeftButton.
	*/
	Click(button int) error
	/* Dobule click */
	DoubleClick() error
	/* Mouse button down */
	ButtonDown() error
	/* Mouse button up */
	ButtonUp() error

	// Misc
	/* Send modifier key to active element.
	modifier can be one of ShiftKey, ControlKey, AltKey, MetaKey.
	*/
	SendModifier(modifier string, isDown bool) error
	/* Take a screenshot */
	Screenshot() ([]byte, error)

	// Alerts
	/* Dismiss current alert. */
	DismissAlert() error
	/* Accept current alert. */
	AcceptAlert() error
	/* Current alert text. */
	AlertText() (string, error)
	/* Set current alert text. */
	SetAlertText(text string) error

	// Scripts
	/* Execute a script. */
	ExecuteScript(script string, args []interface{}) (interface{}, error)
	/* Execute a script async. */
	ExecuteScriptAsync(script string, args []interface{}) (interface{}, error)

	/* Execute a script but don't JSON decode. */
	ExecuteScriptRaw(script string, args []interface{}) ([]byte, error)
	/* Execute a script async but don't JSON decode. */
	ExecuteScriptAsyncRaw(script string, args []interface{}) ([]byte, error)
}

type WebElement interface {
	// Manipulation

	/* Click on element */
	Click() error
	/* Send keys (type) into element */
	SendKeys(keys string) error
	/* Submit */
	Submit() error
	/* Clear */
	Clear() error
	/* Move mouse to relative coordinates */
	MoveTo(xOffset, yOffset int) error

	// Finding

	/* Find children, return one element. */
	FindElement(by, value string) (WebElement, error)
	/* Find children, return list of elements. */
	FindElements(by, value string) ([]WebElement, error)

	// Porperties

	/* Element name */
	TagName() (string, error)
	/* Text of element */
	Text() (string, error)
	/* Check if element is selected. */
	IsSelected() (bool, error)
	/* Check if element is enabled. */
	IsEnabled() (bool, error)
	/* Check if element is displayed. */
	IsDisplayed() (bool, error)
	/* Get element attribute. */
	GetAttribute(name string) (string, error)
	/* Element location. */
	Location() (*Point, error)
	/* Element location once it has been scrolled into view. */
	LocationInView() (*Point, error)
	/* Element size */
	Size() (*Size, error)
	/* Get element CSS property value. */
	CSSProperty(name string) (string, error)
}

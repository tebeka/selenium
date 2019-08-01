package selenium

import (
	"time"

	"github.com/tebeka/selenium/chrome"
	"github.com/tebeka/selenium/firefox"
	"github.com/tebeka/selenium/log"
)

// TODO(minusnine): make an enum type called FindMethod.

// Methods by which to find elements.
const (
	ByID              = "id"
	ByXPATH           = "xpath"
	ByLinkText        = "link text"
	ByPartialLinkText = "partial link text"
	ByName            = "name"
	ByTagName         = "tag name"
	ByClassName       = "class name"
	ByCSSSelector     = "css selector"
)

// Mouse buttons.
const (
	LeftButton = iota
	MiddleButton
	RightButton
)

// Special keyboard keys, for SendKeys.
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

// Capabilities configures both the WebDriver process and the target browsers,
// with standard and browser-specific options.
type Capabilities map[string]interface{}

// AddChrome adds Chrome-specific capabilities.
func (c Capabilities) AddChrome(f chrome.Capabilities) {
	c[chrome.CapabilitiesKey] = f
	c[chrome.DeprecatedCapabilitiesKey] = f
}

// AddFirefox adds Firefox-specific capabilities.
func (c Capabilities) AddFirefox(f firefox.Capabilities) {
	c[firefox.CapabilitiesKey] = f
}

// AddProxy adds proxy configuration to the capabilities.
func (c Capabilities) AddProxy(p Proxy) {
	c["proxy"] = p
}

// AddLogging adds logging configuration to the capabilities.
func (c Capabilities) AddLogging(l log.Capabilities) {
	c[log.CapabilitiesKey] = l
}

// SetLogLevel sets the logging level of a component. It is a shortcut for
// passing a log.Capabilities instance to AddLogging.
func (c Capabilities) SetLogLevel(typ log.Type, level log.Level) {
	if _, ok := c[log.CapabilitiesKey]; !ok {
		c[log.CapabilitiesKey] = make(log.Capabilities)
	}
	m := c[log.CapabilitiesKey].(log.Capabilities)
	m[typ] = level
}

// Proxy specifies configuration for proxies in the browser. Set the key
// "proxy" in Capabilities to an instance of this type.
type Proxy struct {
	// Type is the type of proxy to use. This is required to be populated.
	Type ProxyType `json:"proxyType"`

	// AutoconfigURL is the URL to be used for proxy auto configuration. This is
	// required if Type is set to PAC.
	AutoconfigURL string `json:"proxyAutoconfigUrl,omitempty"`

	// The following are used when Type is set to Manual.
	//
	// Note that in Firefox, connections to localhost are not proxied by default,
	// even if a proxy is set. This can be overridden via a preference setting.
	FTP           string   `json:"ftpProxy,omitempty"`
	HTTP          string   `json:"httpProxy,omitempty"`
	SSL           string   `json:"sslProxy,omitempty"`
	SOCKS         string   `json:"socksProxy,omitempty"`
	SOCKSVersion  int      `json:"socksVersion,omitempty"`
	SOCKSUsername string   `json:"socksUsername,omitempty"`
	SOCKSPassword string   `json:"socksPassword,omitempty"`
	NoProxy       []string `json:"noProxy,omitempty"`

	// The W3C draft spec includes port fields as well. According to the
	// specification, ports can also be included in the above addresses. However,
	// in the Geckodriver implementation, the ports must be specified by these
	// additional fields.
	HTTPPort  int `json:"httpProxyPort,omitempty"`
	SSLPort   int `json:"sslProxyPort,omitempty"`
	SocksPort int `json:"socksProxyPort,omitempty"`
}

// ProxyType is an enumeration of the types of proxies available.
type ProxyType string

const (
	// Direct connection - no proxy in use.
	Direct ProxyType = "direct"
	// Manual proxy settings configured, e.g. setting a proxy for HTTP, a proxy
	// for FTP, etc.
	Manual = "manual"
	// Autodetect proxy, probably with WPAD
	Autodetect = "autodetect"
	// System settings used.
	System = "system"
	// PAC - Proxy autoconfiguration from a URL.
	PAC = "pac"
)

// Status contains information returned by the Status method.
type Status struct {
	// The following fields are used by Selenium and ChromeDriver.
	Java struct {
		Version string
	}
	Build struct {
		Version, Revision, Time string
	}
	OS struct {
		Arch, Name, Version string
	}

	// The following fields are specified by the W3C WebDriver specification and
	// are used by GeckoDriver.
	Ready   bool
	Message string
}

// Point is a 2D point.
type Point struct {
	X, Y int
}

// Size is a size of HTML element.
type Size struct {
	Width, Height int
}

// Cookie represents an HTTP cookie.
type Cookie struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Path   string `json:"path"`
	Domain string `json:"domain"`
	Secure bool   `json:"secure"`
	Expiry uint   `json:"expiry"`
}

// WebDriver defines methods supported by WebDriver drivers.
type WebDriver interface {
	// Status returns various pieces of information about the server environment.
	Status() (*Status, error)

	// NewSession starts a new session and returns the session ID.
	NewSession() (string, error)

	// SessionId returns the current session ID
	//
	// Deprecated: This identifier is not Go-style correct. Use SessionID
	// instead.
	SessionId() string

	// SessionID returns the current session ID.
	SessionID() string

	// SwitchSession switches to the given session ID.
	SwitchSession(sessionID string) error

	// Capabilities returns the current session's capabilities.
	Capabilities() (Capabilities, error)

	// SetAsyncScriptTimeout sets the amount of time that asynchronous scripts
	// are permitted to run before they are aborted. The timeout will be rounded
	// to nearest millisecond.
	SetAsyncScriptTimeout(timeout time.Duration) error
	// SetImplicitWaitTimeout sets the amount of time the driver should wait when
	// searching for elements. The timeout will be rounded to nearest millisecond.
	SetImplicitWaitTimeout(timeout time.Duration) error
	// SetPageLoadTimeout sets the amount of time the driver should wait when
	// loading a page. The timeout will be rounded to nearest millisecond.
	SetPageLoadTimeout(timeout time.Duration) error

	// Quit ends the current session. The browser instance will be closed.
	Quit() error

	// CurrentWindowHandle returns the ID of current window handle.
	CurrentWindowHandle() (string, error)
	// WindowHandles returns the IDs of current open windows.
	WindowHandles() ([]string, error)
	// CurrentURL returns the browser's current URL.
	CurrentURL() (string, error)
	// Title returns the current page's title.
	Title() (string, error)
	// PageSource returns the current page's source.
	PageSource() (string, error)
	// Close closes the current window.
	Close() error
	// SwitchFrame switches to the given frame. The frame parameter can be the
	// frame's ID as a string, its WebElement instance as returned by
	// GetElement, or nil to switch to the current top-level browsing context.
	SwitchFrame(frame interface{}) error
	// SwitchWindow switches the context to the specified window.
	SwitchWindow(name string) error
	// CloseWindow closes the specified window.
	CloseWindow(name string) error
	// MaximizeWindow maximizes a window. If the name is empty, the current
	// window will be maximized.
	MaximizeWindow(name string) error
	// ResizeWindow changes the dimensions of a window. If the name is empty, the
	// current window will be maximized.
	ResizeWindow(name string, width, height int) error

	// Get navigates the browser to the provided URL.
	Get(url string) error
	// Forward moves forward in history.
	Forward() error
	// Back moves backward in history.
	Back() error
	// Refresh refreshes the page.
	Refresh() error

	// FindElement finds exactly one element in the current page's DOM.
	FindElement(by, value string) (WebElement, error)
	// FindElement finds potentially many elements in the current page's DOM.
	FindElements(by, value string) ([]WebElement, error)
	// ActiveElement returns the currently active element on the page.
	ActiveElement() (WebElement, error)

	// DecodeElement decodes a single element response.
	DecodeElement([]byte) (WebElement, error)
	// DecodeElements decodes a multi-element response.
	DecodeElements([]byte) ([]WebElement, error)

	// GetCookies returns all of the cookies in the browser's jar.
	GetCookies() ([]Cookie, error)
	// GetCookie returns the named cookie in the jar, if present. This method is
	// only implemented for Firefox.
	GetCookie(name string) (Cookie, error)
	// AddCookie adds a cookie to the browser's jar.
	AddCookie(cookie *Cookie) error
	// DeleteAllCookies deletes all of the cookies in the browser's jar.
	DeleteAllCookies() error
	// DeleteCookie deletes a cookie to the browser's jar.
	DeleteCookie(name string) error

	// Click clicks a mouse button. The button should be one of RightButton,
	// MiddleButton or LeftButton.
	Click(button int) error
	// DoubleClick clicks the left mouse button twice.
	DoubleClick() error
	// ButtonDown causes the left mouse button to be held down.
	ButtonDown() error
	// ButtonUp causes the left mouse button to be released.
	ButtonUp() error

	// SendModifier sends the modifier key to the active element. The modifier
	// can be one of ShiftKey, ControlKey, AltKey, MetaKey.
	//
	// Deprecated: Use KeyDown or KeyUp instead.
	SendModifier(modifier string, isDown bool) error
	// KeyDown sends a sequence of keystrokes to the active element. This method
	// is similar to SendKeys but without the implicit termination. Modifiers are
	// not released at the end of each call.
	KeyDown(keys string) error
	// KeyUp indicates that a previous keystroke sent by KeyDown should be
	// released.
	KeyUp(keys string) error
	// Screenshot takes a screenshot of the browser window.
	Screenshot() ([]byte, error)
	// Log fetches the logs. Log types must be previously configured in the
	// capabilities.
	//
	// NOTE: will return an error (not implemented) on IE11 or Edge drivers.
	Log(typ log.Type) ([]log.Message, error)

	// DismissAlert dismisses current alert.
	DismissAlert() error
	// AcceptAlert accepts the current alert.
	AcceptAlert() error
	// AlertText returns the current alert text.
	AlertText() (string, error)
	// SetAlertText sets the current alert text.
	SetAlertText(text string) error

	// ExecuteScript executes a script.
	ExecuteScript(script string, args []interface{}) (interface{}, error)
	// ExecuteScriptAsync asynchronously executes a script.
	ExecuteScriptAsync(script string, args []interface{}) (interface{}, error)

	// ExecuteScriptRaw executes a script but does not perform JSON decoding.
	ExecuteScriptRaw(script string, args []interface{}) ([]byte, error)
	// ExecuteScriptAsyncRaw asynchronously executes a script but does not
	// perform JSON decoding.
	ExecuteScriptAsyncRaw(script string, args []interface{}) ([]byte, error)

	// WaitWithTimeoutAndInterval waits for the condition to evaluate to true.
	WaitWithTimeoutAndInterval(condition Condition, timeout, interval time.Duration) error

	// WaitWithTimeout works like WaitWithTimeoutAndInterval, but with default polling interval.
	WaitWithTimeout(condition Condition, timeout time.Duration) error

	//Wait works like WaitWithTimeoutAndInterval, but using the default timeout and polling interval.
	Wait(condition Condition) error
}

// WebElement defines method supported by web elements.
type WebElement interface {
	// Click clicks on the element.
	Click() error
	// SendKeys types into the element.
	SendKeys(keys string) error
	// Submit submits the button.
	Submit() error
	// Clear clears the element.
	Clear() error
	// MoveTo moves the mouse to relative coordinates from center of element, If
	// the element is not visible, it will be scrolled into view.
	MoveTo(xOffset, yOffset int) error

	// FindElement finds a child element.
	FindElement(by, value string) (WebElement, error)
	// FindElement finds multiple children elements.
	FindElements(by, value string) ([]WebElement, error)

	// TagName returns the element's name.
	TagName() (string, error)
	// Text returns the text of the element.
	Text() (string, error)
	// IsSelected returns true if element is selected.
	IsSelected() (bool, error)
	// IsEnabled returns true if the element is enabled.
	IsEnabled() (bool, error)
	// IsDisplayed returns true if the element is displayed.
	IsDisplayed() (bool, error)
	// GetAttribute returns the named attribute of the element.
	GetAttribute(name string) (string, error)
	// Location returns the element's location.
	Location() (*Point, error)
	// LocationInView returns the element's location once it has been scrolled
	// into view.
	LocationInView() (*Point, error)
	// Size returns the element's size.
	Size() (*Size, error)
	// CSSProperty returns the value of the specified CSS property of the
	// element.
	CSSProperty(name string) (string, error)
	// Screenshot takes a screenshot of the attribute scroll'ing if necessary.
	Screenshot(scroll bool) ([]byte, error)
}

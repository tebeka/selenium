// Package sauce interacts with the Sauce Labs hosted browser testing environment.
package sauce

import (
	"encoding/json"
	"fmt"
)

// Addr returns the URL to use for driving a remote web browser.
func Addr(userName, accessKey string) string {
	return fmt.Sprintf("http://%s:%s@ondemand.saucelabs.com/wd/hub", userName, accessKey)
}

// Capabilities are the options to provide to the Sauce infrastructure for each
// test.
//
// See the following URL for more details of each configuration parameter:
// https://wiki.saucelabs.com/display/DOCS/Test+Configuration+Options
type Capabilities struct {
	// The name of the browser test against.
	Browser string `json:"browser,omitempty"`
	// The version of the browser you want to use in your test.
	Version string `json:"version,omitempty"`
	// Which operating system the browser should be running on.
	Platform string `json:"platform,omitempty"`
	// The version of Selenium to use.
	SeleniumVersion string `json:"seleniumVersion,omitempty"`
	// When testing Chrome, the version of ChromeDriver to use.
	ChromeDriverVersion string `json:"chromedriverVersion,omitempty"`
	// When testing IE, the version of IE Driver to use.
	IEDriverVersion string `json:"iedriverVersion,omitempty"`

	// Setting this option will automatically accept any unexpected browser
	// alerts that come up during your test.
	AutoAcceptAlerts *bool `json:"autoAcceptAlerts,omitempty"`

	// Used to record test names for jobs.
	TestName string `json:"name,omitempty"`
	// Used to associate jobs with a build number or app version.
	BuildNumber string `json:"build,omitempty"`
	// User-defined tags for grouping and filtering jobs.
	Tags []string `json:"tags,omitempty"`
	// User-defined custom data, limited to 64KB in size.
	CustomData json.RawMessage `json:"customData,omitempty"`

	// The maximum test duration to allow, in seconds. By default, this is 30
	// minutes. The maximum value is 10800 seconds (three hours).
	MaximumDuration int `json:"maxDuration,omitempty"`
	// The maximum amount of time a command can run in a browser, in seconds. By
	// default, this is 300 seconds (five minutes). The maximum value is 600
	// seconds (ten minutes).
	CommandTimeout int `json:"commandTimeout,omitempty"`
	// The maxmimum amount of time to wait for a new command. By default, this is
	// 90 seconds. The maximum value is 1000 seconds.
	IdleTimeout int `json:"idleTimeout,omitempty"`

	// Run an executable before the test.
	PreRun *PreRun `json:"prerun,omitempty"`

	// The screen resolution should be used during the test session.
	ScreenResolution string `json:"screenResolution,omitempty"`
	// The timezone to configure on Desktop Test VMs.
	TimeZone string `json:"timeZone,omitempty"`

	// Disable use of the Selenium HTTP proxy server.
	AvoidProxy bool `json:"avoidProxy,omitempty"`

	// The visibility of the job.
	Visibility Visibility `json:"public,omitempty"`

	// By default, Sauce records a video of every test run. Set this to false to
	// disable recording video.
	RecordVideo *bool `json:"recordVideo,omitempty"`
	// Set to false to discard videos for passing tests identified using the
	// passed  setting. This disables video post-processing and uploading that
	// may otherwise consume some extra time.
	UploadVideoOnPass *bool `json:"videoUploadOnPass,omitempty"`
	// Set to false to prevent recording of screenshots.
	RecordScreenshots *bool `json:"recordScreenshots,omitempty"`
	// Set to false to disable log recording.
	RecordLogs *bool `json:"recordLogs,omitempty"`
	// Set to false to disable capturing the HTML source at each step.
	CaptureHTML *bool `json:"captureHtml,omitempty"`

	// The priority level of the job. Used for determining which job to start
	// across a collection of sub-accounts. Smaller numbers indicate higher
	// priority.
	Priority int `json:"priority,omitempty"`

	// Selenium WebDriver captures automatic screenshots for every server side
	// failure, for example if an element is not found. Sauce disables this by
	// default to reduce network traffic during tests, resulting in a
	// considerable performance improvement in most tests. Set this to true to
	// reenable this feature.
	WebDriverScreenshot *bool `json:"webdriverRemoteQuietExceptions,omitempty"`
}

// Visibility is a visibility level for a test.
type Visibility string

const (
	// Public is the visibility to specify that the result is accessible to everyone.
	Public Visibility = "public"
	// PublicRestricted is the visibility to specify that anonymous users have
	// access to the result page and video, but not the logs.
	PublicRestricted Visibility = "public restricted"
	// Team is the visibility to specify that the results are only accessible to
	// people under the same root account as the executor's.
	Team Visibility = "team"
	// Private is the visibility to specify that only the owner of the test will
	// be able to view assets and test result page.
	Private Visibility = "private"
)

// PreRun configures a URL to an executable file, which will be downloaded and
// executed to configure the VM before the test starts.
type PreRun struct {
	// The URL to the executable you want to run before your browser session
	// starts.
	Executable string `json:"executable,omitempty"`
	// A list of the command line parameters that you want the executable to
	// receive.
	Args []string `json:"args,omitempty"`
	// Whether Sauce should wait for this executable to finish before your
	// browser session starts.
	Background bool `json:"background,omitempty"`
	// The number of seconds Sauce will wait for your executable to finish before
	// your browser session starts.
	Timeout int `json:"timeout,omitempty"`
}

// ToMap returns the capabilities in a key/value structure.
func (c *Capabilities) ToMap() (map[string]interface{}, error) {
	buf, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	m := make(map[string]interface{})
	if err := json.Unmarshal(buf, &m); err != nil {
		return nil, err
	}
	return m, nil
}

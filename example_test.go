package selenium_test

import (
	"encoding/json"
	"fmt"
	"github.com/chromedp/cdproto"
	"github.com/chromedp/cdproto/network"
	"github.com/mitchellh/mapstructure"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"github.com/tebeka/selenium/log"
	"os"
	"strings"
	"time"
)

// This example shows how to navigate to a http://play.golang.org page, input a
// short program, run it, and inspect its output.
//
// If you want to actually run this example:
//
//   1. Ensure the file paths at the top of the function are correct.
//   2. Remove the word "Example" from the comment at the bottom of the
//      function.
//   3. Run:
//      go test -test.run=Example$ github.com/tebeka/selenium
func Example() {
	// Start a Selenium WebDriver server instance (if one is not already
	// running).
	const (
		// These paths will be different on your system.
		seleniumPath    = "vendor/selenium-server-standalone-3.4.jar"
		geckoDriverPath = "vendor/geckodriver-v0.18.0-linux64"
		port            = 8080
	)
	opts := []selenium.ServiceOption{
		selenium.StartFrameBuffer(),           // Start an X frame buffer for the browser to run in.
		selenium.GeckoDriver(geckoDriverPath), // Specify the path to GeckoDriver in order to use Firefox.
		selenium.Output(os.Stderr),            // Output debug information to STDERR.
	}
	selenium.SetDebug(true)
	service, err := selenium.NewSeleniumService(seleniumPath, port, opts...)
	if err != nil {
		panic(err) // panic is used only as an example and is not otherwise recommended.
	}
	defer service.Stop()

	// Connect to the WebDriver instance running locally.
	caps := selenium.Capabilities{"browserName": "firefox"}
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		panic(err)
	}
	defer wd.Quit()

	// Navigate to the simple playground interface.
	if err := wd.Get("http://play.golang.org/?simple=1"); err != nil {
		panic(err)
	}

	// Get a reference to the text box containing code.
	elem, err := wd.FindElement(selenium.ByCSSSelector, "#code")
	if err != nil {
		panic(err)
	}
	// Remove the boilerplate code already in the text box.
	if err := elem.Clear(); err != nil {
		panic(err)
	}

	// Enter some new code in text box.
	err = elem.SendKeys(`
		package main
		import "fmt"

		func main() {
			fmt.Println("Hello WebDriver!")
		}
	`)
	if err != nil {
		panic(err)
	}

	// Click the run button.
	btn, err := wd.FindElement(selenium.ByCSSSelector, "#run")
	if err != nil {
		panic(err)
	}
	if err := btn.Click(); err != nil {
		panic(err)
	}

	// Wait for the program to finish running and get the output.
	outputDiv, err := wd.FindElement(selenium.ByCSSSelector, "#output")
	if err != nil {
		panic(err)
	}

	var output string
	for {
		output, err = outputDiv.Text()
		if err != nil {
			panic(err)
		}
		if output != "Waiting for remote server..." {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}

	fmt.Printf("%s", strings.Replace(output, "\n\n", "\n", -1))
	// Example Output:
	// Hello WebDriver!
	//
	// Program exited.

	// The following shows an example of using the Actions API.
	// Please refer to the WC3 Actions spec for more detailed information.
	if err := wd.Get("http://play.golang.org/?simple=1"); err != nil {
		panic(err)
	}

	// Create a point which will be used as an offset to click on the
	// code editor text box element on the page.
	offset := selenium.Point{X: 100, Y: 100}

	// Call StorePointerActions to store a number of Pointer actions which
	// will be executed sequentially.
	// "mouse1" is used as a unique virtual device identifier for this
	// and future actions.
	// selenium.MousePointer is used to identify the type of the pointer.
	// The stored action chain will move the pointer and click on the code
	// editor text box on the page.
	selenium.StorePointerActions("mouse1",
		selenium.MousePointer,
		// using selenium.FromViewport as the move origin
		// which calculates the offset from 0,0.
		// the other valid option is selenium.FromPointer.
		selenium.PointerMoveAction(0, offset, selenium.FromViewport),
		selenium.PointerPauseAction(250),
		selenium.PointerDownAction(selenium.LeftButton),
		selenium.PointerPauseAction(250),
		selenium.PointerUpAction(selenium.LeftButton),
	)

	// Call StoreKeyActions to store a number of Key actions which
	// will be executed sequentially.
	// "keyboard1" is used as a unique virtual device identifier
	// for this and future actions.
	// The stored action chain will send keyboard inputs to the browser.
	selenium.StoreKeyActions("keyboard1",
		selenium.KeyDownAction(selenium.ControlKey),
		selenium.KeyPauseAction(50),
		selenium.KeyDownAction("a"),
		selenium.KeyPauseAction(50),
		selenium.KeyUpAction("a"),
		selenium.KeyUpAction(selenium.ControlKey),
		selenium.KeyDownAction("h"),
		selenium.KeyDownAction("e"),
		selenium.KeyDownAction("l"),
		selenium.KeyDownAction("l"),
		selenium.KeyDownAction("o"),
	)

	// Call PerformActions to execute stored action - based on
	// the order of the previous calls, PointerActions will be
	// executed first and then KeyActions.
	if err := wd.PerformActions(); err != nil {
		panic(err)
	}

	// Call ReleaseActions to release any PointerDown or
	// KeyDown Actions that haven't been released through an Action.
	if err := wd.ReleaseActions(); err != nil {
		panic(err)
	}

}

func ExampleChromeCDP() {
	service, err := selenium.NewChromeDriverService("/usr/local/bin/chromedriver", 6789)
	if err != nil {
		panic(err)
	}
	defer service.Stop()

	caps := selenium.Capabilities{"browserName": "chrome"}
	caps.AddLogging(log.Capabilities{
		log.Performance: log.All,
	})
	caps.AddChrome(chrome.Capabilities{
		Args: []string{
			"--user-data-dir=/tmp/chrome",
			"--headless",
		},
	})
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", 6789))
	if err != nil {
		panic(err)
	}
	defer wd.Quit()

	if err := wd.Get("xxxxx"); err != nil {
		panic(err)
	}

	// /resource/RelatedPinFeedResource/get

	log_messages, err := wd.Log(log.Performance)
	if err != nil {
		panic(err)
	}
	for _, log_message := range log_messages {
		message := make(map[string]interface{})
		json.Unmarshal([]byte(log_message.Message), &message)
		message = message["message"].(map[string]interface{})
		method := message["method"]
		if method != "Network.responseReceived" {
			continue
		}

		params := message["params"].(map[string]interface{})
		response := params["response"].(map[string]interface{})
		if strings.Contains(response["url"].(string), "xxxxx") == false {
			continue
		}

		// use cdp
		request_id := network.RequestID(params["requestId"].(string))
		response_body, err := wd.ExecuteChromeDPCommand(cdproto.CommandNetworkGetResponseBody, network.GetResponseBody(request_id))
		if err != nil {
			panic(err)
		}

		response_body_returns := network.GetResponseBodyReturns{}
		mapstructure.Decode(response_body, &response_body_returns)
		fmt.Printf("%v", response_body_returns.Body)

		// or use
		response_body, err = wd.ExecuteChromeDPCommand("Network.getResponseBody", map[string]interface{}{
			"requestId": params["requestId"],
		})
		if err != nil {
			panic(err)
		}
		fmt.Printf("%v", response_body)
	}
}

/*
Package selenium provider a Selenium/Webdriver client.

Currently provides only WebDriver remote client.
This means you'll need to run the Selenium server by yourself (or use a service
like SauceLabs). The easiest way to do that is to grab the Selenium server jar
from http://www.seleniumhq.org/download/ and run it
	java -jar selenium-server-standalone-2.24.1.jar

To use the webdriver with firefox, you may (depending on versions) require the
gecko driver package. You can download it here
        https://github.com/mozilla/geckodriver/releases
and configure the webdriver in your Go code like this
        caps := selenium.Capabilities{
            "browserName":            "firefox",
            "webdriver.gecko.driver": "/path/to/downloaded/geckodriver",
        }
        
Example usage:

	// Run some code on play.golang.org and display the result
	package main

	import (
		"fmt"
		"time"

		"github.com/tebeka/selenium"
	)

	var code = `
	package main
	import "fmt"

	func main() {
		fmt.Println("Hello WebDriver!\n")
	}
	`

	// Errors are ignored for brevity.

	func main() {
		// FireFox driver without specific version
		// *** Add gecko driver here if necessary (see notes above.) ***
		caps := selenium.Capabilities{"browserName": "firefox"}
		wd, err := selenium.NewRemote(caps, "")
		if err != nil {
			panic(err)
		}
		defer wd.Quit()

		// Get simple playground interface
		wd.Get("http://play.golang.org/?simple=1")

		// Enter code in textarea
		elem, _ := wd.FindElement(selenium.ByCSSSelector, "#code")
		elem.Clear()
		elem.SendKeys(code)

		// Click the run button
		btn, _ := wd.FindElement(selenium.ByCSSSelector, "#run")
		btn.Click()

		// Get the result
		div, _ := wd.FindElement(selenium.ByCSSSelector, "#output")

		output := ""
		// Wait for run to finish
		for {
			output, _ = div.Text()
			if output != "Waiting for remote server..." {
				break
			}
			time.Sleep(time.Millisecond * 100)
		}

		fmt.Printf("Got: %s\n", output)
	}
*/
package selenium

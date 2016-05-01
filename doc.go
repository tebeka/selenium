/*
Package selenium provider a Selenium/Webdriver client.

Currently provides only WebDriver remote client.
This means you'll need to run the Selenium server by yourself (or use a service
like SauceLabs). The easiest way to do that is to grab the Selenium server jar
from http://selenium.googlecode.com/files and run it
	java -jar selenium-server-standalone-2.24.1.jar

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
		caps := selenium.Capabilities{"browserName": "firefox"}
		wd, _ := selenium.NewRemote(caps, "")
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

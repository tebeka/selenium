/*
Package selenium provider a WebDriver client.

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
*/
package selenium

# `selenium` - Selenium Client For Go

[![GoDoc](https://godoc.org/github.com/tebeka/selenium?status.svg)](https://godoc.org/github.com/tebeka/selenium)
[![Travis](https://travis-ci.org/tebeka/selenium.svg?branch=master)](https://travis-ci.org/tebeka/selenium)
[![Go Report Card](https://goreportcard.com/badge/github.com/tebeka/selenium)](https://goreportcard.com/report/github.com/tebeka/selenium)

## About

This is a [Selenium][selenium] client for [Go][go]. It supports the WebDriver
protocol and has been tested on [Selenium WebDriver][selenium] and
[ChromeDriver][chromedriver].

`selenium` is currently maintained by Eric Garrido ([@minusnine][minusnine]).

[selenium]: http://seleniumhq.org/
[go]: http://golang.org/
[server]: http://seleniumhq.org/download/
[chromedriver]: https://sites.google.com/a/chromium.org/chromedriver/
[minusnine]: http://github.com/minusnine

## Installing

Run

    go get github.com/tebeka/selenium

## Docs

Docs are at https://godoc.org/github.com/tebeka/selenium

### AppEngine

`GetHTTPClient` exposes the HTTP client used by the driver. You can access it to
add the request context.

    func myRequestHandler(w http.ResponseWriter, r *http.Request) {
        selenium.GetHTTPClient().Transport = &urlfetch.Transport{
            Context:  appengine.NewContext(r),
            Deadline: 30 * time.Second,
        }
        ...
    }

Thanks to [bthomson](https://bitbucket.org/tebeka/selenium/issue/8) for this
one.

## TODO

* Finish full [Selenium API][api].
* Test with Selenium [WebDriver 3.0][webdriver3]
* [SauceLabs][sauce] integration
* Add usage examples
* Support Firefox profiles
* Implement types that provide [all capabilities][allcaps].

[api]: https://www.w3.org/TR/webdriver/
[sauce]: http://saucelabs.com/docs/quickstart
[webdriver3]: https://seleniumhq.wordpress.com/2016/10/13/selenium-3-0-out-now
[allcaps]: https://github.com/SeleniumHQ/selenium/wiki/DesiredCapabilities

## Hacking

Patches are encouraged through GitHub pull requests. Please ensure that a test
is added for anything more than a trivial change and that the existing tests
pass.

### Testing

First, download the ChromeDriver binary and the Selenium WebDriver JARs:

    $ cd vendor
    $ go run init.go
    $ cd ..

You only have to do this once.

Ensure that you have a `firefox` and a `chromium` binary in your path. If the
binaries are named differently, run the tests with the flags
`--firefox_binary=<binary name>` and/or `--chrome_binary=<binary name>` as
appropriate.

#### Locally

Run the tests:

    $ go test 

* There is one top-level test per browser (Chromium and ChromeDriver, and
  Firefox and Selenium). There are subtests that are shared between both
  top-level tests.
* To run only one of the top-level tests, pass `-test.run=TestFirefox` or
  `-test.run=TestChrome`. To run a specific subtest, pass
  `-test.run=Test<Browser>/<subtest>` as appropriate.
* If the `chromium` or `firefox` binaries, the Selenium JAR, or the
  ChromeDriver binary cannot be found, the corresponding tests will be
  skipped.
* The binaries and JAR under test can be configured by passing flags to `go
  test`. See the available flags with `go test --arg --help`.
* Add the argument `-test.v` to see the output from Selenium.

#### With Docker

You will need an installed and running Docker system.

To run the tests under Docker, run:

    $ go test --docker

This will create a new Docker container and run the tests in it. (Note: flags
supplied to this invocation are not curried to the `go test` invocation within
the Docker container).

## License

This project is licensed under the [MIT][mit] license.

[mit]: https://raw.githubusercontent.com/tebeka/selenium/master/LICENSE

# `selenium` - Selenium Client For Go

[![Travis](https://travis-ci.org/tebeka/selenium.svg?branch=master)](https://travis-ci.org/tebeka/selenium)

## About
This is a [Selenium][selenium] client for [Go][go].
Currently it supports only the remote WebDriver client, so you'll need a
[selenium server][server] running.

`selenium` is currently maintained by Eric Garrido (@minusnine).

[selenium]: http://seleniumhq.org/
[go]: http://golang.org/
[server]: http://seleniumhq.org/download/

## Installing

Run

    go get github.com/tebeka/selenium

## Docs

Docs are at [godoc.org][godoc]. 

[godoc]: https://godoc.org/github.com/tebeka/selenium

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

## Changes
See [here][changelog].

[changelog]: ChangeLog

## TODO
* Support Firefox profiles
* Finish full [Selenium API][api].
* More testing
* [Selenium 1][sel1] client
* Our own server for testing (getting out takes too much time)
* [SauceLabs][sauce] integration

[api]: http://code.google.com/p/selenium/wiki/JsonWireProtocol
[sel1]: http://wiki.openqa.org/display/SRC/Specifications+for+Selenium+Remote+Control+Client+Driver+Protocol
[sauce]: http://saucelabs.com/docs/quickstart

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

Run the tests under docker:

    $ go test --docker

This will create a new docker container and run the tests in it. (Note: flags
supplied to this invocation are not curried to the `go test` invocation within
the Docker container).

## License

This project is licensed under the [MIT][mit] license.

[mit]: https://raw.githubusercontent.com/tebeka/selenium/master/LICENSE

## Disclaimer

This is not an official Google product.

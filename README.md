`selenium` - Selenium Client For Go

# About
This is a [Selenium][selenium] client for [Go][go].
Currently it supports only the remote WebDriver client, so you'll need a
[selenium server][server] running.

[selenium]: http://seleniumhq.org/
[go]: http://golang.org/
[server]: http://seleniumhq.org/download/

# Installing
Run

    go get bitbucket.org/tebeka/selenium

# Docs
Docs are at [godoc.org][godoc]. 

[godoc]: http://godoc.org/bitbucket.org/tebeka/selenium

## AppEngine

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

# Changes
See [here][changelog].

[changelog]: ChangeLog

# TODO
* Support Firefox profiles
* Finish full [Selenium API][api].
* More testing
* [Selenium 1][sel1] client
* Our own server for testing (getting out takes too much time)
* [SauceLabs][sauce] integration

[api]: http://code.google.com/p/selenium/wiki/JsonWireProtocol
[sel1]: http://wiki.openqa.org/display/SRC/Specifications+for+Selenium+Remote+Control+Client+Driver+Protocol
[sauce]: http://saucelabs.com/docs/quickstart

# Hacking

* You'll need a Selenium server to run the tests, run `selenium.sh download` to
  get it and `selenium.sh start` to run it.
* Test with `./run-tests.sh`.
* I (Miki) work on `dev` branch since `go get` pull from default.

# License
[MIT][mit]

[mit]: https://bitbucket.org/tebeka/selenium/src/tip/LICENSE.txt

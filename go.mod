module github.com/wanmail/selenium

go 1.12

require (
	cloud.google.com/go v0.41.0
	github.com/BurntSushi/xgbutil v0.0.0-20160919175755-f7c97cef3b4e
	github.com/armon/go-socks5 v0.0.0-20160902184237-e75332964ef5
	github.com/blang/semver v3.5.1+incompatible
	github.com/chromedp/cdproto v0.0.0-20200209033844-7e00b02ea7d2
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.3.4
	github.com/google/go-cmp v0.3.0
	github.com/google/go-github/v27 v27.0.4
	github.com/mailru/easyjson v0.7.0
	github.com/mediabuyerbot/go-crx3 v1.3.1
	github.com/tebeka/selenium v0.9.9 
	google.golang.org/api v0.7.0
)

replace github.com/tebeka/selenium => github.com/wanmail/selenium

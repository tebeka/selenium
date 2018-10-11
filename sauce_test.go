package selenium

import (
	"flag"
	"fmt"
	"strings"
	"testing"

	"github.com/blang/semver"
	"github.com/tebeka/selenium/sauce"
)

var (
	enableSauce      = flag.Bool("experimental_enable_sauce", false, "If true, perform integration tests on SauceLabs remote infrastructure.")
	sauceUserName    = flag.String("sauce_user_name", "", "The username to use for SauceLabs.")
	sauceAccessKey   = flag.String("sauce_access_key", "", "The access key to use for SauceLabs.")
	sauceConnectPath = flag.String("sauce_connect_path", "vendor/sauce-connect/bin/sc", "The path to the Sauce Connect binary.")
)

func TestSauce(t *testing.T) {
	if !*enableSauce {
		t.Skip("Skipping Sauce tests. Enable via --experimental_sauce_tests")
	}
	if testing.Verbose() {
		SetDebug(true)
	}
	if *sauceUserName == "" {
		t.Fatalf("--sauce_user_name is required.")
	}
	if *sauceAccessKey == "" {
		t.Fatalf("--sauce_access_key is required.")
	}

	addr := sauce.Addr(*sauceUserName, *sauceAccessKey)
	if *sauceConnectPath != "" {
		port, err := pickUnusedPort()
		if err != nil {
			t.Fatalf("pickUnusedPort() returned error: %s", err)
		}
		sc := sauce.Connect{
			Path:                *sauceConnectPath,
			UserName:            *sauceUserName,
			AccessKey:           *sauceAccessKey,
			SeleniumPort:        port,
			QuitProcessUponExit: true,
		}
		if testing.Verbose() {
			sc.ExtraVerbose = true
		}
		if err := sc.Start(); err != nil {
			t.Fatalf("sc.Start() returned error: %s", err)
		}
		defer func() {
			if err := sc.Stop(); err != nil {
				t.Fatalf("sc.Stop() returned error: %s", err)
			}
		}()
		addr = sc.Addr()
	}

	const browser = "Firefox"
	for _, tc := range []struct {
		version, platform, selenium string
	}{
		{"Latest", "Windows 10", "3.4.0"},
		{"45.0", "Linux", "2.53.1"},
	} {
		name := fmt.Sprintf("%s/%s/%s/%s", browser, tc.version, tc.platform, tc.selenium)
		browser, version := strings.ToLower(browser), strings.ToLower(tc.version)

		t.Run(name, func(t *testing.T) {
			runFirefoxSubTests(t, config{
				browser:         browser,
				seleniumVersion: semver.MustParse(tc.selenium),
				sauce: &sauce.Capabilities{
					Browser:         browser,
					Version:         version,
					Platform:        tc.platform,
					SeleniumVersion: tc.selenium,
				},
				addr: addr,
			})
		})
	}
}

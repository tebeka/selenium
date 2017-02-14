// Binary init downloads the necessary files to perform an integration test between this WebDriver client and multiple versions of Selenium and browsers.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"

	"github.com/golang/glog"
)

// TODO(minusnine): download the Chrome binary.
var downloadBrowser = flag.Bool("download_browsers", true, "If true, download the Firefox binary.")

type file struct {
	url     string
	name    string
	hash    string
	rename  []string
	browser bool
}

var files = []file{
	{
		url:  "http://selenium-release.storage.googleapis.com/3.0/selenium-server-standalone-3.0.1.jar",
		name: "selenium-server-standalone-3.0.1.jar",
		hash: "1537b6d1b259191ed51586378791bc62b38b0cb18ae5ba1433009dc365e9f26b",
	},
	{
		url:  "http://selenium-release.storage.googleapis.com/2.53/selenium-server-standalone-2.53.1.jar",
		name: "selenium-server-standalone-2.53.1.jar",
		hash: "1cce6d3a5ca5b2e32be18ca5107d4f21bddaa9a18700e3b117768f13040b7cf8",
	},
	{
		url:    "https://chromedriver.storage.googleapis.com/2.26/chromedriver_linux64.zip",
		name:   "chromedriver_2.26_linux64.zip",
		hash:   "59e6b1b1656a20334d5731b3c5a7400f92a9c6f5043bb4ab67f1ccf1979ee486",
		rename: []string{"chromedriver", "chromedriver-linux64-2.26"},
	},
	{
		url:    "https://chromedriver.storage.googleapis.com/2.27/chromedriver_linux64.zip",
		name:   "chromedriver_2.27_linux64.zip",
		hash:   "9c88402f4d9dca822697c3aa4623285e3b5b83a12b1261421c9a50d1960eb137",
		rename: []string{"chromedriver", "chromedriver-linux64-2.27"},
	},
	{
		url:    "https://github.com/mozilla/geckodriver/releases/download/v0.14.0/geckodriver-v0.14.0-linux64.tar.gz",
		name:   "geckodriver-v0.14.0-linux64.tar.gz",
		hash:   "aaae25e9197360261f966f6129b47ebdf75ee9da63c74c9f39397b1100cd9653",
		rename: []string{"geckodriver", "geckodriver-v0.14.0-linux64"},
	},
	{
		url:     "https://ftp.mozilla.org/pub/firefox/releases/47.0.2/linux-x86_64/en-US/firefox-47.0.2.tar.bz2",
		name:    "firefox-47-0.2.tar.bz2",
		hash:    "ea88e5d18438d1b80e6048fa2cfbaa90875fba8f42ef5bddc191b6bfd90af672",
		browser: true,
		rename:  []string{"firefox", "firefox-47"},
	},
	{
		// This is a recent nightly. Update this path periodically.
		url:     "https://archive.mozilla.org/pub/firefox/nightly/2017/02/2017-02-14-11-02-12-mozilla-central/firefox-54.0a1.en-US.linux-x86_64.tar.bz2",
		name:    "firefox-54.0a1.en-US.linux-x86_64.tar.bz2",
		hash:    "df3dcecbb630ca662851160b4d619a11c3ab52a8ceb238eb00e522248a3890ee",
		browser: true,
		rename:  []string{"firefox", "firefox-nightly"},
	},
}

func main() {
	flag.Parse()

	for _, file := range files {
		if file.browser && !*downloadBrowser {
			glog.Infof("Skipping %q because --download_browser is not set.", file.name)
			continue
		}
		if !fileSameHash(file.name, file.hash) {
			glog.Infof("Downloading %q from %q", file.name, file.url)
			if err := downloadFile(file); err != nil {
				glog.Exit(err.Error())
			}
		} else {
			glog.Infof("Skipping file %q which has already been downloaded.", file.name)
		}
		switch path.Ext(file.name) {
		case ".zip":
			glog.Infof("Unzipping %q", file.name)
			if err := exec.Command("unzip", file.name).Run(); err != nil {
				glog.Exitf("Error unzipping %q: %v", file.name, err)
			}
		case ".gz":
			glog.Infof("Unzipping %q", file.name)
			if err := exec.Command("tar", "-xzf", file.name).Run(); err != nil {
				glog.Exitf("Error unzipping %q: %v", file.name, err)
			}
		case ".bz2":
			glog.Infof("Unzipping %q", file.name)
			if err := exec.Command("tar", "-xjf", file.name).Run(); err != nil {
				glog.Exitf("Error unzipping %q: %v", file.name, err)
			}
		}
		if rename := file.rename; len(rename) == 2 {
			glog.Infof("Renaming %q to %q", rename[0], rename[1])
			os.RemoveAll(rename[1]) // Ignore error.
			if err := os.Rename(rename[0], rename[1]); err != nil {
				glog.Warningf("Error renaming %q to %q: %v", rename[0], rename[1], err)
			}
		}
	}
}

func downloadFile(file file) (err error) {
	f, err := os.Create(file.name)
	if err != nil {
		return fmt.Errorf("error creating %q: %v", file.name, err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("error closing %q: %v", file.name, err)
		}
	}()

	resp, err := http.Get(file.url)
	if err != nil {
		return fmt.Errorf("%s: error downloading %q: %v", file.name, file.url, err)
	}
	defer resp.Body.Close()

	hash := sha256.New()
	tee := io.MultiWriter(f, hash)
	if _, err := io.Copy(tee, resp.Body); err != nil {
		return fmt.Errorf("%s: error downloading %q: %v", file.name, file.url, err)
	}
	if h := hex.EncodeToString(hash.Sum(nil)); h != file.hash {
		return fmt.Errorf("%s: got sha256 hash %q, want %q", file.name, h, file.hash)
	}
	return nil
}

func fileSameHash(fileName, wantHash string) bool {
	if _, err := os.Stat(fileName); err != nil {
		return false
	}
	h := sha256.New()
	f, err := os.Open(fileName)
	if err != nil {
		return false
	}
	defer f.Close()

	if _, err := io.Copy(h, f); err != nil {
		return false
	}

	sum := hex.EncodeToString(h.Sum(nil))
	if sum != wantHash {
		glog.Warningf("File %q: got hash %q, expect hash %q", fileName, sum, wantHash)
		return false
	}
	return true
}

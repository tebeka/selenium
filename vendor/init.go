// Binary init downloads the necessary files to perform an integration test between this WebDriver client and multiple versions of Selenium and browsers.
package main

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/golang/glog"
	"google.golang.org/api/option"
)

// consts for downloading Google chrome browser.
const (
	// Bucket URL: https://console.cloud.google.com/storage/browser/chromium-browser-continuous/?pli=1
	storageBktName = "chromium-browser-continuous"
	prefixLinux64  = "Linux_x64"
	lastChangeFile = "Linux_x64/LAST_CHANGE"
	chromeFilename = "chrome-linux.zip"
)

// TODO(minusnine): download the Chrome binary.
var downloadBrowsers = flag.Bool("download_browsers", true, "If true, download the Firefox and Chrome browsers.")

type file struct {
	url      string
	name     string
	hash     string
	hashtype string //default is sha256
	rename   []string
	browser  bool
}

var files = []file{
	{
		url:  "http://selenium-release.storage.googleapis.com/3.3/selenium-server-standalone-3.3.1.jar",
		name: "selenium-server-standalone-3.3.1.jar",
		hash: "94a0bd034636a2430d9d52b73b8e29e819af103ab84000de241ca83eb4e142f6",
	},
	{
		url:  "http://selenium-release.storage.googleapis.com/2.53/selenium-server-standalone-2.53.1.jar",
		name: "selenium-server-standalone-2.53.1.jar",
		hash: "1cce6d3a5ca5b2e32be18ca5107d4f21bddaa9a18700e3b117768f13040b7cf8",
	},
	{
		url:    "https://chromedriver.storage.googleapis.com/2.28/chromedriver_linux64.zip",
		name:   "chromedriver_2.28_linux64.zip",
		hash:   "8f5b0ab727c326a2f7887f08e4f577cb4452a9e5783d1938728946a8557a37bc",
		rename: []string{"chromedriver", "chromedriver-linux64-2.28"},
	},
	{
		url:    "https://github.com/mozilla/geckodriver/releases/download/v0.15.0/geckodriver-v0.15.0-linux64.tar.gz",
		name:   "geckodriver-v0.15.0-linux64.tar.gz",
		hash:   "6e24178195e6552375c3fd45dc50593e46fe2711e7907e84fecb4e3a5cb013ea",
		rename: []string{"geckodriver", "geckodriver-v0.15.0-linux64"},
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
		url:     "https://archive.mozilla.org/pub/firefox/nightly/2017/03/2017-03-11-11-02-44-mozilla-central/firefox-55.0a1.en-US.linux-x86_64.tar.bz2",
		name:    "firefox-54.0a1.en-US.linux-x86_64.tar.bz2",
		hash:    "006a42297df774c4bd34bcf7f889ccdbc5ca3d2d443204915d4ad55ac3b5d01b",
		browser: true,
		rename:  []string{"firefox", "firefox-nightly"},
	},
}

func addChrome() (err error) {
	var chromeFile file
	chromeFile.name = chromeFilename
	chromeFile.browser = true
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithHTTPClient(http.DefaultClient))
	if err != nil {
		return fmt.Errorf("Cannot create a storage client for downloading the chrome browser: %s", err)
	}
	BktHandle := client.Bucket(storageBktName)
	lcFileObj := BktHandle.Object(lastChangeFile)
	rc, err := lcFileObj.NewReader(ctx)
	if err != nil {
		return fmt.Errorf("Cannot create a reader for last_change file: %s", err)
	}
	defer rc.Close()
	// Read the last change file content for the latest build directory name
	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return fmt.Errorf("Cannot read from LAST_CHANGE: %s", err)
	}
	latestChromeBuild := string(data)
	latestChromePackage := strings.Join([]string{prefixLinux64, latestChromeBuild, chromeFilename}, "/")
	chromeObjHandler := BktHandle.Object(latestChromePackage)
	cpAttrs, err := chromeObjHandler.Attrs(ctx)
	if err != nil {
		return fmt.Errorf("Cannot get the chrome package attrs: %s", err)
	}
	// 02x is needed as the returned []byte is not a valid utf8
	chromeFile.hash = fmt.Sprintf("%02x", string(cpAttrs.MD5))
	chromeFile.url = cpAttrs.MediaLink
	chromeFile.hashtype = "md5"
	files = append(files, chromeFile)
	return nil
}

func main() {
	flag.Parse()
	if *downloadBrowsers {
		err := addChrome()
		if err != nil {
			glog.Infof("Unable to Download Google Chrome browser. Continuing...")
		}
	}

	for _, file := range files {
		if file.browser && !*downloadBrowsers {
			glog.Infof("Skipping %q because --download_browser is not set.", file.name)
			continue
		}
		if !fileSameHash(file) {
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
	var h hash.Hash
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
	htype := strings.ToLower(file.hashtype)
	switch htype {
	case "md5":
		h = md5.New()
	default:
		h = sha256.New()
	}
	tee := io.MultiWriter(f, h)
	if _, err := io.Copy(tee, resp.Body); err != nil {
		return fmt.Errorf("%s: error downloading %q: %v", file.name, file.url, err)
	}
	if h := hex.EncodeToString(h.Sum(nil)); h != file.hash {
		return fmt.Errorf("%s: got %s hash %q, want %q", file.name, file.hashtype, h, file.hash)
	}
	return nil
}

func fileSameHash(file file) bool {
	var h hash.Hash
	if _, err := os.Stat(file.name); err != nil {
		return false
	}
	htype := strings.ToLower(file.hashtype)
	switch htype {
	case "md5":
		h = md5.New()
	default:
		h = sha256.New()
	}
	f, err := os.Open(file.name)
	if err != nil {
		return false
	}
	defer f.Close()

	if _, err := io.Copy(h, f); err != nil {
		return false
	}

	sum := hex.EncodeToString(h.Sum(nil))
	if sum != file.hash {
		glog.Warningf("File %q: got hash %q, expect hash %q", file.name, sum, file.hash)
		return false
	}
	return true
}

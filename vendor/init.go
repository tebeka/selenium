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
	"sync"

	"cloud.google.com/go/storage"
	"github.com/golang/glog"
	"google.golang.org/api/option"
)

var downloadBrowsers = flag.Bool("download_browsers", true, "If true, download the Firefox and Chrome browsers.")

type file struct {
	url      string
	name     string
	hash     string
	hashType string // default is sha256
	rename   []string
	browser  bool
}

var files = []file{
	{
		url:  "https://selenium-release.storage.googleapis.com/3.8/selenium-server-standalone-3.8.1.jar",
		name: "selenium-server-standalone-3.8.1.jar",
		hash: "2ca30da4a482688263b0eed5c73d1a4bbf3116316a1f2ffb96310a1643dbe663",
	},
	{
		url:    "https://chromedriver.storage.googleapis.com/2.35/chromedriver_linux64.zip",
		name:   "chromedriver_2.35_linux64.zip",
		hash:   "67fad24c4a85e3f33f51c97924a98b619722db15ce92dcd27484fb748af93e8e",
		rename: []string{"chromedriver", "chromedriver-linux64-2.35"},
	},
	{
		url:    "https://github.com/mozilla/geckodriver/releases/download/v0.19.1/geckodriver-v0.19.1-linux64.tar.gz",
		name:   "geckodriver-v0.19.1-linux64.tar.gz",
		hash:   "7f55c4c89695fd1e6f8fc7372345acc1e2dbaa4a8003cee4bd282eed88145937",
		rename: []string{"geckodriver", "geckodriver-v0.19.1-linux64"},
	},
	{
		// This is a recent nightly. Update this path periodically.
		url:     "https://archive.mozilla.org/pub/firefox/nightly/2017/08/2017-08-21-10-03-50-mozilla-central/firefox-57.0a1.en-US.linux-x86_64.tar.bz2",
		name:    "firefox-57.0a1.en-US.linux-x86_64.tar.bz2",
		hash:    "77c57356935f66a5a59b1b2cffeaa53b70204195e6a7b15ee828fd3308561e46",
		browser: true,
		rename:  []string{"firefox", "firefox-nightly"},
	},
	{
		url:    "https://saucelabs.com/downloads/sc-4.4.9-linux.tar.gz",
		name:   "sauce-connect-4.4.9-linux.tar.gz",
		hash:   "b1bedccc2690b48d6708ac71f23189c85b0da62c56ee943a1b20d8f17fa8bbde",
		rename: []string{"sc-4.4.9-linux", "sauce-connect"},
	},
}

func addChrome(ctx context.Context) error {
	const (
		// Bucket URL: https://console.cloud.google.com/storage/browser/chromium-browser-continuous/?pli=1
		storageBktName = "chromium-browser-snapshots"
		prefixLinux64  = "Linux_x64"
		lastChangeFile = "Linux_x64/LAST_CHANGE"
		chromeFilename = "chrome-linux.zip"
	)
	gcsPath := fmt.Sprintf("gs://%s/", storageBktName)
	client, err := storage.NewClient(ctx, option.WithHTTPClient(http.DefaultClient))
	if err != nil {
		return fmt.Errorf("cannot create a storage client for downloading the chrome browser: %v", err)
	}
	bkt := client.Bucket(storageBktName)
	r, err := bkt.Object(lastChangeFile).NewReader(ctx)
	if err != nil {
		return fmt.Errorf("cannot create a reader for %s%s file: %v", gcsPath, lastChangeFile, err)
	}
	defer r.Close()
	// Read the last change file content for the latest build directory name
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return fmt.Errorf("cannot read from %s%s file: %v", gcsPath, lastChangeFile, err)
	}
	latestChromeBuild := string(data)
	latestChromePackage := path.Join(prefixLinux64, latestChromeBuild, chromeFilename)
	cpAttrs, err := bkt.Object(latestChromePackage).Attrs(ctx)
	if err != nil {
		return fmt.Errorf("cannot get the chrome package %s%s attrs: %v", gcsPath, latestChromePackage, err)
	}
	files = append(files, file{
		name:     chromeFilename,
		browser:  true,
		hash:     hex.EncodeToString(cpAttrs.MD5),
		hashType: "md5",
		url:      cpAttrs.MediaLink,
	})
	return nil
}

func main() {
	flag.Parse()
	ctx := context.Background()
	if *downloadBrowsers {
		if err := addChrome(ctx); err != nil {
			glog.Errorf("unable to Download Google Chrome browser: %v", err)
		}
	}
	var wg sync.WaitGroup
	for _, file := range files {
		wg.Add(1)
		file := file
		go func() {
			if err := handleFile(file); err != nil {
				glog.Exitf("Error handling %s: %s", file.name, err)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func handleFile(file file) error {
	if file.browser && !*downloadBrowsers {
		glog.Infof("Skipping %q because --download_browser is not set.", file.name)
		return nil
	}
	if !fileSameHash(file) {
		glog.Infof("Downloading %q from %q", file.name, file.url)
		if err := downloadFile(file); err != nil {
			return err
		}
	} else {
		glog.Infof("Skipping file %q which has already been downloaded.", file.name)
	}

	switch path.Ext(file.name) {
	case ".zip":
		glog.Infof("Unzipping %q", file.name)
		if err := exec.Command("unzip", "-o", file.name).Run(); err != nil {
			return fmt.Errorf("Error unzipping %q: %v", file.name, err)
		}
	case ".gz":
		glog.Infof("Unzipping %q", file.name)
		if err := exec.Command("tar", "-xzf", file.name).Run(); err != nil {
			return fmt.Errorf("Error unzipping %q: %v", file.name, err)
		}
	case ".bz2":
		glog.Infof("Unzipping %q", file.name)
		if err := exec.Command("tar", "-xjf", file.name).Run(); err != nil {
			return fmt.Errorf("Error unzipping %q: %v", file.name, err)
		}
	}
	if rename := file.rename; len(rename) == 2 {
		glog.Infof("Renaming %q to %q", rename[0], rename[1])
		os.RemoveAll(rename[1]) // Ignore error.
		if err := os.Rename(rename[0], rename[1]); err != nil {
			glog.Warningf("Error renaming %q to %q: %v", rename[0], rename[1], err)
		}
	}
	return nil
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
	var h hash.Hash
	switch strings.ToLower(file.hashType) {
	case "md5":
		h = md5.New()
	default:
		h = sha256.New()
	}
	if _, err := io.Copy(io.MultiWriter(f, h), resp.Body); err != nil {
		return fmt.Errorf("%s: error downloading %q: %v", file.name, file.url, err)
	}
	if h := hex.EncodeToString(h.Sum(nil)); h != file.hash {
		return fmt.Errorf("%s: got %s hash %q, want %q", file.name, file.hashType, h, file.hash)
	}
	return nil
}

func fileSameHash(file file) bool {
	if _, err := os.Stat(file.name); err != nil {
		return false
	}
	var h hash.Hash
	switch strings.ToLower(file.hashType) {
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

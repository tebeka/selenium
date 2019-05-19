// Binary init downloads the necessary files to perform an integration test
// between this WebDriver client and multiple versions of Selenium and
// browsers.
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
		url:  "https://selenium-release.storage.googleapis.com/3.14/selenium-server-standalone-3.14.0.jar",
		name: "selenium-server-standalone-3.14.0.jar",
		hash: "d38decb0a2ca9196e404838cd3866cd0129af5ee167c592b690e502f0b3d01ad",
	},
	{
		url:    "https://chromedriver.storage.googleapis.com/2.42/chromedriver_linux64.zip",
		name:   "chromedriver_2.42_linux64.zip",
		hash:   "ced920ea978a658d915a960ed7bff62c37c9b7d62dfabbb42ffd7fe7cef7f082",
		rename: []string{"chromedriver", "chromedriver-linux64-2.42"},
	},
	{
		url:    "https://github.com/mozilla/geckodriver/releases/download/v0.23.0/geckodriver-v0.23.0-linux64.tar.gz",
		name:   "geckodriver-v0.23.0-linux64.tar.gz",
		hash:   "2abf02cb69b48f2ba65ea344b752ff547e5431659aad80b03bf68cdb4f8df14b",
		rename: []string{"geckodriver", "geckodriver-v0.23.0-linux64"},
	},
	{
		// This is a recent nightly. Update this path periodically.
		url:     "https://archive.mozilla.org/pub/firefox/nightly/2018/09/2018-09-14-10-01-56-mozilla-central/firefox-64.0a1.en-US.linux-x86_64.tar.bz2",
		name:    "firefox-64.0a1.en-US.linux-x86_64.tar.bz2",
		hash:    "8b7eb6a0c9985e9baf4adb6e557ec25e63cd6d55aa8c23718c1f126d3c5e6bc0",
		browser: true,
		rename:  []string{"firefox", "firefox-nightly"},
	},
	{
		url:    "https://saucelabs.com/downloads/sc-4.5.1-linux.tar.gz",
		name:   "sauce-connect-4.5.1-linux.tar.gz",
		hash:   "dd53f2cdcec489fbc2443942b853b51bf44af39f230600573119cdd315ddee52",
		rename: []string{"sc-4.5.1-linux", "sauce-connect"},
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

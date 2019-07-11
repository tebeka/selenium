package download

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/golang/glog"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/option"
)

// File describes how to download a file from the Web.
type File struct {
	url      string
	Name     string
	hash     string
	hashType string // default is sha256
	Rename   []string
	Browser  bool
	// The directory in which to store the file.
	directory string
}

func (f File) Path() string {
	if f.directory != "" {
		return filepath.Join(f.directory, f.Name)
	}
	return f.Name
}

var (
	// SeleniumFile describes how to download the Selenium standalone JAR.
	SeleniumFile = File{
		url:  "https://selenium-release.storage.googleapis.com/3.141/selenium-server-standalone-3.141.59.jar",
		Name: "selenium-server.jar",
		// TODO(minusnine): reimplement hashing so that it is less annoying for maintenance.
		hash: "acf71b77d1b66b55db6fb0bed6d8bae2bbd481311bcbedfeff472c0d15e8f3cb",
	}

	// ChromeDriverFile describes how to download the ChromeDriver binary.
	ChromeDriverFile = File{
		url:  "https://chromedriver.storage.googleapis.com/76.0.3809.25/chromedriver_linux64.zip",
		Name: "chromedriver.zip",
		hash: "0a264a8b2fa881edf33657ba88709ae3dbaec72d8b41beebf1c89d5e3bc3e594",
	}

	// GeckodriverFile describes how to download the Geckodriver binary.
	GeckodriverFile = File{
		url:  "https://github.com/mozilla/geckodriver/releases/download/v0.24.0/geckodriver-v0.24.0-linux64.tar.gz",
		Name: "geckodriver.tar.gz",
		hash: "03be3d3b16b57e0f3e7e8ba7c1e4bf090620c147e6804f6c6f3203864f5e3784",
	}

	// FirefoxNightly describes how to download the nightly Firefox binary.
	FirefoxNightlyFile = File{
		// This is a recent nightly. Update this path periodically.
		url:     "https://download.mozilla.org/?product=firefox-nightly-latest-ssl&os=linux64&lang=en-US",
		Name:    "firefox-nightly.tar.bz2",
		hash:    "cedd2028dd01280e9f68b835b4c569c6eee46d2227f7cbda88f8365b8ec315a1",
		Browser: true,
		Rename:  []string{"firefox", "firefox-nightly"},
	}

	// SauceConnectProxyFile describes how to download the SauceConnect binary.
	SauceConnectProxyFile = File{
		url:    "https://saucelabs.com/downloads/sc-4.5.3-linux.tar.gz",
		Name:   "sauce-connect.tar.gz",
		hash:   "0de7fcbcb03ad400e886f2c4b34661eda55808e69c7bc4db6aa6aff85e4edb15",
		Rename: []string{"sc-4.5.3-linux", "sauce-connect"},
	}
)

// AllFiles includes all binary dependencies required to test the selenium
// package.
func AllFiles(ctx context.Context) ([]File, error) {
	allFiles := []File{
		SeleniumFile, ChromeDriverFile, GeckodriverFile, FirefoxNightlyFile, SauceConnectProxyFile,
	}

	chrome, err := ChromeSnapshotFile(ctx)
	if err != nil {
		return nil, err
	}
	return append(allFiles, chrome), nil
}

// Obtain a File that describes how to download the latest Chrome snapshot.
func ChromeSnapshotFile(ctx context.Context) (File, error) {
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
		return File{}, fmt.Errorf("cannot create a storage client for downloading the chrome browser: %v", err)
	}

	bkt := client.Bucket(storageBktName)
	r, err := bkt.Object(lastChangeFile).NewReader(ctx)
	if err != nil {
		return File{}, fmt.Errorf("cannot create a reader for %s%s file: %v", gcsPath, lastChangeFile, err)
	}
	defer r.Close()

	// Read the last change file content for the latest build directory name
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return File{}, fmt.Errorf("cannot read from %s%s file: %v", gcsPath, lastChangeFile, err)
	}

	latestChromeBuild := string(data)
	latestChromePackage := path.Join(prefixLinux64, latestChromeBuild, chromeFilename)
	cpAttrs, err := bkt.Object(latestChromePackage).Attrs(ctx)
	if err != nil {
		return File{}, fmt.Errorf("cannot get the chrome package %s%s attrs: %v", gcsPath, latestChromePackage, err)
	}

	return File{
		Name:     chromeFilename,
		Browser:  true,
		hash:     hex.EncodeToString(cpAttrs.MD5),
		hashType: "md5",
		url:      cpAttrs.MediaLink,
	}, nil
}

// Download a file if it is not already present. If directory is the empty
// string, the files will be downloaded to the current directory.
func Download(file File, directory string) error {
	file.directory = directory

	if file.hash != "" && fileSameHash(file) {
		glog.Infof("Skipping file %q which has already been downloaded.", file.Name)
	} else {
		glog.Infof("Downloading %q from %q", file.Name, file.url)
		if err := downloadFile(file); err != nil {
			return err
		}
	}

	if err := unzipArchive(file); err != nil {
		return err
	}

	if rename := file.Rename; len(rename) == 2 {
		from := filepath.Join(file.directory, rename[0])
		to := filepath.Join(file.directory, rename[1])
		glog.Infof("Renaming %q to %q", from, to)
		os.RemoveAll(rename[1]) // Ignore error.
		if err := os.Rename(from, to); err != nil {
			glog.Warningf("Error renaming %q to %q: %v", from, to, err)
		}
	}
	return nil
}

func DownloadAll(ctx context.Context, directory string) error {
	allFiles, err := AllFiles(ctx)
	if err != nil {
		glog.Exit(err.Error())
	}

	var wg errgroup.Group
	for _, file := range allFiles {
		file := file
		wg.Go(func() error {
			file.directory = directory
			if err := Download(file, directory); err != nil {
				return fmt.Errorf("error handling %s: %s", file.Name, err)
			}
			return nil
		})
	}
	return wg.Wait()
}

func downloadFile(file File) (err error) {
	f, err := os.Create(file.Path())
	if err != nil {
		return fmt.Errorf("error creating %q: %v", file.Path(), err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("error closing %q: %v", file.Path(), err)
		}
	}()

	resp, err := http.Get(file.url)
	if err != nil {
		return fmt.Errorf("%s: error downloading %q: %v", file.Name, file.url, err)
	}
	defer resp.Body.Close()
	if file.hash != "" {
		var h hash.Hash
		switch strings.ToLower(file.hashType) {
		case "md5":
			h = md5.New()
		case "sha1":
			h = sha1.New()
		default:
			h = sha256.New()
		}
		if _, err := io.Copy(io.MultiWriter(f, h), resp.Body); err != nil {
			return fmt.Errorf("%s: error downloading %q: %v", file.Name, file.url, err)
		}
		if h := hex.EncodeToString(h.Sum(nil)); h != file.hash {
			return fmt.Errorf("%s: got %s hash %q, want %q", file.Name, file.hashType, h, file.hash)
		}
	} else {
		if _, err := io.Copy(f, resp.Body); err != nil {
			return fmt.Errorf("%s: error downloading %q: %v", file.Name, file.url, err)
		}
	}
	return nil
}

func fileSameHash(file File) bool {
	if _, err := os.Stat(file.Path()); err != nil {
		return false
	}
	var h hash.Hash
	switch strings.ToLower(file.hashType) {
	case "md5":
		h = md5.New()
	default:
		h = sha256.New()
	}
	f, err := os.Open(file.Path())
	if err != nil {
		return false
	}
	defer f.Close()

	if _, err := io.Copy(h, f); err != nil {
		return false
	}

	sum := hex.EncodeToString(h.Sum(nil))
	if sum != file.hash {
		glog.Warningf("File %q: got hash %q, expect hash %q", file.Name, sum, file.hash)
		return false
	}
	return true
}

func unzipArchive(file File) error {
	var unzipCmd []string

	dir := "."
	if file.directory != "" {
		dir = file.directory
	}

	switch path.Ext(file.Name) {
	case ".zip":
		unzipCmd = []string{"unzip", "-d", dir, "-o", file.Path()}
	case ".gz":
		unzipCmd = []string{"tar", "-xzf", file.Path(), "-C", dir}
	case ".bz2":
		unzipCmd = []string{"tar", "-xjf", file.Path(), "-C", dir}
	default:
		return nil
	}

	glog.Infof("Unzipping %q", file.Path())
	if err := exec.Command(unzipCmd[0], unzipCmd[1:]...).Run(); err != nil {
		return fmt.Errorf("error unzipping %q: %v", file.Name, err)
	}

	return nil
}

func archiveUnchanged(file File) {
}

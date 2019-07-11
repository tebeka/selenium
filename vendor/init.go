// Binary init downloads the necessary files to perform an integration test
// between this WebDriver client and multiple versions of Selenium and
// browsers.
package main

import (
	"context"
	"flag"
	"os"
	"strings"

	"github.com/golang/glog"
	"github.com/tebeka/selenium/internal/download"
)

var clean = flag.Bool("clean", false, "Instead of downloading files, remove them.")

func main() {
	// TODO(minusnine): If --alsologtostderr was not set at all, set it to true.
	flag.Parse()
	ctx := context.Background()

	if *clean {
		allFiles, err := download.AllFiles(ctx)
		if err != nil {
			glog.Exit(err.Error())
		}
		for _, f := range allFiles {
			glog.Infof("Removing %s", f.Path())
			if err := os.Remove(f.Path()); err != nil && !os.IsNotExist(err) {
				glog.Errorf("Error removing %s: %s", f.Path(), err)
			}
			for _, ext := range []string{".tar.gz", ".tar.bz2", ".zip"} {
				if strings.HasSuffix(f.Name, ext) {
					path := strings.TrimSuffix(f.Path(), ext)
					glog.Infof("Removing %s", path)
					if err := os.RemoveAll(path); err != nil && !os.IsNotExist(err) {
						glog.Errorf("Error removing %s: %s", f.Path(), err)
					}
					break
				}
			}
		}
		return
	}

	if err := download.DownloadAll(ctx, ""); err != nil {
		glog.Exit(err.Error())
	}
}

// Package zip creates Zip files.
package zip

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// New returns a buffer that contains the payload of a Zip file.
func New(basePath string) (*bytes.Buffer, error) {
	fi, err := os.Stat(basePath)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return nil, fmt.Errorf("path %q is not a directory, which is required for a Firefox profile", basePath)
	}

	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	err = filepath.Walk(basePath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}

		zipFI, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// Strip the prefix from the filename (and the trailing directory
		// separator) so that the files are at the root of the zip file.
		zipFI.Name = filePath[len(basePath)+1:]

		// Without this, the Java zip reader throws a java.util.zip.ZipException:
		// "only DEFLATED entries can have EXT descriptor".
		zipFI.Method = zip.Deflate

		w, err := w.CreateHeader(zipFI)
		if err != nil {
			return err
		}

		f, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(w, bufio.NewReader(f))
		return err
	})
	if err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf, nil
}

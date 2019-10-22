package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrIllegalPath = errors.New("unzip: illegal file path")
)

func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		err = unzipFile(dest, f)
		if err != nil {
			return err
		}
	}

	return nil
}

func unzipFile(dest string, f *zip.File) error {
	filePath := filepath.Join(dest, f.Name)
	if !strings.HasPrefix(filePath, filepath.Clean(dest)+string(os.PathSeparator)) {
		return ErrIllegalPath
	}

	if f.FileInfo().IsDir() {
		// Create directory
		if err := os.MkdirAll(filePath, f.Mode()); err != nil {
			return fmt.Errorf("unzip: %w", err)
		}
	} else {
		// Create file
		rdr, err := f.Open()
		if err != nil {
			return fmt.Errorf("unzip: %w", err)
		}
		defer rdr.Close()

		out, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		defer out.Close()

		_, err = io.Copy(out, rdr)
		if err != nil {
			return err
		}
	}

	return nil
}

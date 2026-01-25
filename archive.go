package main

import (
	"archive/tar"
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ulikunitz/xz"
)

func extractArchive(archive, destination string) {
	if strings.HasSuffix(archive, ".zip") {
		extractZip(archive, destination)
	} else if strings.HasSuffix(archive, ".tar.xz") {
		extractTarXz(archive, destination)
	} else {
		panic(fmt.Sprintf("unknown compression for file %q", archive))
	}
}

func extractZip(archive, destination string) {
	zipReader, err := zip.OpenReader(archive)
	if err != nil {
		panic(fmt.Sprintf("Could not open zip file %q: %s", archive, err))
	}
	defer zipReader.Close()

	for _, file := range zipReader.File {
		err := extractZipFile(file, destination)
		if err != nil {
			panic(fmt.Sprintf("Could not extract %q from %q: %s", file.Name, archive, err))
		}
	}
}

func extractZipFile(zipFile *zip.File, destination string) error {
	readCloser, err := zipFile.Open()
	if err != nil {
		return err
	}
	defer readCloser.Close()

	path := filepath.Join(destination, zipFile.Name)

	if zipFile.FileInfo().IsDir() {
		os.MkdirAll(path, zipFile.Mode())
	} else {
		os.MkdirAll(filepath.Dir(path), zipFile.Mode())
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zipFile.Mode())
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(file, readCloser)
		if err != nil {
			return err
		}
	}

	return nil
}

func extractTarXz(archive, destination string) {
	tarFile, err := os.Open(archive)
	if err != nil {
		panic(err)
	}

	xzReader, err := xz.NewReader(tarFile)
	if err != nil {
		panic(err)
	}
	tarReader := tar.NewReader(xzReader)

	for {
		tarHeader, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		extractTarFile(tarReader, tarHeader, destination)
	}
	tarFile.Close()
}

func extractTarFile(tarReader *tar.Reader, header *tar.Header, destination string) {
	path := filepath.Join(destination, header.Name)

	switch header.Typeflag {
	case tar.TypeDir:
		if err := os.MkdirAll(path, header.FileInfo().Mode()); err != nil {
			panic(err)
		}
	case tar.TypeReg:
		file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, header.FileInfo().Mode())
		if err != nil {
			panic(err)
		}
		defer file.Close()

		if _, err := io.Copy(file, tarReader); err != nil {
			panic(err)
		}
	}
}

package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

func getLocalZigBin() string {
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		goPath = filepath.Join(os.Getenv("HOME"), "go")
	}

	modPath := filepath.Join(goPath, "pkg", "mod", "github.com", "goz")
	if _, err := os.Stat(modPath); os.IsNotExist(err) {
		if err := os.MkdirAll(modPath, os.ModePerm); err != nil {
			panic(err)
		}
	}

	zigVersion := "zig-" + getZigHost() + "-" + ZIG_VERSION
	zigPath := filepath.Join(modPath, zigVersion)
	if _, err := os.Stat(zigPath); os.IsNotExist(err) {
		archive := downloadZig(modPath, zigVersion)
		defer os.Remove(archive)
		os.RemoveAll(zigPath)
		extractArchive(archive, modPath)
	}

	return filepath.Join(zigPath, withExeExtention(runtime.GOOS, "zig"))
}

func downloadZig(modPath, zigVersion string) string {
	ext := ".tar.xz"
	if runtime.GOOS == "windows" {
		ext = ".zip"
	}
	fileName := zigVersion + ext
	url := ZIG_URL + fileName

	fmt.Printf("goz: downloading %q\n", url)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("bad response status: %s", resp.Status))
	}

	archivePath := filepath.Join(modPath, fileName)
	out, err := os.Create(archivePath)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		panic(err)
	}

	return archivePath
}

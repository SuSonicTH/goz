package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

const ZIG_VERSION = "0.15.2"
const ZIG_URL = "https://ziglang.org/download/" + ZIG_VERSION + "/"

const UPX_VERSION = "5.1.0"
const UPX_URL = "https://github.com/upx/upx/releases/download/v" + UPX_VERSION + "/"

func getLocalZigBin() string {
	version := "zig-" + getZigHost() + "-" + ZIG_VERSION
	return getLocalBin(ZIG_URL, version, "zig")
}

func getLocalUpxBin() string {
	version := "upx-" + UPX_VERSION + "-" + upxHost()
	return getLocalBin(UPX_URL, version, "upx")
}

func getLocalBin(url, version, exe string) string {
	modPath := getModPath()

	path := filepath.Join(modPath, version)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		archive := download(modPath, url, version)
		defer os.Remove(archive)
		os.RemoveAll(path)
		extractArchive(archive, modPath)
	}

	return filepath.Join(path, withExeExtention(runtime.GOOS, exe))
}

func getModPath() string {
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		goPath = filepath.Join(os.Getenv("HOME"), "go")
	}

	modPath := filepath.Join(goPath, "pkg", "mod", "github.com", "SuSonicTH", "goz")
	if _, err := os.Stat(modPath); os.IsNotExist(err) {
		if err := os.MkdirAll(modPath, os.ModePerm); err != nil {
			panic(err)
		}
	}

	return modPath
}

func download(modPath, baseUrl, version string) string {
	ext := ".tar.xz"
	if runtime.GOOS == "windows" {
		ext = ".zip"
	}
	fileName := version + ext
	url := baseUrl + fileName

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

func upxHost() string {
	switch runtime.GOOS {
	case "windows":
		return "win64"
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			return "x86_64_linunx"
		case "arm64":
			return "aarch64_linunx"
		default:
			panic(fmt.Sprintf("unsupported host architecture %q", runtime.GOARCH))
		}
	default:
		panic(fmt.Sprintf("unsupported host os %q", runtime.GOOS))
	}
}

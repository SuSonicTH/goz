package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const ZIG_VERSION = "0.15.2"
const ZIG_URL = "https://ziglang.org/download/" + ZIG_VERSION + "/"

func main() {
	setEnv()
	executeGo()
}

func executeGo() {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "go", os.Args[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func setEnv() {
	zigBin := getZigBin()
	target := getZigTarget()

	os.Setenv("CGO_ENABLED", "1")
	os.Setenv("CC", zigBin+" cc -target "+target)
	os.Setenv("CXX", zigBin+" c++ -target "+target)
}

func getZigBin() string {
	zigBin := os.Getenv("GOZBIN")
	if zigBin == "" {
		zigBin = getLocalZigBin()
	}
	return zigBin
}

func getLocalZigBin() string {
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		goPath = ".goz"
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

	exe := ""
	if runtime.GOOS == "windows" {
		exe = ".exe"
	}

	return filepath.Join(zigPath, "bin", "zig"+exe)
}

func extractArchive(archive, modPath string) {
	panic("unimplemented")
}

func downloadZig(modPath, zigVersion string) string {
	ext := ".tar.xz"
	if runtime.GOOS == "windows" {
		ext = ".zip"
	}
	fileName := zigVersion + ext
	url := ZIG_URL + fileName

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("bad response status: %s", resp.Status))
	}

	archivePath := filepath.Join(modPath, fileName)
	tempPath := archivePath + ".tmp"
	out, err := os.Create(archivePath)
	if err != nil {
		panic(err)
	}
	defer out.Close()
	defer os.Remove(tempPath)

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		panic(err)
	}

	return archivePath
}

var zigTarget *string

func getZigTarget() string {
	if zigTarget == nil {
		zt := getZigTargetArch() + "-" + getZigTargetOs()
		zigTarget = &zt
	}
	return *zigTarget
}

func getZigTargetArch() string {
	goArch := os.Getenv("GOARCH")
	if goArch == "" {
		goArch = runtime.GOARCH
	}
	return zigArchFromGoArch(goArch)
}

func getZigTargetOs() string {
	goOs := os.Getenv("GOOS")
	if goOs == "" {
		goOs = runtime.GOOS
	}
	return zigOsFromGoOs(goOs)
}

var zigHost *string

func getZigHost() string {
	if zigHost == nil {
		var zh string = getZigHostArch() + "-" + getZigHostOs()
		zigHost = &zh
	}
	return *zigHost
}

func getZigHostOs() string {
	return zigOsFromGoOs(runtime.GOOS)
}

func getZigHostArch() string {
	return zigArchFromGoArch(runtime.GOARCH)
}

func zigOsFromGoOs(goOs string) string {
	switch goOs {
	case "windows":
		return "windows"
	case "linux":
		return "linux"
	default:
		panic(fmt.Sprintf("unsupported GOOS %q", goOs))
	}
}

func zigArchFromGoArch(goArch string) string {
	switch goArch {
	case "amd64":
		return "x86_64"
	case "arm64":
		return "aarch64"
	default:
		panic(fmt.Sprintf("unsupported GOARCH %q", goArch))
	}
}

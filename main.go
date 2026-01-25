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
	"strings"
)

const ZIG_VERSION = "0.15.2"
const ZIG_URL = "https://ziglang.org/download/" + ZIG_VERSION + "/"

func main() {
	setEnv()
	execute("go", getGoArgs())
	if upx := os.Getenv("GOZ_UPX"); upx == "1" {
		//todo: check if upx is on path, if not download
		execute("upx", []string{"--lzma", getExeName()})
	}
}

func getExeName() string {
	for i, arg := range os.Args {
		if arg == "-o" && i < len(os.Args)-1 {
			return os.Args[i+1]
		}
	}
	/*todo:
		check if arguments have name.go exename=> name
	 	else check if arguments have *.go exename=> name of file with main
		else check if there is a go.mod exename=> name of module
	*/
	panic("goz: to use upx you have to set an output name with -o fileName")
}

func setEnv() {
	zigBin := getZigBin()
	target := getZigTarget()

	os.Setenv("CGO_ENABLED", "1")
	os.Setenv("CC", getZigCenv("cc", zigBin, target))
	os.Setenv("CXX", getZigCenv("c++", zigBin, target))
}

func execute(command string, arguments []string) {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, command, arguments...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func getGoArgs() []string {
	if strip := os.Getenv("GOZ_STRIP"); strip == "1" {
		goArgs := make([]string, len(os.Args)-1)
		copy(goArgs, os.Args[1:])

		goArgs = append(goArgs, "-ldflags")
		goArgs = append(goArgs, "-s -w")
		goArgs = append(goArgs, "-trimpath")
	}
	return os.Args[1:]
}

func getZigCenv(comp, zigBin, target string) string {
	env := []string{zigBin, comp, "-target", target}
	if strip := os.Getenv("GOZ_STRIP"); strip == "1" {
		env = append(env, "-Wl,-s")
	}
	return strings.Join(env, " ")
}

func getZigBin() string {
	zigBin := os.Getenv("GOZ_ZIG")
	if zigBin == "" {
		zigBin = getLocalZigBin()
	}
	return zigBin
}

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

	exe := ""
	if runtime.GOOS == "windows" {
		exe = ".exe"
	}

	return filepath.Join(zigPath, "zig"+exe)
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
	tempPath := archivePath + ".tmp"
	out, err := os.Create(tempPath)
	if err != nil {
		panic(err)
	}
	defer out.Close()
	defer os.Remove(tempPath)

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		panic(err)
	}

	if err := os.Rename(tempPath, archivePath); err != nil {
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

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

const GOZ_VERSION = "0.1.0"
const VERSION = "goz version " + GOZ_VERSION + " using zig " + ZIG_VERSION

var targetOs string

func main() {
	if len(os.Args) == 1 {
		fmt.Println(VERSION)
	} else {
		switch os.Args[1] {
		case "build":
			runCGO()
		case "install":
			runCGO()
		case "help":
			if len(os.Args) == 2 {
				fmt.Println(VERSION)
			}
		case "version":
			fmt.Println(VERSION)
		}
	}
	execute("go", os.Args[1:])
}

func runCGO() {
	setEnv()
	execute("go", getGoArgs())
	executeUpxIfSet()
	os.Exit(0)
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
	if os.Getenv("GOZ_STRIP") == "1" || os.Getenv("GOZ_SMALL") == "1" {
		goArgs := make([]string, len(os.Args)-1)
		copy(goArgs, os.Args[1:])

		goArgs = append(goArgs, "-ldflags")
		goArgs = append(goArgs, "-s -w")
		goArgs = append(goArgs, "-trimpath")
		return goArgs
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

func getZigTarget() string {
	return getZigTargetArch() + "-" + getZigTargetOs()
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
	targetOs = goOs
	return zigOsFromGoOs(goOs)
}

func getZigHost() string {
	return zigArchFromGoArch(runtime.GOARCH) + "-" + zigOsFromGoOs(runtime.GOOS)
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

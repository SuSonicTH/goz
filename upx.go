package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func executeUpxIfSet() {
	if upx := os.Getenv("GOZ_UPX"); upx == "1" {
		execute(getLocalUpxBin(), []string{"--lzma", getExeName()})
	}
}

func getExeName() string {
	for i, arg := range os.Args {
		if arg == "-o" && i < len(os.Args)-1 {
			return withPathAndExt(os.Args[i+1])
		}
	}
	for _, arg := range os.Args {
		if strings.HasSuffix(arg, ".go") && !strings.HasSuffix(arg, "_test.go") {
			name := arg[:len(arg)-3]
			return withPathAndExt(name)
		}
	}
	return getModuleName()
}

func withPathAndExt(name string) string {
	if targetOs == "windows" && !strings.HasSuffix(name, ".exe") {
		name += ".exe"
	}
	if os.Args[1] == "install" {
		return filepath.Join(getGoPath(), "bin", name)
	}
	return name
}

func getModuleName() string {
	file, err := os.ReadFile("go.mod")
	if err != nil {
		panic(err)
	}

	re := regexp.MustCompile(`(?m)^\s*module\s+(\S+)$`)
	if matches := re.FindStringSubmatch(string(file)); len(matches) == 2 {
		module := strings.Split(matches[1], "/")
		name := module[len(module)-1]
		filepath.Join()
		return withPathAndExt(name)
	}
	panic("goz: no module found, to use upx use a go.mod or set an output name with -o fileName")
}

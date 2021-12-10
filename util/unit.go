package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fsnotify/fsnotify"
)

type ScanInfo struct {
	Path  string
	IsDir bool
}

func AppPath() string {
	curPath, _ := os.Getwd()
	return curPath
}

func ScanDir(dir string, ignoreDirs []string) (list []ScanInfo, err error) {
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}
	list = append(list, ScanInfo{Path: dir, IsDir: true})

	for _, f := range fs {
		if f.IsDir() {
			if InSlice(f.Name(), ignoreDirs) {
				continue
			}
			l, err := ScanDir(filepath.Join(dir, f.Name()), ignoreDirs)
			if err != nil {
				return nil, err
			}
			list = append(list, l...)
		}
		list = append(list, ScanInfo{Path: filepath.Join(dir, f.Name()), IsDir: f.IsDir()})
	}
	return
}

func InSlice(needle string, haystacks []string) bool {
	for _, d := range haystacks {
		if needle == d {
			return true
		}
	}
	return false
}

func Replace(str, old, new string) string {
	return strings.ReplaceAll(str, old, new)
}

func IsWindows() bool {
	return runtime.GOOS == "windows"
}

func IsChildMode() bool {
	return os.Getenv("XIUSIN_RELOAD_RUN_MODE") == "child"
}

func GetChildEnv() string {
	return "XIUSIN_RELOAD_RUN_MODE=child"
}

func IsIgnoreAction(event *fsnotify.Event) bool {
	return strings.HasSuffix(event.Name, "__") || event.Op.String() == "CHMOD"
}

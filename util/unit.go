package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func AppPath() string {
	curPath, _ := os.Getwd()
	return curPath
}

func ScanDir(dir string, ignoreDirs []string) (list []struct {
	Path  string
	IsDir bool
}, err error) {
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}
	list = append(list, struct {
		Path  string
		IsDir bool
	}{Path: dir, IsDir: true})

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
		list = append(list, struct {
			Path  string
			IsDir bool
		}{Path: filepath.Join(dir, f.Name()), IsDir: f.IsDir()})
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

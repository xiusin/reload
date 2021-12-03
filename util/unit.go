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

func SnakeString(s string) string {
	data := make([]byte, 0, len(s)*2)
	j := false
	num := len(s)
	for i := 0; i < num; i++ {
		d := s[i]
		if i > 0 && d >= 'A' && d <= 'Z' && j {
			data = append(data, '_')
		}
		if d != '_' {
			j = true
		}
		data = append(data, d)
	}
	return strings.ToLower(string(data[:]))
}

func StrFirstToUpper(str string) string {
	temp := strings.Split(strings.ReplaceAll(str, "_", "-"), "-")
	var upperStr string
	for y := 0; y < len(temp); y++ {
		vv := []rune(temp[y])
		if y != 0 {
			for i := 0; i < len(vv); i++ {
				if i == 0 {
					vv[i] -= 32
					upperStr += string(vv[i])
				} else {
					upperStr += string(vv[i])
				}
			}
		}
	}
	return temp[0] + upperStr
}

func CamelString(s string) string {
	data := make([]byte, 0, len(s))
	j := false
	k := false
	num := len(s) - 1
	for i := 0; i <= num; i++ {
		d := s[i]
		if !k && d >= 'A' && d <= 'Z' {
			k = true
		}
		if d >= 'a' && d <= 'z' && (j || !k) {
			d = d - 32
			j = false
			k = true
		}
		if k && d == '_' && num > i && s[i+1] >= 'a' && s[i+1] <= 'z' {
			j = true
			continue
		}
		data = append(data, d)
	}
	return string(data[:])
}

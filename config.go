package reload

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/xiusin/reload/util"
	"gopkg.in/yaml.v3"
)

type Config struct {
	FileExts   []string `yaml:"types"`
	IgnoreDirs []string `yaml:"ignoreDirs"`
	RootDir    string   `yaml:"rootDir"`
	DelayMS    uint     `yaml:"delay"`
	Limit      uint     `yaml:"limit"`
	BuildName  string   `yaml:"tempBin"`
	RunCmd     string   `yaml:"cmd"`
}

type Conf struct {
	Cmd  *CmdConf
	File string
	conf *Config
}
type CmdConf struct {
	Envs   map[string]string
	Base   func(string) string
	Params []string
}

func (c *CmdConf) buildEnv() []string {
	envs := os.Environ()
	for k, v := range c.Envs {
		envs = append(envs, fmt.Sprintf("%s=%s", k, v))
	}
	envs = append(envs, util.GetChildEnv())
	return envs
}

var defaultConf = Conf{
	Cmd:  nil,
	File: "reload.yaml",
	conf: &Config{},
}

func exampleConf() string {
	return `
tempBin: "runtime/dev-build"
ignoreDirs: 
  - vendor
  - runtime
  - temp
  - assets
  - tmp
  - node_modules

delay: 1000
limit: 500
types: 
  - .go
  - .gohtml
  - .tpl
  
rootDir: "."`
}

func parseConf() {

	confPath := filepath.Join(util.AppPath(), defaultConf.File)
	byts, err := ioutil.ReadFile(confPath)
	if err != nil {
		byts = []byte(exampleConf())
	}
	if err := yaml.Unmarshal(byts, defaultConf.conf); err != nil {
		panic(err)
	}
}

func GetExampleConf() string {
	return exampleConf()
}

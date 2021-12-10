package reload

import (
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

type CmdConf struct {
	Envs   map[string]string
	Base   func(string) string
	Params []string
}

func (c *CmdConf) buildEnv() []string {
	envs := os.Environ()
	for k, v := range execCmdConf.Envs {
		envs = append(envs, k+"="+v)
	}
	envs = append(envs, util.GetChildEnv())
	return envs
}

var cmdConf = CmdConf{}

var conf = &Config{}

func init() {
	parseConf()
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
	confPath := filepath.Join(util.AppPath(), "reload.yaml")
	byts, err := ioutil.ReadFile(confPath)
	if err != nil {
		byts = []byte(exampleConf())
	}
	if err := yaml.Unmarshal(byts, conf); err != nil {
		panic(err)
	}
}

func GetExampleConf() string {
	return exampleConf()
}

package reload

type Config struct {
	FileExts   []string
	IgnoreDirs []string
	RootDir    string
	DelayMS    uint
	Limit      uint
	BuildName  string
	RunCmd     string
}

var defaultConf = Config{
	FileExts:   []string{".go"},                    // 参与reload的文件类型
	RootDir:    ".",                                // 扫描文件的根目录
	IgnoreDirs: []string{"vendor", "node_modules"}, // 不参与reload的目录
	DelayMS:    300,
	Limit:      500,
	BuildName:  "runtime/temp-build",
	RunCmd:     "{bin}",
}

package reload

import (
	"net/http"
	"os"
	"testing"
)

func TestReload(t *testing.T) {
	SetPrintRegisterInfo(true) // 打印监听文件
	Loop(func() error {
		return http.ListenAndServe(":8776", nil)
	}, &Conf{
		Cmd: &CmdConf{
			Params: os.Args[1:],
		}, // 命令模板
		File: "reload.yaml", // 配置文件地址
	})
}

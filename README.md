# reload

`xiusin/reload` 是一个开发期间更新组件, 旨在最小化侵入代码,无需下载其他的热更新软件

## 示例

```go
// main.go
package main

import "github.com/xiusin/reload"

func main() {
 reload.SetPrintRegisterInfo(true) // 打印监听文件
 reload.Loop(func() error {
  return http.ListenAndServe(":8776", nil)
 }, &reload.Conf{
  Cmd: &CmdConf{
   Params: os.Args[1:],
  }, // 命令模板
  File: "reload.yaml", // 配置文件地址
 })
}
```

> `reload`本身会阻塞进程, 构建一个`dev-build`文件启动调用`exec.Command`启动, 当修改文件时监测文件变化重新编译并且重启.

```shell
go run main.go
```

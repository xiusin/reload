package reload

import (
	"fmt"
	"net/http"
	"testing"
)

func TestReload(t *testing.T) {
	Loop(func() error {
		fmt.Println("启动测试服务", conf)
		return http.ListenAndServe(":8776", nil)
	})
}

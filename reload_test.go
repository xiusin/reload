package reload

import (
	"net/http"
	"testing"
)

func TestReload(t *testing.T) {
	Loop(func() error {
		return http.ListenAndServe(":8776", nil)
	}, nil)
}

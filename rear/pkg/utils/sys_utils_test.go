package utils

import (
	"testing"
)

func TestSysUtils(t *testing.T) {
	sys_u := NewSysUtils()
	sys_u.PrintSystemInfo()
}

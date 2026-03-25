//go:build !windows

package cli

import "fmt"

func RegisterFileType(_ []byte) int {
	fmt.Println("register-filetype is only supported on Windows.")
	return 0
}

package utils

import (
	"syscall"
	"unsafe"
)

var (
	user32       = syscall.NewLazyDLL("user32.dll")
	meessageBoxW = user32.NewProc("MessageBoxW")
)

func ShowDialog(title, message string) {
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	messagePtr, _ := syscall.UTF16PtrFromString(message)

	meessageBoxW.Call(
		0,
		uintptr(unsafe.Pointer(messagePtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		0x10,
	)
}

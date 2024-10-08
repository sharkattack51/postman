// +build windows

package main

import (
	"syscall"

	"golang.org/x/sys/windows"
)

func RegisterOSHandler(callback func()) {
	kernel32 := windows.NewLazySystemDLL("kernel32.dll")
	setConsoleCtrlHandler := kernel32.NewProc("SetConsoleCtrlHandler")
	setConsoleCtrlHandler.Call(
		syscall.NewCallback(func(controlType uint) uint {
			callback()
			return 0
		}), 1)
}

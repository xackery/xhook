package xhook

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	sendInputProc = user32.NewProc("SendInput")
)

type keyboardInput struct {
	wVk         uint16
	wScan       uint16
	dwFlags     uint32
	time        uint32
	dwExtraInfo uint64
}

type input struct {
	inputType uint32
	ki        keyboardInput
	padding   uint64
}

// KeyToggle presses or unpresses provided key
func KeyToggle(keys string, isPressed bool) error {
	if len(keys) < 1 {
		return fmt.Errorf("key must be provided")
	}

	var err error
	for _, k := range keys {
		var i input

		i.inputType = 1 //INPUT_KEYBOARD
		//i.ki.wVk = 0x41 // virtual key code for a
		i.ki.wScan = uint16(k)
		i.ki.dwFlags = 0 | 0x0004
		if !isPressed {
			i.ki.dwFlags = 0x0002 | 0x0004
		}

		ret, _, err := sendInputProc.Call(
			uintptr(1),
			uintptr(unsafe.Pointer(&i)),
			uintptr(unsafe.Sizeof(i)),
		)
		if err == nil {
			return nil
		}

		if errno, ok := err.(syscall.Errno); ok {
			if errno == 0 {
				return nil
			}
			return err
		}
		if ret != 0 {
			return fmt.Errorf("non-zero return")
		}
	}
	return err
}

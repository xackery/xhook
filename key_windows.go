package hook

import (
	"fmt"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const TH32CS_SNAPPROCESS = 0x00000002

var (
	user32 = syscall.NewLazyDLL("user32.dll")
	psapi  = syscall.NewLazyDLL("psapi.dll")
)

var (
	sendInputProc = user32.NewProc("SendInput")
	getFocus      = user32.NewProc("GetFocus")
	enumProcesses = psapi.NewProc("EnumProcesses")
)

var (
	hookMu         sync.RWMutex
	processes      []WindowsProcess
	processRefresh time.Time
)

type WindowsProcess struct {
	ProcessID       int
	ParentProcessID int
	Exe             string
}

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

func Title() string {

}

// KeyToggle presses or unpresses provided key
func KeyToggle(key string, isPressed bool) error {
	if len(key) < 0 {
		return fmt.Errorf("key must be provided")
	}

	var i input

	i.inputType = 1 //INPUT_KEYBOARD
	//i.ki.wVk = 0x41 // virtual key code for a
	i.ki.wScan = uint16(key[0:0])
	i.ki.dwFlags = KEYEVENTF_KEYDOWN | KEYEVENTF_UNICODE
	if !isPressed {
		i.ki.dwFlags = KEYEVENTF_KEYUP | KEYEVENTF_UNICODE
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
	return err
}

// Title returns the current active process window
func Title() string {
	ret, _, err := getFocus.Call()
	//uintptr ret
	//return HWND(ret)
	if time.Now().After(processRefresh) {
		processRefresh = time.Now().Add(10 * time.Second)
		refreshProcesses()
	}
	hookMu.RLock()
	defer hookMu.RUnlock()
	for _, p := range processes {
		if p.ProcessID == int(ret) {
			return
		}
	}

	return
}

func refreshProcesses() error {
	hookMu.Lock()
	defer hookMu.Unlock()

	handle, err := windows.CreateToolhelp32Snapshot(0x00000002, 0)
	if err != nil {
		return fmt.Errorf("CreateToolhelp32Snapshot: %w", err)
	}
	defer windows.CloseHandle(handle)

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	// get the first process
	err = windows.Process32First(handle, &entry)
	if err != nil {
		return nil, err
	}

	for {
		processes = append(processes, newWindowsProcess(&entry))

		err = windows.Process32Next(handle, &entry)
		if err != nil {
			if err == syscall.ERROR_NO_MORE_FILES {
				return nil
			}
			return fmt.Errorf("process32Next: %w", err)
		}
	}
}

func newWindowsProcess(e *windows.ProcessEntry32) WindowsProcess {
	// Find when the string ends for decoding
	end := 0
	for {
		if e.ExeFile[end] == 0 {
			break
		}
		end++
	}

	return WindowsProcess{
		ProcessID:       int(e.ProcessID),
		ParentProcessID: int(e.ParentProcessID),
		Exe:             syscall.UTF16ToString(e.ExeFile[:end]),
	}
}

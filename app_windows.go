package xhook

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	enumProcesses       = psapi.NewProc("EnumProcesses")
	getForegroundWindow = user32.NewProc("GetForegroundWindow")
	getWindowTextW      = user32.NewProc("GetWindowTextW")
	getWindowTextLength = user32.NewProc("GetWindowTextLengthW")

	processes      []WindowsProcess
	processRefresh time.Time
)

// WindowsProcess represents a windows process
type WindowsProcess struct {
	ProcessID       int
	ParentProcessID int
	Exe             string
}

// Title returns the current active process window
func Title() (string, error) {
	hwnd, err := foregroundWindow()
	if err != nil {
		return "", fmt.Errorf("foregroundWindow: %w", err)
	}

	title, err := windowText(hwnd)
	if err != nil {
		return "", fmt.Errorf("windowText: %w", err)
	}

	return title, nil
}

func windowTextLength(hwnd uintptr) (int, error) {
	ret, _, err := getWindowTextLength.Call(hwnd)
	if errno, ok := err.(syscall.Errno); ok {
		if errno == 0 {
			return int(ret), nil
		}
		return 0, err
	}
	if ret != 0 {
		return 0, fmt.Errorf("non-zero return")
	}
	return int(ret), nil
}

func windowText(hwnd uintptr) (string, error) {
	textLen, err := windowTextLength(hwnd)
	if err != nil {
		return "", fmt.Errorf("windowTextLength: %w", err)
	}
	textLen++
	buf := make([]uint16, textLen)
	ret, _, err := getWindowTextW.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(textLen))

	if errno, ok := err.(syscall.Errno); ok {
		if errno == 0 {
			return syscall.UTF16ToString(buf), nil
		}
		return "", fmt.Errorf("getWindowTextW: %w", err)
	}
	if ret != 0 {
		return "", fmt.Errorf("getWindowTextW: non-zero return")
	}

	return syscall.UTF16ToString(buf), nil
}

func foregroundWindow() (hwnd uintptr, err error) {
	ret, _, err := getForegroundWindow.Call()
	if errno, ok := err.(syscall.Errno); ok {
		if errno == 0 {
			return ret, nil
		}
		return 0, err
	}
	if ret != 0 {
		return 0, fmt.Errorf("non-zero return")
	}
	return ret, nil
}

func processCommand() {
	if time.Now().After(processRefresh) {
		processRefresh = time.Now().Add(10 * time.Second)
		refreshProcesses()
	}
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
		return fmt.Errorf("process32first: %w", err)
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

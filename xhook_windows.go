package xhook

import (
	"sync"
	"syscall"
)

var (
	user32 = syscall.NewLazyDLL("user32.dll")
	psapi  = syscall.NewLazyDLL("psapi.dll")
)

var (
	hookMu sync.RWMutex
)

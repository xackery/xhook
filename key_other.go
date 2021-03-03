// +build !windows

package xhook

import "fmt"

// KeyToggle presses or unpresses provided key
func KeyToggle(key string, isPressed bool) error {
	return fmt.Errorf("KeyToggle not supported in this OS")
}

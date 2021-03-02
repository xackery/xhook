package hook

import "fmt"

// KeyToggle presses or unpresses provided key
func KeyToggle(key string, isPressed bool) error {
	return fmt.Errorf("KeyToggle not supported in this OS")
}

// Title returns the current active process window
func Title() string {
	return ""
}

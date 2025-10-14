package utils

import "fmt"

// Simple wrapper to format errors consistently across the project.
func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}

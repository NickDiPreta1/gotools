// Package cli provides utilities for building command-line interfaces,
// including colored output and table rendering.
package cli

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// colorsEnabled indicates whether ANSI color codes should be emitted.
// This is automatically set based on whether stdout is a terminal.
var colorsEnabled bool

// ANSI escape codes for terminal colors and text styling.
const (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	Bold    = "\033[1m"
	Dim     = "\033[2m"
)

func init() {
	colorsEnabled = term.IsTerminal(int(os.Stdout.Fd()))
}

// SetColorsEnabled allows manual control over color output.
// This is useful for testing or when piping output.
func SetColorsEnabled(enabled bool) {
	colorsEnabled = enabled
}

// Colorize wraps text with the specified ANSI color code.
// If colors are disabled (e.g., non-terminal output), returns text unchanged.
func Colorize(color, text string) string {
	if !colorsEnabled {
		return text
	}
	return fmt.Sprintf("%s%s%s", color, text, Reset)
}

// Success returns text colored green, typically for success messages.
func Success(text string) string {
	return Colorize(Green, text)
}

// Error returns text colored red, typically for error messages.
func Error(text string) string {
	return Colorize(Red, text)
}

// Warning returns text colored yellow, typically for warning messages.
func Warning(text string) string {
	return Colorize(Yellow, text)
}

// Info returns text colored cyan, typically for informational messages.
func Info(text string) string {
	return Colorize(Cyan, text)
}

package cli

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

var colorsEnabled bool

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

func Colorize(color, text string) string {
	if !colorsEnabled {
		return text
	}
	return fmt.Sprintf("%s%s%s", color, text, Reset)
}
func Success(text string) string {
	return Colorize(Green, text)
}
func Error(text string) string {
	return Colorize(Red, text)
}
func Warning(text string) string {
	return Colorize(Yellow, text)
}
func Info(text string) string {
	return Colorize(Cyan, text)
}

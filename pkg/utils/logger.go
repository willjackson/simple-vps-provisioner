package utils

import (
	"fmt"
	"os"
)

// ANSI color codes for terminal output
const (
	ColorReset  = "\033[0m"
	ColorGreen  = "\033[1;32m"
	ColorYellow = "\033[1;33m"
	ColorRed    = "\033[1;31m"
	ColorCyan   = "\033[0;36m"
	ColorGray   = "\033[0;90m"
)

// Log prints a CREATE message in green
func Log(format string, args ...interface{}) {
	fmt.Printf("\n%s[CREATE] %s%s\n", ColorGreen, fmt.Sprintf(format, args...), ColorReset)
}

// Verify prints a VERIFY message in cyan
func Verify(format string, args ...interface{}) {
	fmt.Printf("%s[VERIFY] %s%s\n", ColorCyan, fmt.Sprintf(format, args...), ColorReset)
}

// Skip prints a SKIP message in gray
func Skip(format string, args ...interface{}) {
	fmt.Printf("%s[SKIP]   %s%s\n", ColorGray, fmt.Sprintf(format, args...), ColorReset)
}

// Fix prints a FIX message in yellow
func Fix(format string, args ...interface{}) {
	fmt.Printf("%s[FIX]    %s%s\n", ColorYellow, fmt.Sprintf(format, args...), ColorReset)
}

// Warn prints a warning message in yellow
func Warn(format string, args ...interface{}) {
	fmt.Printf("\n%s[!] %s%s\n", ColorYellow, fmt.Sprintf(format, args...), ColorReset)
}

// Err prints an error message in red to stderr
func Err(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "\n%s[-] %s%s\n", ColorRed, fmt.Sprintf(format, args...), ColorReset)
}

// Ok prints a success message with checkmark in green
func Ok(format string, args ...interface{}) {
	fmt.Printf("%s[✓] %s%s\n", ColorGreen, fmt.Sprintf(format, args...), ColorReset)
}

// Fail prints a failure message with X in red
func Fail(format string, args ...interface{}) {
	fmt.Printf("%s[✗] %s%s\n", ColorRed, fmt.Sprintf(format, args...), ColorReset)
}

// Section prints a section header
func Section(title string) {
	fmt.Printf("\n=== %s ===\n", title)
}

package main

import "fmt"

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[0;31m"
	colorGreen  = "\033[0;32m"
	colorYellow = "\033[1;33m"
	colorBlue   = "\033[0;34m"
)

var totalPhases = 6

func phaseLog(phase int, msg string) {
	fmt.Printf("%s[%d/%d]%s %s\n", colorGreen, phase, totalPhases, colorReset, msg)
}

func stepLog(msg string) {
	fmt.Printf("  %s->%s %s\n", colorBlue, colorReset, msg)
}

func warnLog(msg string) {
	fmt.Printf("%s[!]%s %s\n", colorYellow, colorReset, msg)
}

func errLog(msg string) {
	fmt.Printf("%s[ERROR]%s %s\n", colorRed, colorReset, msg)
}

func successLog(msg string) {
	fmt.Printf("%s[ok]%s %s\n", colorGreen, colorReset, msg)
}

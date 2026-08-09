package main

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[0;31m"
	colorGreen  = "\033[0;32m"
	colorYellow = "\033[1;33m"
	colorBlue   = "\033[0;34m"
)

var totalPhases = 6

func phaseLog(phase int, msg string) {
	installerOutput.diagnosticf("%s[%d/%d]%s %s\n", color(colorGreen), phase, totalPhases, color(colorReset), msg)
}

func stepLog(msg string) {
	installerOutput.diagnosticf("  %s->%s %s\n", color(colorBlue), color(colorReset), msg)
}

func warnLog(msg string) {
	installerOutput.diagnosticf("%s[!]%s %s\n", color(colorYellow), color(colorReset), msg)
}

func errLog(msg string) {
	installerOutput.diagnosticf("%s[ERROR]%s %s\n", color(colorRed), color(colorReset), msg)
}

func successLog(msg string) {
	installerOutput.diagnosticf("%s[ok]%s %s\n", color(colorGreen), color(colorReset), msg)
}

func color(code string) string {
	if !installerOutput.color {
		return ""
	}
	return code
}

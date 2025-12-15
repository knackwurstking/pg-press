package env

import (
	"log"
	"os"
)

const (
	LoggerFlags = log.LstdFlags | log.Lshortfile

	ANSIReset         string = "\u001b[0m"
	ANSIBold          string = "\u001b[1m"
	ANSIDim           string = "\u001b[2m"
	ANSIItalic        string = "\u001b[3m"
	ANSIUnderline     string = "\u001b[4m"
	ANSIBlink         string = "\u001b[5m"
	ANSINegative      string = "\u001b[7m"
	ANSIStrikethrough string = "\u001b[9m"

	ANSIBlack   string = "\u001b[30m"
	ANSIRed     string = "\u001b[31m"
	ANSIGreen   string = "\u001b[32m"
	ANSIYellow  string = "\u001b[33m"
	ANSIBlue    string = "\u001b[34m"
	ANSIMagenta string = "\u001b[35m"
	ANSICyan    string = "\u001b[36m"
	ANSIWhite   string = "\u001b[37m"

	ANSIBgBlack   string = "\u001b[40m"
	ANSIBgRed     string = "\u001b[41m"
	ANSIBgGreen   string = "\u001b[42m"
	ANSIBgYellow  string = "\u001b[43m"
	ANSIBgBlue    string = "\u001b[44m"
	ANSIBgMagenta string = "\u001b[45m"
	ANSIBgCyan    string = "\u001b[46m"
	ANSIBgWhite   string = "\u001b[47m"

	ANSIMiddleware = ANSIRed
	ANSIService    = ANSIGreen
	ANSIHandler    = ANSIBlue
	ANSIVerbose    = ANSIDim + ANSIItalic
)

func init() {
	log.Default().SetFlags(LoggerFlags)
}

func NewLogger(prefix string) *log.Logger {
	return log.New(os.Stderr, prefix, LoggerFlags)
}

package main

// HowTo: (this will only work if the server is prefixed to "pg-press")
//
// First create a server log file from the journalctl command
// `journalctl --user -u pg-press --output cat --no-tail > scripts/server/server.log`
//
// Second run that thing
// `go run scripts/server/main.go -log scripts/server/server.log`

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	ExitCodeGeneric = 1
)

type User struct {
	TelegramID int64
	Name       string
}

func (u *User) String() string {
	return fmt.Sprintf("%s", u.Name)
}

func (u *User) Validate() error {
	if u.TelegramID == 0 {
		return fmt.Errorf("missing TelegramID")
	}
	if u.Name == "" {
		return fmt.Errorf("missing Name")
	}
	return nil
}

type ParseLogLineOptions struct {
	HideGetMethods   bool
	IncludeHTMXPaths bool
	Debug            bool
}

type ReportOptions struct {
	*ParseLogLineOptions
	Debug bool
}

type ParsedLogLineResult struct {
	User       *User
	ServerPath string
}

func (pll *ParsedLogLineResult) String() string {
	return fmt.Sprintf("%s; %s", pll.ServerPath, pll.User.String())
}

// This is just a small "script" to parse a server log file and create a activity report
// `go run main.go -log=server.log`

func main() {
	var logFile string
	var debug bool
	var includeHTMXPaths bool
	var hideGetMethods bool

	flag.BoolVar(&debug, "debug", false, "enable debug mode")
	flag.StringVar(&logFile, "log", "", "path to the server log file")
	flag.BoolVar(&includeHTMXPaths, "htmx", false, "include HTMX paths in report")
	flag.BoolVar(&hideGetMethods, "hide-get", false, "hide GET methods in report")
	flag.Parse()

	if logFile == "" {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to provide log file path, use: -log=path/to/logfile\n")
		os.Exit(ExitCodeGeneric)
	}

	log, err := os.Open(logFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to open log file: %v\n", err)
		os.Exit(ExitCodeGeneric)
	}
	defer log.Close()

	o := ReportOptions{
		ParseLogLineOptions: &ParseLogLineOptions{
			HideGetMethods:   hideGetMethods,
			IncludeHTMXPaths: includeHTMXPaths,
			Debug:            debug,
		},
		Debug: debug,
	}

	if err = report(log, o); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to parse log file: %v\n", err)
		os.Exit(ExitCodeGeneric)
	}
}

func report(log *os.File, o ReportOptions) error {
	report := make(map[string]int)

	scanner := bufio.NewScanner(log)

	// Read log line by line and do some parsing
	for scanner.Scan() {
		text := scanner.Text()

		parsedLogLine := parseLogLine(text, o.ParseLogLineOptions)
		if parsedLogLine == nil {
			continue
		}

		// Parse the line and do something with it
		report[parsedLogLine.User.String()]++

		if o.Debug {
			fmt.Printf("[DEBUG] Parsed log line: %s\n", parsedLogLine.String())
		}
	}

	printReport(report)
	return nil
}

func printReport(report map[string]int) {
	// Get spacing
	var maxSpacing int
	for user, _ := range report {
		spacing := len(user)
		if spacing > maxSpacing {
			maxSpacing = spacing
		}
	}

	// Header
	fmt.Printf("\033[4mUser_Name%s => Count\033[0m\n", strings.Repeat(" ", maxSpacing-len("User_Name")))

	// Convert map to string sorted via count
	var users []string
	for user := range report {
		users = append(users, user)
	}
	sort.Slice(users, func(i, j int) bool {
		return report[users[i]] > report[users[j]]
	})

	// Body
	for _, user := range users {
		fmt.Printf("%s%s => %d\n", user, strings.Repeat(" ", maxSpacing-len(user)), report[user])
	}
}

// parseLogLine parses a log line and returns a ParsedLogLine object. (can be nil)
var firstServerLogLine = false

func parseLogLine(text string, o *ParseLogLineOptions) *ParsedLogLineResult {
	// Example log line:
	// `2025/10/24 13:29:37 [Server] \x1b[42m\x1b[1m200\x1b[0m \x1b[34m\x1b[1mGET    \x1b[0m \x1b[4m\x1b[36m/pg-press/htmx/nav/feed-counter\x1b[0m (\x1b[37m61.8.145.9\x1b[0m) \x1b[1m\x1b[35m17.204537118s\x1b[0m User{ID: 284727649, Name: Celle [has API key`

	// Find the server logs
	if !matchString(`\[Server\]\s.*[1-5][0-9][0-9].*`, text) {
		return nil
	}

	if o.HideGetMethods {
		if matchString(`GET`, text) {
			return nil
		}
	}

	// Find the server path from this line with the server path prefix "pg-press"
	// TODO: the hardcoded server path prefix is a problem, but for now it will do the job
	serverPath := findString(`\/pg-press\/[0-9a-zA-Z\/\-\.]+`, text)
	if serverPath == "" {
		return nil
	}

	if !o.IncludeHTMXPaths {
		// Filter out /htmx paths
		if matchString(`\/htmx`, serverPath) {
			return nil
		}
	}

	// Filter out paths containing dots
	if matchString(`\.`, serverPath) {
		return nil
	}

	// Get the user info from the log line
	user := parseUserFromLogLine(text)
	if user == nil {
		return nil
	}
	if err := user.Validate(); err != nil {
		return nil
	}

	if !firstServerLogLine {
		firstServerLogLine = true

		timeString := findString(
			`[0-9][0-9][0-9][0-9]/[0-9][0-9]/[0-9][0-9] [0-9][0-9]:[0-9][0-9]:[0-9][0-9]`,
			text,
		)

		fmt.Fprintf(os.Stderr, "[INFO] Since: %s\n", timeString)
	}

	return &ParsedLogLineResult{
		ServerPath: serverPath,
		User:       user,
	}
}

func parseUserFromLogLine(logLine string) *User {
	userMatch := findString(`User\{ID: ([0-9]+), Name: (.+)\}`, logLine)
	if userMatch == "" {
		return nil
	}

	// Find the Telegram ID
	idMatch := findString(`([0-9]+)`, userMatch)
	if idMatch == "" {
		return nil
	}

	telegramID, err := strconv.ParseInt(idMatch, 10, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to parse Telegram ID: %#v\n", err)
		return nil
	}

	// Find the user name
	userNameMatch := findString(`(Name: (.+))`, userMatch)
	if userNameMatch == "" {
		return nil
	}

	// Remove the pre. and suffix
	userName := strings.TrimPrefix(
		strings.TrimSuffix(userNameMatch, findString(` \[.*\]}$`, userNameMatch)),
		"Name: ",
	)

	return &User{
		TelegramID: telegramID,
		Name:       userName,
	}
}

func matchString(regex, text string) bool {
	r, err := regexp.Compile(regex)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to compile regex for match string: %#v\n", err)
		return false
	}

	// Check with the regex the text content if it contains the server keyword
	return r.MatchString(text)
}

//func findAllString(regex, text string) []string {
//	r, err := regexp.Compile(regex)
//	if err != nil {
//		fmt.Fprintf(os.Stderr, "[ERROR] Failed to compile regex for find all strings: %#v\n", err)
//		return nil
//	}
//
//	// Get matching strings
//	return r.FindAllString(text, -1)
//}

func findString(regex, text string) string {
	r, err := regexp.Compile(regex)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to compile regex for find string: %#v\n", err)
		return ""
	}

	// Get matching strings
	return r.FindString(text)
}

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
	"strconv"
	"strings"
)

const (
	ExitCodeGeneric = 1
)

type User struct {
	TelegramID int
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

type ParsedLogLine struct {
	User       *User
	ServerPath string
}

func (pll *ParsedLogLine) String() string {
	return fmt.Sprintf("%s; %s", pll.ServerPath, pll.User.String())
}

// This is jus a small "script" to parse a server log file and create a activity report
// `go run main.go -log=server.log`

func main() {
	var logFile string

	flag.StringVar(&logFile, "log", "", "path to the server log file")
	flag.Parse()

	if logFile == "" {
		fmt.Fprintf(os.Stderr, "Failed to provide log file path, use: -log=path/to/logfile\n")
		os.Exit(ExitCodeGeneric)
	}

	log, err := os.Open(logFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file: %v\n", err)
		os.Exit(ExitCodeGeneric)
	}
	defer log.Close()

	if err = report(log); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse log file: %v\n", err)
		os.Exit(ExitCodeGeneric)
	}
}

func report(log *os.File) error {
	report := make(map[string]int)

	scanner := bufio.NewScanner(log)

	// Read log line by line and do some parsing
	for scanner.Scan() {
		text := scanner.Text()

		parsedLogLine, err := parseLogLine(text)
		if err != nil {
			return err
		}

		// Parse the line and do something with it
		if parsedLogLine != nil {
			report[parsedLogLine.User.String()]++
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

	// Body
	for user, count := range report {
		fmt.Printf("%s%s => %d\n", user, strings.Repeat(" ", maxSpacing-len(user)), count)
	}
}

// parseLogLine parses a log line and returns a ParsedLogLine object. (can be nil)
func parseLogLine(text string) (*ParsedLogLine, error) {
	// Example log line:
	// `2025/10/24 13:29:37 [Server] \x1b[42m\x1b[1m200\x1b[0m \x1b[34m\x1b[1mGET    \x1b[0m \x1b[4m\x1b[36m/pg-press/htmx/nav/feed-counter\x1b[0m (\x1b[37m61.8.145.9\x1b[0m) \x1b[1m\x1b[35m17.204537118s\x1b[0m User{ID: 284727649, Name: Celle [has API key`

	// Find the server logs
	regex, err := regexp.Compile(`\[Server\]\s.*[1-5][0-9][0-9].*`)
	if err != nil {
		return nil, err
	}

	// Check with the regex the text content if it contains the server keyword
	if !regex.MatchString(text) {
		return nil, nil
	}

	// Find the server path from this line with the server path prefix "pg-press"
	// TODO: the hardcoded server path prefix is a problem, but for now it will do the job
	regex, err = regexp.Compile(`\/pg-press\/[0-9a-zA-Z\/\-\.]+`)
	if err != nil {
		return nil, err
	}

	// Get matching strings
	matches := regex.FindAllString(text, -1)
	if len(matches) != 1 {
		return nil, nil
	}

	serverPath := matches[0]

	// Filter out /htmx paths and paths containing dots
	regex, err = regexp.Compile(`\/htmx|\.`)
	if err != nil {
		return nil, err
	}

	if regex.MatchString(serverPath) {
		return nil, nil
	}

	// Get the user info from the log line
	user, err := parseUserFromLogLine(text)
	if err != nil {
		return nil, err
	}
	if err = user.Validate(); err != nil {
		return nil, nil
	}

	return &ParsedLogLine{
		ServerPath: serverPath,
		User:       user,
	}, nil
}

func parseUserFromLogLine(logLine string) (*User, error) {
	regex, err := regexp.Compile(`User\{ID: ([0-9]+), Name: (.+)\}`)
	if err != nil {
		return nil, err
	}

	matches := regex.FindAllString(logLine, -1)
	if len(matches) != 1 {
		return &User{}, nil
	}

	// Find the Telegram ID
	regex, err = regexp.Compile(`([0-9]+)`)
	if err != nil {
		return nil, err
	}

	telegramID, err := strconv.Atoi(regex.FindString(matches[0])) // NOTE: Maybe i should parse string to int64?
	if err != nil {
		return nil, err
	}

	// Find the user name
	regex, err = regexp.Compile(`(Name: (.+))`)
	if err != nil {
		return nil, err
	}

	userName := regex.FindString(matches[0])
	if userName == "" {
		return nil, fmt.Errorf("invalid user name: %s", matches[0])
	}

	// Remove the pre. and suffix
	// First get the suffix to remove
	regex, err = regexp.Compile(` \[.*\]}$`)
	if err != nil {
		return nil, err
	}

	prefix := "Name: "
	suffix := regex.FindString(userName)
	userName = strings.TrimPrefix(strings.TrimSuffix(userName, suffix), prefix)

	return &User{
		TelegramID: telegramID,
		Name:       userName,
	}, nil
}

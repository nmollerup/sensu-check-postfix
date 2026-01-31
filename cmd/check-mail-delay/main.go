package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/sensu/sensu-plugin-sdk/sensu"
	"github.com/sensu/sensu-go/types"
)

// Config represents the check plugin config.
type Config struct {
	sensu.PluginConfig
	Path     string
	Queue    string
	Warning  int
	Critical int
}

var (
	plugin = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "check-mail-delay",
			Short:    "Sensu check to monitor Postfix mail queue message delays",
			Keyspace: "sensu.io/plugins/check-mail-delay/config",
		},
	}

	options = []sensu.ConfigOption{
		&sensu.PluginConfigOption[string]{
			Path:      "path",
			Env:       "CHECK_MAIL_DELAY_PATH",
			Argument:  "path",
			Shorthand: "p",
			Default:   "/usr/bin/mailq",
			Usage:     "Path to the postfix mailq binary",
			Value:     &plugin.Path,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "queue",
			Env:       "CHECK_MAIL_DELAY_QUEUE",
			Argument:  "queue",
			Shorthand: "q",
			Default:   "all",
			Usage:     "The queue to check (active, deferred, hold, incoming, or all)",
			Value:     &plugin.Queue,
		},
		&sensu.PluginConfigOption[int]{
			Path:      "warning",
			Env:       "CHECK_MAIL_DELAY_WARNING",
			Argument:  "warning",
			Shorthand: "w",
			Default:   3600,
			Usage:     "Age in seconds of messages considered to be a warning",
			Value:     &plugin.Warning,
		},
		&sensu.PluginConfigOption[int]{
			Path:      "critical",
			Env:       "CHECK_MAIL_DELAY_CRITICAL",
			Argument:  "critical",
			Shorthand: "c",
			Default:   7200,
			Usage:     "Age in seconds of messages considered to be critical",
			Value:     &plugin.Critical,
		},
	}
)

func main() {
	check := sensu.NewCheck(&plugin.PluginConfig, options, checkArgs, executeCheck, false)
	check.Execute()
}

func checkArgs(event *types.Event) (int, error) {
	if plugin.Critical < 0 || plugin.Warning < 0 {
		return sensu.CheckStateUnknown, fmt.Errorf("invalid threshold values")
	}
	if plugin.Critical < plugin.Warning {
		return sensu.CheckStateUnknown, fmt.Errorf("critical threshold must be greater than or equal to warning threshold")
	}
	validQueues := map[string]bool{"all": true, "active": true, "deferred": true, "hold": true, "incoming": true}
	if !validQueues[plugin.Queue] {
		return sensu.CheckStateUnknown, fmt.Errorf("invalid queue name: %s", plugin.Queue)
	}
	return sensu.CheckStateOK, nil
}

func executeCheck(event *types.Event) (int, error) {
	oldMessages, err := getOldMessages(plugin.Path, plugin.Queue, plugin.Critical)
	if err != nil {
		return sensu.CheckStateUnknown, fmt.Errorf("failed to get old messages: %v", err)
	}

	queueName := plugin.Queue
	if queueName == "all" {
		queueName = "mail"
	}

	criticalCount := len(oldMessages)
	warningMessages, _ := getOldMessages(plugin.Path, plugin.Queue, plugin.Warning)
	warningCount := len(warningMessages)

	if criticalCount > 0 {
		msg := fmt.Sprintf("%d messages in the postfix %s queue older than %d seconds", criticalCount, queueName, plugin.Critical)
		fmt.Printf("PostfixMailDelay CRITICAL: %s\n", msg)
		return sensu.CheckStateCritical, nil
	}

	if warningCount > 0 {
		msg := fmt.Sprintf("%d messages in the postfix %s queue older than %d seconds", warningCount, queueName, plugin.Warning)
		fmt.Printf("PostfixMailDelay WARNING: %s\n", msg)
		return sensu.CheckStateWarning, nil
	}

	msg := fmt.Sprintf("0 messages in the postfix %s queue older than %d seconds", queueName, plugin.Warning)
	fmt.Printf("PostfixMailDelay OK: %s\n", msg)
	return sensu.CheckStateOK, nil
}

func getOldMessages(mailqPath string, queue string, maxAge int) ([]string, error) {
	var cmd *exec.Cmd

	if queue == "all" {
		cmd = exec.Command(mailqPath)
	} else {
		// For specific queues, use mailq with qshape or find
		cmd = exec.Command(mailqPath)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(err.Error(), "executable file not found") {
			return nil, fmt.Errorf("command not found: %s", mailqPath)
		}
	}

	outputStr := string(output)

	// Empty queue
	if strings.Contains(outputStr, "Mail queue is empty") {
		return []string{}, nil
	}

	// Parse mailq output for message ages
	oldMessages := []string{}
	scanner := bufio.NewScanner(strings.NewReader(outputStr))

	// Regex to match mailq entries with queue ID and date
	// Example: "7C8A51234  5 Tue Jan 31 15:30:00  user@example.com"
	queueLineRe := regexp.MustCompile(`^([A-F0-9]+)[\*\!]?\s+\d+\s+(\w+\s+\w+\s+\d+\s+\d+:\d+:\d+)\s+`)

	now := time.Now()

	for scanner.Scan() {
		line := scanner.Text()
		matches := queueLineRe.FindStringSubmatch(line)

		if len(matches) >= 3 {
			queueID := matches[1]
			dateStr := matches[2]

			// Parse the date - format like "Tue Jan 31 15:30:00"
			// Add current year since mailq doesn't include it
			dateWithYear := fmt.Sprintf("%s %d", dateStr, now.Year())
			msgTime, err := time.Parse("Mon Jan 2 15:04:05 2006", dateWithYear)
			if err != nil {
				continue
			}

			// Calculate age in seconds
			age := int(now.Sub(msgTime).Seconds())

			// Filter by queue if needed
			if queue != "all" {
				// Check if message is in the specified queue
				// This is a simplified check - in production might need to parse queue indicators
				if strings.Contains(line, "*") && queue != "active" {
					continue
				}
				if strings.Contains(line, "!") && queue != "hold" {
					continue
				}
			}

			if age >= maxAge {
				oldMessages = append(oldMessages, queueID)
			}
		}
	}

	return oldMessages, nil
}

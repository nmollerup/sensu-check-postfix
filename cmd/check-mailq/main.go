package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-plugin-sdk/sensu"
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
			Name:     "check-mailq",
			Short:    "Sensu check to monitor Postfix mail queue size",
			Keyspace: "sensu.io/plugins/check-mailq/config",
		},
	}

	options = []sensu.ConfigOption{
		&sensu.PluginConfigOption[string]{
			Path:      "path",
			Env:       "CHECK_MAILQ_PATH",
			Argument:  "path",
			Shorthand: "p",
			Default:   "/usr/bin/mailq",
			Usage:     "Path to the postfix mailq binary",
			Value:     &plugin.Path,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "queue",
			Env:       "CHECK_MAILQ_QUEUE",
			Argument:  "queue",
			Shorthand: "q",
			Default:   "all",
			Usage:     "The queue to check (active, deferred, hold, incoming, or all)",
			Value:     &plugin.Queue,
		},
		&sensu.PluginConfigOption[int]{
			Path:      "warning",
			Env:       "CHECK_MAILQ_WARNING",
			Argument:  "warning",
			Shorthand: "w",
			Default:   0,
			Usage:     "Number of messages in the queue considered to be a warning",
			Value:     &plugin.Warning,
		},
		&sensu.PluginConfigOption[int]{
			Path:      "critical",
			Env:       "CHECK_MAILQ_CRITICAL",
			Argument:  "critical",
			Shorthand: "c",
			Default:   0,
			Usage:     "Number of messages in the queue considered to be critical",
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
	if plugin.Critical > 0 && plugin.Warning > 0 && plugin.Critical < plugin.Warning {
		return sensu.CheckStateUnknown, fmt.Errorf("critical threshold must be greater than or equal to warning threshold")
	}
	validQueues := map[string]bool{"all": true, "active": true, "deferred": true, "hold": true, "incoming": true}
	if !validQueues[plugin.Queue] {
		return sensu.CheckStateUnknown, fmt.Errorf("invalid queue name: %s", plugin.Queue)
	}
	return sensu.CheckStateOK, nil
}

func executeCheck(event *types.Event) (int, error) {
	numMessages, err := getMailqCount(plugin.Path, plugin.Queue)
	if err != nil {
		return sensu.CheckStateUnknown, fmt.Errorf("failed to get mail queue count: %v", err)
	}

	queueName := plugin.Queue
	if queueName == "all" {
		queueName = "mail"
	}

	msg := fmt.Sprintf("%d messages in the postfix %s queue", numMessages, queueName)

	if plugin.Critical > 0 && numMessages >= plugin.Critical {
		fmt.Printf("CheckMailq CRITICAL: %s\n", msg)
		return sensu.CheckStateCritical, nil
	}

	if plugin.Warning > 0 && numMessages >= plugin.Warning {
		fmt.Printf("CheckMailq WARNING: %s\n", msg)
		return sensu.CheckStateWarning, nil
	}

	fmt.Printf("CheckMailq OK: %s\n", msg)
	return sensu.CheckStateOK, nil
}

func getMailqCount(mailqPath string, queue string) (int, error) {
	var cmd *exec.Cmd

	if queue == "all" {
		cmd = exec.Command(mailqPath)
	} else {
		cmd = exec.Command("find", fmt.Sprintf("/var/spool/postfix/%s", queue), "-type", "f")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		// mailq returns exit code 0 when queue is empty, but check if command exists
		if strings.Contains(err.Error(), "executable file not found") {
			return 0, fmt.Errorf("command not found: %s", mailqPath)
		}
	}

	outputStr := string(output)

	if queue == "all" {
		// Parse mailq output for total messages
		// Empty queue: "Mail queue is empty"
		if strings.Contains(outputStr, "Mail queue is empty") {
			return 0, nil
		}

		// Parse line like "-- 5 Kbytes in 3 Request." or "-- 0 Kbytes in 0 Requests."
		re := regexp.MustCompile(`(\d+)\s+Kbytes in\s+(\d+)\s+Request`)
		matches := re.FindStringSubmatch(outputStr)
		if len(matches) >= 3 {
			count, err := strconv.Atoi(matches[2])
			if err != nil {
				return 0, fmt.Errorf("failed to parse message count: %v", err)
			}
			return count, nil
		}

		return 0, fmt.Errorf("unable to parse mailq output")
	}

	// For specific queues, count files
	lines := strings.Split(strings.TrimSpace(outputStr), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return 0, nil
	}
	return len(lines), nil
}

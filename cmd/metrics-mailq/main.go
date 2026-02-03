package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-plugin-sdk/sensu"
)

// Config represents the metric plugin config.
type Config struct {
	sensu.PluginConfig
	Path   string
	Scheme string
}

var (
	plugin = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "metrics-mailq",
			Short:    "Sensu metric plugin to collect Postfix mail queue metrics",
			Keyspace: "sensu.io/plugins/metrics-mailq/config",
		},
	}

	options = []sensu.ConfigOption{
		&sensu.PluginConfigOption[string]{
			Path:      "path",
			Env:       "METRICS_MAILQ_PATH",
			Argument:  "path",
			Shorthand: "p",
			Default:   "/usr/bin/mailq",
			Usage:     "Path to the postfix mailq binary",
			Value:     &plugin.Path,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "scheme",
			Env:       "METRICS_MAILQ_SCHEME",
			Argument:  "scheme",
			Shorthand: "s",
			Default:   getDefaultScheme(),
			Usage:     "Metric naming scheme, text to prepend to metric",
			Value:     &plugin.Scheme,
		},
	}
)

func main() {
	metric := sensu.NewCheck(&plugin.PluginConfig, options, checkArgs, executeMetric, false)
	metric.Execute()
}

func checkArgs(event *types.Event) (int, error) {
	return sensu.CheckStateOK, nil
}

func executeMetric(event *types.Event) (int, error) {
	count, err := getMailqCount(plugin.Path)
	if err != nil {
		return sensu.CheckStateUnknown, fmt.Errorf("failed to get mail queue count: %v", err)
	}

	timestamp := time.Now().Unix()
	fmt.Printf("%s.postfixMailqCount %d %d\n", plugin.Scheme, count, timestamp)

	return sensu.CheckStateOK, nil
}

func getMailqCount(mailqPath string) (int, error) {
	cmd := exec.Command(mailqPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// mailq returns exit code 0 when queue is empty
		if strings.Contains(err.Error(), "executable file not found") {
			return 0, fmt.Errorf("command not found: %s", mailqPath)
		}
	}

	outputStr := string(output)

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

func getDefaultScheme() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	return hostname
}

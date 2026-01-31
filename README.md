# Sensu Check Postfix

[![Go Report Card](https://goreportcard.com/badge/github.com/nmollerup/sensu-check-postfix)](https://goreportcard.com/report/github.com/nmollerup/sensu-check-postfix)
[![GoDoc](https://godoc.org/github.com/nmollerup/sensu-check-postfix?status.svg)](https://godoc.org/github.com/nmollerup/sensu-check-postfix)

## Description

This plugin provides native Postfix instrumentation for monitoring and metrics collection of the mail queue via `mailq`.

The plugin includes:

- **check-mailq**: Monitor mail queue size with configurable warning and critical thresholds
- **metrics-mailq**: Collect mail queue metrics in Graphite format for time-series databases
- **check-mail-delay**: Monitor mail queue message delays with configurable age thresholds

All commands work with Postfix mail queues and use the standard `mailq` command to retrieve queue information.

## Installation

### Binary Releases

Download the latest release from the [releases page](https://github.com/nmollerup/sensu-check-postfix/releases).

### From Source

```bash
go install github.com/nmollerup/sensu-check-postfix/cmd/check-mailq@latest
go install github.com/nmollerup/sensu-check-postfix/cmd/metrics-mailq@latest
go install github.com/nmollerup/sensu-check-postfix/cmd/check-mail-delay@latest
```

## Configuration

### Asset Registration

Assets are the recommended way to install plugins in Sensu. The asset can be configured using sensuctl:

```bash
sensuctl asset add nmollerup/sensu-check-postfix
```

### Check Definition

Example check definition:

```yaml
---
type: CheckConfig
api_version: core/v2
metadata:
  name: check-mailq
spec:
  command: check-mailq --warning 200 --critical 400
  subscriptions:
    - linux
  runtime_assets:
    - nmollerup/sensu-check-postfix
  interval: 60
  publish: true
```

## Usage

### check-mailq

Monitor the size of the Postfix mail queue.

```bash
Sensu check to monitor Postfix mail queue size

Usage:
  check-mailq [flags]

Flags:
  -c, --critical int        Number of messages in the queue considered to be critical (default 0)
  -h, --help                help for check-mailq
  -p, --path string         Path to the postfix mailq binary (default "/usr/bin/mailq")
  -q, --queue string        The queue to check (active, deferred, hold, incoming, or all) (default "all")
  -w, --warning int         Number of messages in the queue considered to be a warning (default 0)
```

#### Examples

Check with default thresholds:

```bash
check-mailq --warning 200 --critical 400
```

Check a specific queue:

```bash
check-mailq --queue deferred --warning 100 --critical 200
```

Check with custom mailq path:

```bash
check-mailq --path /usr/local/bin/mailq --warning 50 --critical 100
```

Check using environment variables:

```bash
CHECK_MAILQ_WARNING=200 CHECK_MAILQ_CRITICAL=400 check-mailq
```

#### Output

The check will output one of the following states:

- **OK**: Message count is below the warning threshold
- **WARNING**: Message count is at or above the warning threshold but below the critical threshold
- **CRITICAL**: Message count is at or above the critical threshold
- **UNKNOWN**: Unable to read queue or invalid thresholds

Example outputs:

```bash
CheckMailq OK: 15 messages in the postfix mail queue
CheckMailq WARNING: 250 messages in the postfix mail queue
CheckMailq CRITICAL: 450 messages in the postfix mail queue
```

### metrics-mailq

Collect mail queue metrics in Graphite format.

```bash
Sensu metric plugin to collect Postfix mail queue metrics

Usage:
  metrics-mailq [flags]

Flags:
  -h, --help           help for metrics-mailq
  -p, --path string    Path to the postfix mailq binary (default "/usr/bin/mailq")
  -s, --scheme string  Metric naming scheme, text to prepend to metric (default: hostname)
```

#### Examples

Collect metrics with default scheme:

```bash
metrics-mailq
# Output: hostname.postfixMailqCount 15 1738367100
```

Collect metrics with custom scheme:

```bash
metrics-mailq --scheme production.mailserver01
# Output: production.mailserver01.postfixMailqCount 15 1738367100
```

Example metric check definition:

```yaml
---
type: CheckConfig
api_version: core/v2
metadata:
  name: metrics-mailq
spec:
  command: metrics-mailq
  subscriptions:
    - linux
  runtime_assets:
    - nmollerup/sensu-check-postfix
  interval: 60
  publish: true
  output_metric_format: graphite_plaintext
  output_metric_handlers:
    - influxdb
```

### check-mail-delay

Monitor message delays in the Postfix mail queue.

```bash
Sensu check to monitor Postfix mail queue message delays

Usage:
  check-mail-delay [flags]

Flags:
  -c, --critical int        Age in seconds of messages considered to be critical (default 7200)
  -h, --help                help for check-mail-delay
  -p, --path string         Path to the postfix mailq binary (default "/usr/bin/mailq")
  -q, --queue string        The queue to check (active, deferred, hold, incoming, or all) (default "all")
  -w, --warning int         Age in seconds of messages considered to be a warning (default 3600)
```

#### Examples

Check with default thresholds (warning: 1 hour, critical: 2 hours):

```bash
check-mail-delay
```

Check with custom age thresholds:

```bash
check-mail-delay --warning 1800 --critical 3600
```

Check a specific queue:

```bash
check-mail-delay --queue deferred --warning 7200 --critical 14400
```

Check using environment variables:

```bash
CHECK_MAIL_DELAY_WARNING=1800 CHECK_MAIL_DELAY_CRITICAL=3600 check-mail-delay
```

#### Output

The check will output one of the following states:

- **OK**: No messages older than the warning threshold
- **WARNING**: Messages exist that are older than the warning threshold but younger than the critical threshold
- **CRITICAL**: Messages exist that are older than the critical threshold
- **UNKNOWN**: Unable to read queue or invalid thresholds

Example outputs:

```bash
PostfixMailDelay OK: 0 messages in the postfix mail queue older than 3600 seconds
PostfixMailDelay WARNING: 5 messages in the postfix mail queue older than 3600 seconds
PostfixMailDelay CRITICAL: 10 messages in the postfix mail queue older than 7200 seconds
```

## Platform Support

This plugin is designed for Linux systems with Postfix installed. It requires:

- Postfix mail server
- `mailq` command (typically at `/usr/bin/mailq`)
- Access to Postfix queue directories (for specific queue checks)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details.

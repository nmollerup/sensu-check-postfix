# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release of sensu-check-postfix
- `check-mailq` command to monitor Postfix mail queue size
- `metrics-mailq` command to collect mail queue metrics
- `check-mail-delay` command to monitor mail queue message delays
- Support for all Postfix queues (active, deferred, hold, incoming, all)
- Configurable warning and critical thresholds
- Environment variable support for all configuration options
- Comprehensive README with usage examples

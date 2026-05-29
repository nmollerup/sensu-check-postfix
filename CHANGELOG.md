# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.1] - 2026-05-29

### Changed

- Bumped Go toolchain to 1.26.3
- Bumped `google.golang.org/grpc` from 1.59.0 to 1.79.3

### Fixed

- Applied `gofmt -s` formatting to all Go source files

## [0.1.0] - 2026-02-02

### Added

- Initial release of sensu-check-postfix
- `check-mailq` command to monitor Postfix mail queue size
- `metrics-mailq` command to collect mail queue metrics
- `check-mail-delay` command to monitor mail queue message delays
- Support for all Postfix queues (active, deferred, hold, incoming, all)
- Configurable warning and critical thresholds
- Environment variable support for all configuration options
- Comprehensive README with usage examples


[Unreleased]: https://github.com/nmollerup/sensu-check-postfix/compare/0.1.1...HEAD
[0.1.1]: https://github.com/nmollerup/sensu-check-postfix/compare/0.1.0...0.1.1
[0.1.0]: https://github.com/nmollerup/sensu-check-postfix/releases/tag/0.1.0

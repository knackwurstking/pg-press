# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v0.2.2] - 2026-04-02

### Fixed

- Add missing script import "popover" to fix `SelectBox` component not opening
- Filter out already bound cassettes from the list of bindable cassettes in the tool binding `SelectBox` component

## [v0.2.1] - 2026-04-02

### Changed

- Use tint slog handler for better log formatting and output

## [v0.2.0] - 2026-04-02

### Changed

- Set tools page tab bar to full width
- Make press cycles list tool entries an anchor for a faster navigation to a specific tool page
- Make tool cycles list entries an anchor for a faster navigation to a specific press page
- Components path changed to "/internal/templates/components/PACKAGE_NAME/"
- Switch logger to slog (JSON)

## [v0.1.0] - 2026-03-29

### Added

- Initial implementation of PG Press
- Version package with v0.1.0
- Core functionality for PDF generation and processing
- Command-line interface
- Track press cycles, tools, trouble reports and add notes for tools or press

### Changed

- Initial project structure established

### Fixed

- Build issues resolved in Makefile

[unreleased]: https://github.com/knackwurstking/pg-press/compare/v0.1.0...HEAD
[v0.1.0]: https://github.com/knackwurstking/pg-press/releases/tag/v0.1.0

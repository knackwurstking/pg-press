# PG Press

A tool for PDF generation and processing.

## Installation

### Build

```bash
make    # Build the project, the executable will be located in the `bin` directory.
```

### MacOS

```bash
make macos-install
```

## Usage

Run `pg-press --help` for more information.

## TODO

- [x] Make press cycles list tool entries an anchor for a faster navigation to a specific tool page [v0.2.0]
- [x] Make tool cycles list entries an anchor for a faster navigation to a specific press page [v0.2.0]
- [x] Make the tools tab bar full width [v0.2.0]
- [x] Switch logger to slog (JSON) [v0.2.0]
- [-] Improve project structure, move all templ components and pages to /internal/templates [v0.2.0]
  - Move all components/\* to a separate components package "/internal/templates/components/PACKAGE_NAME/"
  - Move all handler templates to "/internal/templates/pages/"
- [ ] Add share for press cycles list, maybe a section action button [v0.3.0]
- [ ] Implement an global alert system, like in picow-led v0.1.1, so the hx-on::response-error stuff can be removed [v0.3.0]

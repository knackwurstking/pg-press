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

### v0.3.0
- [ ] Move all handler templates to "/internal/templates/pages/"
- [ ] Update to templui v1.9, for this to work i need to change all dialog templates

### v0.4.0
- [ ] Implement an global alert system, like in picow-led v0.1.1, so the hx-on::response-error stuff can be removed
- [ ] Add share for press cycles list, maybe a section action button

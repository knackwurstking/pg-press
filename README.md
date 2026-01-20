# PG Press

PG Press is a modern web application built with Go, featuring a comprehensive UI library and various backend services. This system is designed to manage press machines, tools, cycles, and related operations in a manufacturing environment.

## Features

- Press machine management
- Tool tracking and lifecycle management
- Cycle monitoring and reporting
- Trouble report system
- Metal sheet management
- User authentication and authorization

## Technology Stack

### Core Technologies

- **Language**: Go (1.25.3)
- **Web Framework**: Echo v4.13.4
- **Templating Engine**: A-H Templ v0.3.960
- **Database**: SQLite (via mattn/go-sqlite3 v1.14.32)
- **Authentication**: Keymaker v1.0.0
- **UI Library**: Custom UI library (Recursive font-based design)
- **Dynamic Content**: HTMX

### Project Structure

```
/pg-press/
├── cmd/              # Main application entry points
├── internal/         # Main application code
│   ├── assets/       # UI assets (CSS, JS, fonts)
│   ├── components/   # Shared UI components
│   ├── db/           # Database operations
│   ├── env/          # Environment configuration
│   ├── errors/       # Error handling
│   ├── handlers/     # HTTP request handlers
│   ├── logger/       # Logging infrastructure
│   ├── shared/       # Shared utilities and types
│   ├── urlb/         # URL building
│   └── utils/        # Utility functions
├── docs/             # Documentation
└── scripts/          # Development scripts
```

## Getting Started

### Prerequisites

- Go 1.25.3+
- SQLite database support

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/knackwurstking/pg-press.git
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Build the application:
   ```bash
   make build
   ```

## Development

### Running the Application

```bash
make dev
```

### Generating Templates

```bash
make generate
```

## Contributing

Please read [CONTRIBUTING.md](docs/CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.

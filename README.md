# Osiris

Osiris is a tool for dumping and managing control plane configurations from Kong
Gateway. It provides functionality to fetch, sanitize, and store configuration
data from Kong control planes.

## Features

- Dump complete control plane configurations
- Support for various Kong Gateway resources (services, routes, consumers,
  plugins, etc.)

## Prerequisites

- Go 1.22 or higher (based on project requirements)
- Access credentials for Kong Control Plane API

## Installation

### Building from Source

```bash
git clone https://github.com/mikefero/osiris.git
cd osiris
make build
```

The binary will be built as `osiris` in the `bin/` directory.

## Usage

### Osiris Commands

#### dump

The dump command gathers a control plane configuration, sanitizes it (if enabled),
and saves it to a file.

```bash
osiris dump [flags]
```

#### version

Display version information for the Osiris application.

```bash
osiris version
```

#### license

Display license information for the Osiris application.

```bash
osiris license
```

### Developer Commands

| Command | Description |
|---------|-------------|
| `make build` | Build the application |
| `make dump` | Run the dump command |
| `make version` | Display version information |
| `make license` | Display license information |
| `make test` | Run tests |
| `make test-coverage` | Run tests with coverage |
| `make test-no-cache` | Run tests without cache |
| `make test-no-race` | Run tests without race detector |
| `make install-tools` | Install required tools |
| `make lint` | Lint the source code |
| `make format` | Format the source code |
| `make deadcode` | Run deadcode check |
| `make go-mod-upgrade` | Upgrade go modules |
| `make help` | Display help screen with available commands |

## Configuration

Osiris uses a YAML configuration file named `osiris.yml` in the current
directory. You can also set configuration via environment variables with the
`OSIRIS_` prefix.

### Configuration Options

| Environment Variable | Configuration Key | Description |
|---------------------|-------------------|-------------|
| `OSIRIS_BASE_URL` | `base_url` | Base URL for the Kong Admin API |
| `OSIRIS_BEARER_TOKEN` | `bearer_token` | Bearer token for API authentication |
| `OSIRIS_CONTROL_PLANE_ID` | `control_plane_id` | Control plane ID for API requests |
| `OSIRIS_SANITIZE` | `sanitize` | Enable/disable sanitization of response body fields |
| `OSIRIS_OUTPUT_FILE` | `output_file` | Output file for the sanitized configuration |
| `OSIRIS_LOGGER_LEVEL` | `logger.level` | Log level (debug, info, warn, error) |
| `OSIRIS_LOGGER_FILENAME` | `logger.filename` | Log file name |
| `OSIRIS_LOGGER_RETENTION` | `logger.retention` | Number of days to retain log files |
| `OSIRIS_TIMEOUTS_TIMEOUT` | `timeouts.timeout` | General request timeout |
| `OSIRIS_TIMEOUTS_RESPONSE_HEADER` | `timeouts.response_header` | Response header timeout |

```yaml
# Base URL for the admin API
base_url: "http://localhost:3737"

# Bearer token for API authentication
bearer_token: "your-token"

# Control plane ID for API requests
control_plane_id: "4168295f-015e-4190-837e-0fcc5d72a52f"

# Enable/disable sanitization of response body fields
sanitize: true

# Output file for the sanitized configuration
output_file: "osiris.json"

# Logger configuration
logger:
  level: "info"
  filename: "osiris.log"
  retention: 7

# API request timeouts
timeouts:
  timeout: 15s
  response_header: 15s
```

## TODO Roadmap

- [ ] Implement configurable replacement values for sanitized fields to enhance
      data protection.
- [ ] Develop functionality to push configurations to newly created control
      planes.
- [ ] Transition from using control plane ID to organization ID as the primary
      operational context.
- [ ] Create automated workflows for control plane creation and deletion.
- [ ] Add support for managing and organizing control planes into logical groups.

## License

This project is licensed under the Apache License 2.0 - see the LICENSE file for
details.

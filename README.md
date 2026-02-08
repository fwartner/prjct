# prjct

A cross-platform CLI tool that creates project directory structures from YAML-configured templates. One command, consistent folders, every time.

## Why

Professionals working with asset-heavy production pipelines (video, photo, design, development) repeatedly create identical folder hierarchies. `prjct` eliminates manual folder creation, naming inconsistencies, and misplaced assets.

- No GUI, no cloud, no account required
- Single binary, zero runtime dependencies
- Fully offline, fully portable
- Configuration-driven and extensible

## Installation

### Homebrew (macOS)

```bash
brew tap fwartner/tap
brew install prjct
```

### From source

```bash
go install github.com/fwartner/prjct@latest
```

### From binary

Download the latest release for your platform from [Releases](https://github.com/fwartner/prjct/releases).

| Platform | Archive |
|----------|---------|
| macOS (Apple Silicon) | `prjct_*_darwin_arm64.tar.gz` |
| macOS (Intel) | `prjct_*_darwin_amd64.tar.gz` |
| Linux (x86_64) | `prjct_*_linux_amd64.tar.gz` |
| Linux (arm64) | `prjct_*_linux_arm64.tar.gz` |
| Windows (x86_64) | `prjct_*_windows_amd64.zip` |

## Quick Start

```bash
# 1. Install default config
prjct install

# 2. Edit config to match your paths
# macOS/Linux: ~/.config/prjct/config.yaml
# Windows:     %USERPROFILE%\.prjct\config.yaml

# 3. Create a project
prjct video "Client Commercial 2026"
```

## Usage

### Interactive mode

Run without arguments to get a template menu:

```
$ prjct
Available templates:

  [1] Video Production (video)
  [2] Photography (photo)
  [3] Software Development (dev)

Select template [1-3]: 1
Project name: Client Commercial 2026

Project created successfully!
  Template: Video Production
  Name:     Client Commercial 2026
  Path:     /Users/you/Projects/Video/Client Commercial 2026
  Folders:  31
```

### Non-interactive mode

Pass template and name as arguments:

```bash
prjct video "Client Commercial 2026"
prjct photo "Product Shoot Q1"
prjct dev "api-gateway"
```

### Searching projects

Every project you create is automatically indexed. Search by name, template, or path:

```bash
prjct search                    # list all indexed projects
prjct search "commercial"       # substring search (case-insensitive)
prjct search -t video           # filter by template ID
```

To index projects created before this feature or outside of `prjct`:

```bash
prjct reindex                   # scan all template base paths
prjct reindex -t video          # scan only one template
```

The index is stored at `~/.config/prjct/projects.json` (macOS/Linux) or `%USERPROFILE%\.prjct\projects.json` (Windows).

### Commands

| Command | Description |
|---------|-------------|
| `prjct` | Interactive project creation |
| `prjct <template> <name>` | Non-interactive creation |
| `prjct search [query]` | Search indexed projects by name, template, or path |
| `prjct search -t <id>` | Search filtered by template ID |
| `prjct reindex` | Discover existing projects from template base paths |
| `prjct list` | List available templates |
| `prjct config` | Show config file location |
| `prjct doctor` | Validate configuration |
| `prjct install` | Create default config file |
| `prjct version` | Print version information |

### Flags

| Flag | Description |
|------|-------------|
| `-v, --verbose` | Show detailed output during creation |
| `--config <path>` | Override config file location |
| `-h, --help` | Show help |

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Config file not found |
| 3 | Config syntax/validation error |
| 4 | Template not found |
| 5 | Project directory already exists |
| 6 | Permission denied |
| 7 | Directory creation failed |
| 8 | Invalid project name |
| 9 | User cancelled |

## Configuration

### Config Location

| OS | Path |
|----|------|
| macOS | `~/.config/prjct/config.yaml` |
| Linux | `~/.config/prjct/config.yaml` |
| Windows | `%USERPROFILE%\.prjct\config.yaml` |

### Config Format

```yaml
templates:
  - id: video                    # Used in CLI: prjct video "name"
    name: "Video Production"     # Displayed in interactive menu
    base_path: "~/Projects/Video" # Where projects are created
    directories:
      - name: "01_Pre-Production"
        children:
          - name: "Scripts"
          - name: "Storyboards"
      - name: "02_Production"
        children:
          - name: "Footage"
            children:
              - name: "A-Roll"
              - name: "B-Roll"
          - name: "Audio"
      - name: "03_Post-Production"
      - name: "04_Delivery"
```

### Config Rules

- Template `id` must be unique and cannot conflict with commands (`list`, `config`, `doctor`, `install`, `help`)
- `base_path` supports `~` expansion
- Base path directories are created automatically if they don't exist
- Nesting depth is limited to 20 levels

### Project Name Handling

- Spaces are preserved: `"Client Commercial 2026"` creates a folder with that exact name
- Illegal filesystem characters (`<>:"/\|?*`) are replaced with `_`
- Windows reserved names (`CON`, `PRN`, `AUX`, `NUL`, `COM1`-`COM9`, `LPT1`-`LPT9`) are blocked on all platforms
- Unicode characters are fully supported
- Maximum name length: 255 characters

## Development

### Prerequisites

- Go 1.25+

### Build

```bash
go build -o prjct .
```

### Test

```bash
go test ./...
```

### Cross-compile

```bash
GOOS=darwin GOARCH=arm64 go build -o prjct-darwin-arm64 .
GOOS=darwin GOARCH=amd64 go build -o prjct-darwin-amd64 .
GOOS=linux GOARCH=amd64 go build -o prjct-linux-amd64 .
GOOS=windows GOARCH=amd64 go build -o prjct-windows-amd64.exe .
```

## Project Structure

```
prjct/
  main.go                    # Entry point
  cmd/
    root.go                  # Root command, interactive/non-interactive modes
    install.go               # prjct install
    list.go                  # prjct list
    config.go                # prjct config
    doctor.go                # prjct doctor
    search.go                # prjct search
    reindex.go               # prjct reindex
  internal/
    config/                  # YAML config loading, validation, path resolution
    index/                   # Project index (JSON persistence, search)
    project/                 # Directory creation, name sanitization
```

## License

MIT License. See [LICENSE](LICENSE) for details.

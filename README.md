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

### Dry run

Preview what would be created without making any changes:

```bash
prjct --dry-run video "Test Project"
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
| `prjct search [query]` | Search indexed projects |
| `prjct search -t <id>` | Search filtered by template ID |
| `prjct reindex` | Discover existing projects from template base paths |
| `prjct list` | List available templates |
| `prjct tree <template-id>` | Preview template directory structure as ASCII tree |
| `prjct open <query>` | Open a project in the file manager |
| `prjct open --terminal <query>` | Open a project in a terminal |
| `prjct path <query>` | Print matching project path (for scripting) |
| `prjct recent [n]` | Show recently created projects (default: 10) |
| `prjct stats` | Show project statistics grouped by template |
| `prjct rename <query> <new-name>` | Rename a project on disk and in the index |
| `prjct archive <query>` | Archive a project as `.tar.gz` |
| `prjct diff <template-id> <path>` | Compare project directories against template |
| `prjct export <template-id>` | Export a template to a standalone YAML file |
| `prjct import <file>` | Import templates from a YAML file |
| `prjct init <path>` | Generate a template from an existing directory |
| `prjct config` | Show config file location |
| `prjct config --edit` | Open config file in your editor |
| `prjct doctor` | Validate configuration |
| `prjct install` | Create default config file |
| `prjct completion <shell>` | Generate shell completions (bash/zsh/fish/powershell) |
| `prjct version` | Print version information |

### Flags

| Flag | Description |
|------|-------------|
| `-v, --verbose` | Show detailed output during creation |
| `--config <path>` | Override config file location |
| `--dry-run` | Preview changes without creating anything |
| `--profile <name>` | Load `config.<name>.yaml` instead of default |
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

### Config Profiles

Use `--profile` to load alternative config files:

```bash
prjct --profile work list       # loads config.work.yaml
prjct --profile personal video "My Project"
```

### Config Format

```yaml
# Optional: preferred editor for `prjct config --edit`
# editor: "code --wait"

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

### File Templates

Create files alongside directories:

```yaml
directories:
  - name: "src"
    files:
      - name: "main.go"
        content: "package main"
      - name: ".gitkeep"
```

### Optional Directories

Mark directories as optional to prompt the user during interactive creation:

```yaml
directories:
  - name: "src"
  - name: "vendor"
    optional: true
```

### Template Variables

Define variables that are resolved during project creation. Built-in variables: `{name}`, `{date}`, `{year}`, `{month}`, `{day}`.

```yaml
templates:
  - id: video
    name: "Video Production"
    base_path: "~/Projects/Video"
    variables:
      - name: client
        prompt: "Client name"
        default: "Unknown"
    directories:
      - name: "{client}"
        children:
          - name: "Footage"
```

### Post-Creation Hooks

Run commands after project creation:

```yaml
templates:
  - id: dev
    name: "Development"
    base_path: "~/Projects"
    hooks:
      - "git init"
      - "npm init -y"
    directories:
      - name: "src"
```

### Template Inheritance

Extend a base template to avoid repeating common directories:

```yaml
templates:
  - id: base
    name: "Base"
    base_path: "~/Projects"
    directories:
      - name: "docs"
      - name: "assets"

  - id: webapp
    name: "Web App"
    base_path: "~/Projects/Web"
    extends: base
    directories:
      - name: "src"
      - name: "tests"
```

The child inherits the parent's directories, hooks, and variables. Child values override parent values for variables with the same name. Child `base_path` overrides parent if set.

### Config Rules

- Template `id` must be unique and cannot conflict with built-in commands
- `base_path` supports `~` expansion
- Base path directories are created automatically if they don't exist
- Nesting depth is limited to 20 levels
- Variable names must match `[a-zA-Z_][a-zA-Z0-9_]*`

### Editor Configuration

`prjct config --edit` opens the config file in your preferred editor.

Editor resolution order:
1. `editor` field in `config.yaml` (e.g. `editor: "code --wait"`)
2. `$VISUAL` environment variable
3. `$EDITOR` environment variable
4. Platform default: `open` (macOS), `notepad` (Windows), `vi` (Linux)

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
    config.go                # prjct config [--edit]
    doctor.go                # prjct doctor
    search.go                # prjct search
    reindex.go               # prjct reindex
    completion.go            # prjct completion
    tree.go                  # prjct tree
    open.go                  # prjct open
    path.go                  # prjct path
    recent.go                # prjct recent
    stats.go                 # prjct stats
    rename.go                # prjct rename
    archive.go               # prjct archive
    diff.go                  # prjct diff
    export.go                # prjct export
    import_cmd.go            # prjct import
    init.go                  # prjct init
  internal/
    config/                  # YAML config loading, validation, inheritance
    index/                   # Project index (JSON persistence, search, sort)
    project/                 # Directory/file creation, hooks, name sanitization
    template/                # Variable resolution engine
```

## License

MIT License. See [LICENSE](LICENSE) for details.

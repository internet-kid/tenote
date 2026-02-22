# tenote

A minimal TUI note-taking application for the terminal.

![platform](https://img.shields.io/badge/platform-linux%20%7C%20macos-lightgrey)
![license](https://img.shields.io/github/license/internet-kid/tenote)
![release](https://img.shields.io/github/v/release/internet-kid/tenote)

## Install

### macOS / Linux — one-liner

```sh
curl -sfL https://raw.githubusercontent.com/internet-kid/tenote/master/install.sh | sh
```

### Via Go

```sh
go install github.com/internet-kid/tenote/cmd/tenote@latest
```

### Manual

Download a pre-built binary for your platform from the [Releases](https://github.com/internet-kid/tenote/releases) page, extract it and place in your `PATH`.

## Usage

```sh
tenote
```

## Keybindings

### Main menu

| Key | Action |
|-----|--------|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `enter` | Select |
| `q` | Quit |

### Notes

| Key | Action |
|-----|--------|
| `↑` / `k` | Previous note |
| `↓` / `j` | Next note |
| `←` / `h` | Focus sidebar |
| `→` / `l` | Focus preview |
| `tab` | Toggle focus |
| `K` | Previous section |
| `J` | Next section |
| `n` | New note |
| `e` | Edit note |
| `d` | Move to Trash |
| `r` | Restore from Trash |
| `?` | Toggle help |
| `q` | Quit |

### Edit mode

| Key | Action |
|-----|--------|
| `ctrl+s` | Save |
| `esc` | Cancel |

### Trash

| Key | Action |
|-----|--------|
| `d` | Delete permanently |
| `r` | Restore to Notes |

## Configuration

Config file: `~/.config/tenote/config.json`

| Field | Default | Description |
|-------|---------|-------------|
| `storage_dir` | `~/.local/share/tenote` | Directory where notes are stored |

The storage directory can also be changed from the **Settings** screen inside the app.

## Data

Notes are stored as plain Markdown files (`.md`) on disk:

```
~/.local/share/tenote/
├── notes/
└── trash/
```

## Build from source

```sh
git clone https://github.com/internet-kid/tenote.git
cd tenote
go build -o tenote ./cmd/tenote
```

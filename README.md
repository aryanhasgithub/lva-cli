# lva-cli

Command-line interface for the LVA OS supervisor.

## Usage

```
lva <subcommand> <action> [flags]
```

### Global flags

| Flag | Default | Description |
|---|---|---|
| `--socket` | `/run/lva/supervisor.sock` | Supervisor Unix socket path |
| `--raw-json` | `false` | Print raw JSON output |
| `--log-level` | `warn` | Log verbosity |

All flags also accept `LVA_`-prefixed environment variables (e.g. `LVA_SOCKET`, `LVA_RAW_JSON`).

---

## Subcommands

### `lva containers`

| Command | Description |
|---|---|
| `lva containers list` | List all managed containers and their state |
| `lva containers start <name>` | Start a container |
| `lva containers stop <name>` | Stop a container |
| `lva containers restart <name>` | Restart a container |
| `lva containers state <name>` | Get current state |
| `lva containers update <name>` | Pull latest image and recreate |
| `lva containers stats <name>` | CPU and memory stats |
| `lva containers logs <name> [-n N]` | Recent log lines (default 100) |

### `lva system`

| Command | Description |
|---|---|
| `lva system health` | Supervisor and container health |
| `lva system reboot` | Reboot the host |
| `lva system poweroff` | Power off the host |
| `lva system os-update --bundle-url <url>` | Trigger RAUC OTA install |
| `lva system update-stream <container>` | Stream live update progress (SSE) |

### `lva network`

| Command | Description |
|---|---|
| `lva network info` | All interfaces with full IP info |
| `lva network interfaces` | Interface names and states |
| `lva network hostname <name>` | Set system hostname |
| `lva network dhcp <interface>` | Switch interface to DHCP |
| `lva network static --interface eth0 --address 192.168.1.10 --prefix 24 --gateway 192.168.1.1 --dns 1.1.1.1,8.8.8.8` | Set static IP |

#### `lva network wifi`

| Command | Description |
|---|---|
| `lva network wifi scan [--interface wlan0]` | Scan for visible WiFi networks, sorted by signal strength |
| `lva network wifi connect --ssid MyNetwork --password secret [--interface wlan0]` | Connect to a WPA2/WPA3 network |
| `lva network wifi connect --ssid OpenNetwork [--interface wlan0]` | Connect to an open network (omit --password) |
| `lva network wifi disconnect [--interface wlan0]` | Disconnect the active WiFi connection |

All wifi commands default `--interface` to `wlan0`.

### `lva updates`

| Command | Description |
|---|---|
| `lva updates check` | Compare local vs remote versions |
| `lva updates versions` | Show local OCI version labels |
| `lva updates os` | Check for new LVA OS bundle |
| `lva updates apply <component>` | Pull latest image for a component |

### `lva audio`

| Command | Description |
|---|---|
| `lva audio devices` | List available audio input/output devices |

---

## Shell completion

```bash
# bash
lva completion bash > /etc/bash_completion.d/lva

# zsh
lva completion zsh > "${fpath[1]}/_lva"

# fish
lva completion fish > ~/.config/fish/completions/lva.fish
```

---

## Building

```bash
# Native
go mod tidy
CGO_ENABLED=0 go build -ldflags="-s -w" -o lva .

# Cross-compile for aarch64
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o lva_arm64 .
```

### Docker (multi-arch)

```bash
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --build-arg BUILD_VERSION=0.1.0 \
  -t ghcr.io/aryanhasgithub/lva-cli:0.1.0 \
  --push .
```

---

## Container structure

The container uses the same `ghcr.io/home-assistant/base` image as HA's plugin-cli,
which ships s6-overlay v3. s6 bootstraps and hands off to `cli.sh` as the main
process. `cli.sh` execs into `rlwrap lva banner` which presents the banner and
runs the interactive `lva >` REPL. When the user exits the container stops.

```
rootfs/
└── usr/
    └── bin/
        └── cli.sh    ← execs into rlwrap lva banner
```
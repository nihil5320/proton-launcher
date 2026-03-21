# proton-launcher

A command-line tool for launching Windows executables via Proton on Linux. Integrates with KDE Dolphin's right-click context menu and supports double-click launching by registering as a MIME handler for `.exe` files.

No persistent GUI runs in the background. The binary is invoked per action (launch, configure, create shortcut) and exits when done.

## Requirements

- Go 1.21+
- Proton (GE-Proton, Valve Proton via Steam, or Lutris Wine runners)
- KDE Plasma (for service menu integration)
- Optional: MangoHud, Gamescope, GameMode

## Installation

```sh
git clone https://github.com/nihil5320/proton-launcher.git
cd proton-launcher
make install
```

This installs:

- `~/.local/bin/proton-launcher` (binary)
- `~/.local/share/applications/proton-launcher.desktop` (MIME handler)
- `~/.local/share/applications/proton-launcher-config.desktop` (settings entry in KDE app launcher)
- `~/.local/share/kio/servicemenus/proton-launcher-service.desktop` (Dolphin right-click menu)

Ensure `~/.local/bin` is in your `PATH`.

To set proton-launcher as the default handler for `.exe` files (enables double-click to launch):

```sh
xdg-mime default proton-launcher.desktop application/x-ms-dos-executable
```

For system-wide installation:

```sh
PREFIX=/usr sudo make install
```

## Uninstallation

```sh
make uninstall
```

## Usage

Most interaction happens through KDE's UI rather than the command line.

### Dolphin context menu

Right-click any `.exe` file in Dolphin to see three actions:

- "Launch with Proton" -- runs the game
- "Configure Proton Settings" -- opens per-game config GUI
- "Create Desktop Shortcut" -- generates a `.desktop` file in `~/.local/share/applications/`

### Double-click

After setting the MIME default (see installation), double-clicking an `.exe` in Dolphin launches it through Proton.

### KDE app launcher

"Proton Launcher Settings" appears under Settings in the KDE application menu. Opens the global configuration GUI.

### CLI

```sh
proton-launcher run <exe>            # launch a game
proton-launcher config               # open global config GUI
proton-launcher config <exe>         # open per-game config GUI
proton-launcher list                 # list discovered Proton versions
proton-launcher desktop <exe>        # create a .desktop shortcut
proton-launcher desktop <exe> -name "My Game"
```

If the first argument is a path ending in `.exe`, it is treated as `run`.

## Configuration

See [docs/configuration.md](docs/configuration.md) for the full config reference.

Configuration uses TOML with two tiers:

- Global: `~/.config/proton-launcher/config.toml`
- Per-game: `<game-dir>/.proton-launcher.toml`

Per-game settings override global settings. Fields not set in the per-game file inherit from global. A default global config is created on first launch using the newest discovered Proton version.

## Proton discovery

proton-launcher scans these locations for Proton and Wine installations:

- `~/.steam/root/compatibilitytools.d/`
- `~/.local/share/Steam/compatibilitytools.d/`
- `~/.local/share/Steam/steamapps/common/Proton*/`
- `~/.local/share/lutris/runners/wine/`
- `/usr/share/steam/compatibilitytools.d/`

Run `proton-launcher list` to see what was discovered.

## Logs

Launch output (stdout/stderr from Proton) is written to:

```text
~/.local/share/proton-launcher/logs/<game-name>.log
```

## License

MIT

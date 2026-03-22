# proton-launcher

A shell-integrated launcher for running Windows applications via Proton on Linux -- no visible third-party launcher required. It plugs directly into KDE Dolphin's right-click context menu and registers as a handler for `.exe` files, so launching a game feels native.

Nothing runs in the background. Each action (launch, configure, create shortcut) invokes the binary, does its job, and exits immediately.

## Requirements

- Go 1.26.1+
- Proton (GE-Proton, Valve Proton via Steam, or Lutris Wine runners)
- [umu-launcher](https://github.com/Open-Wine-Components/umu-launcher) (provides umu-run; can be disabled via config)
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

This removes the binary and desktop files, all config and game data (prefixes, logs):

```sh
make uninstall
```

## Usage

All day-to-day interaction happens through KDE's shell integration -- there is no launcher window to manage.

### Dolphin context menu

Right-click any `.exe` file in Dolphin to see two actions in the service menu:

- "Configure Proton Settings" -- opens per-game config GUI
- "Create Desktop Shortcut" -- generates a `.desktop` file in `~/.local/share/applications/`

If proton-launcher is set as the default handler for `.exe` files (see installation), "Launch with Proton" also appears in the "Open With" menu.

### Double-click

After setting the MIME default (see installation), double-clicking an `.exe` in Dolphin launches it through Proton.

### KDE app launcher

"Proton Launcher Settings" appears under Settings in the KDE application menu. Opens the global configuration GUI.

### CLI

The CLI exists to support the shell integration above, but can be used directly if needed:

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
- Per-game: `~/.config/proton-launcher/games/<game-name>-<hash>.toml`

Per-game configs are stored centrally, named using a slug derived from the executable name plus a short hash of the parent directory (e.g., `mygame-a1b2c3d4.toml`). Per-game fields override global fields; anything not set inherits from global. A default global config is created on first launch.

## Proton discovery

proton-launcher scans these locations for Proton and Wine installations:

- `~/.steam/root/compatibilitytools.d/`
- `~/.local/share/Steam/compatibilitytools.d/`
- `~/.local/share/Steam/steamapps/common/` (directories starting with `Proton`)
- `~/.local/share/lutris/runners/wine/` (Proton runners only)
- `/usr/share/steam/compatibilitytools.d/`
- `~/.var/app/com.valvesoftware.Steam/data/Steam/compatibilitytools.d/` (Flatpak)
- `~/.var/app/com.valvesoftware.Steam/data/Steam/steamapps/common/` (Flatpak, directories starting with `Proton`)

Run `proton-launcher list` to see what was discovered.

Discovered versions are sorted by preference: CachyOS Proton > Steam-bundled Proton (non-experimental) > GE-Proton > Experimental/Hotfix > Lutris > other. Within each tier, newer versions sort first. The first discovered version is used as the default if no global config exists. If a configured version is no longer found, the launcher falls back to the best available version.

## Logs

Launch output (stdout/stderr from Proton) is written to:

```text
~/.local/share/proton-launcher/logs/<game-name>-<hash>.log
```

The log file is overwritten on each launch (only the most recent session is kept).

## License

MIT

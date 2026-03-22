# proton-launcher

A shell-integrated launcher for running Windows applications via Proton on Linux -- no visible third-party launcher required. It integrates with your desktop environment's file manager and registers as a handler for `.exe` files, so launching a game feels native.

Nothing runs in the background. Each action (launch, configure, create launcher entry) invokes the binary, does its job, and exits immediately.

## Supported Desktop Environments

- **KDE Plasma** -- Dolphin service menu (right-click → Configure Proton / Create Launcher Entry), MIME handler, settings entry
- **GNOME** -- Nautilus scripts (right-click → Scripts), MIME handler, settings entry
- **COSMIC** -- MIME handler (Open With), settings entry in application launcher

## Requirements

- Go 1.26.1+ (build only)
- Proton (GE-Proton, Valve Proton via Steam, or Lutris Wine runners)
- [umu-launcher](https://github.com/Open-Wine-Components/umu-launcher) (provides umu-run; can be disabled via config)
- Optional: MangoHud, Gamescope, GameMode
- Optional: kdialog (KDE) or zenity (GNOME/GTK) for error dialogs

## Installation

```sh
git clone https://github.com/nihil5320/proton-launcher.git
cd proton-launcher
make install
```

This installs:

- `~/.local/bin/proton-launcher` (binary)
- `~/.local/share/applications/proton-launcher.desktop` (MIME handler)
- `~/.local/share/applications/proton-launcher-config.desktop` (settings entry)
- `~/.local/share/kio/servicemenus/proton-launcher-service.desktop` (KDE Dolphin right-click menu)
- `~/.local/share/nautilus/scripts/proton-launcher-configure` (GNOME Nautilus right-click script)
- `~/.local/share/nautilus/scripts/proton-launcher-shortcut` (GNOME Nautilus right-click script)

All DE integrations are installed regardless of which desktop you use. See below for DE specific installation.

Ensure `~/.local/bin` is in your `PATH`.

You can also install only specific DE integrations:

```sh
make build && make install-common              # binary + .desktop files only
make install-kde                               # KDE Dolphin service menu
make install-gnome                             # GNOME Nautilus scripts
```

To set proton-launcher as the default handler for `.exe` files (enables double-click to launch):

```sh
xdg-mime default proton-launcher.desktop application/x-ms-dos-executable
```

For system-wide installation:

```sh
PREFIX=/usr sudo make install
```

### Arch Linux (AUR)

A `PKGBUILD` is provided in `packaging/aur/`. To build and install:

```sh
cd packaging/aur
makepkg -si
```

### Debian / Ubuntu (.deb)

A `.deb` package can be built using [nfpm](https://nfpm.goreleaser.com/):

```sh
make build
nfpm package --packager deb --config packaging/nfpm.yaml
sudo dpkg -i proton-launcher_*.deb
```

## Uninstallation

This removes the binary, all DE integration files, config, and game data (prefixes, logs):

```sh
make uninstall
```

## Usage

All day-to-day interaction happens through your desktop's shell integration -- there is no launcher window to manage.

### KDE Plasma (Dolphin)

Right-click any `.exe` file in Dolphin to see two actions in the service menu:

- "Configure Proton Settings" -- opens per-game config GUI
- "Create Launcher Entry" -- creates a launcher entry in `~/.local/share/applications/` so the game appears in your application menu

"Proton Launcher Settings" appears under Settings in the KDE application menu.

### GNOME (Nautilus / Files)

Right-click any `.exe` file in Files → **Scripts** to see:

- "proton-launcher-configure" -- opens per-game config GUI
- "proton-launcher-shortcut" -- creates a launcher entry

The scripts only act on `.exe` files and exit silently for anything else.

"Proton Launcher Settings" appears in the GNOME application overview.

### COSMIC

COSMIC Files follows XDG standards. The MIME handler provides "Open With Proton Launcher" for `.exe` files. "Proton Launcher Settings" appears in the COSMIC application launcher.

Custom right-click actions will be added when cosmic-files supports them upstream.

### Double-click (all DEs)

After setting the MIME default (see installation), double-clicking an `.exe` launches it through Proton.

### CLI

The CLI exists to support the shell integration above, but can be used directly if needed:

```sh
proton-launcher run <exe>            # launch a game
proton-launcher config               # open global config GUI
proton-launcher config <exe>         # open per-game config GUI
proton-launcher list                 # list discovered Proton versions
proton-launcher desktop <exe>        # create a launcher entry
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

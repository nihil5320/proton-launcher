# Configuration

proton-launcher uses TOML configuration files with two tiers:

- Global: `~/.config/proton-launcher/config.toml`
- Per-game: `<game-dir>/.proton-launcher.toml` (same directory as the `.exe`)

Per-game fields override global fields. Any field not present in the per-game file inherits from the global config. A default global config is created automatically on first launch.

## Fields

| Field | Type | Default | Description |
| ----- | ---- | ------- | ----------- |
| `proton_version` | string | (auto-detected) | Name of the Proton version to use. Must match a discovered version from `proton-launcher list`. |
| `prefix_path` | string | `~/.local/share/proton-launcher/prefixes/<game-name>` | Path to the Wine/Proton prefix directory. Created automatically if it doesn't exist. Defaults to a per-game directory derived from the executable name. |
| `use_umu` | bool | `true` | Launch via `umu-run` (recommended). Provides the Steam Linux Runtime container, ProtonFixes, and proper environment setup. Set to `false` for direct Proton invocation. |
| `game_id` | string | `"umu-default"` | UMU game ID for ProtonFixes lookup. Only used when `use_umu = true`. Find game IDs at [umu-database](https://github.com/Open-Wine-Components/umu-database). |
| `locale` | string | (system default) | Locale override for the game process (`LC_ALL`). Useful for games that need a specific locale for correct text rendering, e.g. `ja_JP.UTF-8` for Japanese. |
| `launch_args` | string array | `[]` | Extra arguments passed to the executable. |
| `mangohud` | bool | `false` | Wrap the launch command with `mangohud`. Requires MangoHud to be installed. |
| `gamescope` | bool | `false` | Wrap the launch command with `gamescope`. Requires Gamescope to be installed. |
| `gamemode` | bool | `false` | Wrap the launch command with `gamemoderun`. Requires GameMode to be installed. |

### `[env]` table

Key-value pairs added to the environment when launching. Per-game env vars are merged with global env vars (per-game values win on conflict).

### `[gamescope_opts]` table

Only used when `gamescope = true`.

| Field | Type | Description |
| ----- | ---- | ----------- |
| `width` | int | Output width |
| `height` | int | Output height |
| `fullscreen` | bool | Run in fullscreen mode (`-f` flag) |

## Global config example

```toml
proton_version = "GE-Proton10-32"
use_umu = true
# game_id = "umu-default"
mangohud = false
gamescope = false
gamemode = false

[env]
# DXVK_HUD = "fps"

[gamescope_opts]
# width = 1920
# height = 1080
# fullscreen = true
```

## Per-game config example

Only include fields you want to override. Everything else inherits from global.

```toml
proton_version = "GE-Proton10-32"
prefix_path = "~/.local/share/proton-launcher/prefixes/my-game"
game_id = "umu-35140"
launch_args = ["-fullscreen", "-skipintro"]
mangohud = true

[env]
DXVK_HUD = "fps,frametime"
WINEDLLOVERRIDES = "d3d11=n"
```

## Prefix management

By default, each game gets its own isolated prefix directory derived from the executable name:

```text
~/.local/share/proton-launcher/prefixes/<game-name>/
```

For example, launching `PS.exe` uses the prefix `~/.local/share/proton-launcher/prefixes/ps/`.

To override this, set `prefix_path` in the per-game config to a custom directory. The prefix directory is created automatically on first launch.

## umu-run

By default, proton-launcher uses [umu-run](https://github.com/Open-Wine-Components/umu-launcher) to launch games. This provides:

- **Steam Linux Runtime container** (pressure-vessel/sniper) — ensures library compatibility
- **ProtonFixes** — automatic game-specific patches and workarounds
- **Proper environment setup** — `GAMEID`, `PROTONPATH`, and container mounts

This is required for modern Proton versions (especially SLR builds like `proton-cachyos-slr`). If you need direct Proton invocation (e.g., for older Wine/Proton versions), set `use_umu = false`.

To apply game-specific ProtonFixes, set `game_id` to the game's UMU ID (e.g., `umu-35140` for Batman: Arkham Asylum). Find IDs in the [umu-database](https://github.com/Open-Wine-Components/umu-database).

## Editing config

There are three ways to edit configuration:

1. Config GUI via the Dolphin context menu ("Configure Proton Settings" on a `.exe` file)
2. Config GUI via the KDE app launcher ("Proton Launcher Settings")
3. Edit the TOML files directly with a text editor

Changes take effect on the next launch.

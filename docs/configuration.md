# Configuration

proton-launcher uses TOML configuration files with two tiers:

- Global: `~/.config/proton-launcher/config.toml`
- Per-game: `<game-dir>/.proton-launcher.toml` (same directory as the `.exe`)

Per-game fields override global fields. Any field not present in the per-game file inherits from the global config. A default global config is created automatically on first launch.

## Fields

| Field | Type | Default | Description |
| ----- | ---- | ------- | ----------- |
| `proton_version` | string | (auto-detected) | Name of the Proton version to use. Must match a discovered version from `proton-launcher list`. |
| `prefix_path` | string | `~/.local/share/proton-launcher/prefixes/default` | Path to the Wine/Proton prefix directory. Created automatically if it doesn't exist. |
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
prefix_path = "~/.local/share/proton-launcher/prefixes/default"
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
launch_args = ["-fullscreen", "-skipintro"]
mangohud = true

[env]
DXVK_HUD = "fps,frametime"
WINEDLLOVERRIDES = "d3d11=n"
```

## Prefix management

The default prefix is shared across all games that don't specify their own `prefix_path`. To isolate a game, set `prefix_path` in its per-game config to a unique directory. The prefix directory is created automatically on first launch.

Recommended convention:

```text
~/.local/share/proton-launcher/prefixes/<game-name>/
```

## Editing config

There are three ways to edit configuration:

1. Config GUI via the Dolphin context menu ("Configure Proton Settings" on a `.exe` file)
2. Config GUI via the KDE app launcher ("Proton Launcher Settings")
3. Edit the TOML files directly with a text editor

Changes take effect on the next launch.

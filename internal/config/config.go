package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	ProtonVersion *string           `toml:"proton_version,omitempty"`
	PrefixPath    *string           `toml:"prefix_path,omitempty"`
	LaunchArgs    []string          `toml:"launch_args,omitempty"`
	MangoHud      *bool             `toml:"mangohud,omitempty"`
	Gamescope     *bool             `toml:"gamescope,omitempty"`
	GameMode      *bool             `toml:"gamemode,omitempty"`
	Env           map[string]string `toml:"env,omitempty"`
	GamescopeOpts *GamescopeOpts    `toml:"gamescope_opts,omitempty"`
}

type GamescopeOpts struct {
	Width      *int  `toml:"width,omitempty"`
	Height     *int  `toml:"height,omitempty"`
	Fullscreen *bool `toml:"fullscreen,omitempty"`
}

func GlobalConfigDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "proton-launcher"), nil
}

func GlobalConfigPath() (string, error) {
	dir, err := GlobalConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.toml"), nil
}

func GameConfigPath(exePath string) string {
	return filepath.Join(filepath.Dir(exePath), ".proton-launcher.toml")
}

func Load(path string) (*Config, error) {
	cfg := &Config{}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}
	return cfg, nil
}

func Save(path string, cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating config file: %w", err)
	}
	defer f.Close()
	enc := toml.NewEncoder(f)
	return enc.Encode(cfg)
}

func Resolve(exePath string) (*Config, error) {
	globalPath, err := GlobalConfigPath()
	if err != nil {
		return nil, err
	}
	global, err := Load(globalPath)
	if err != nil {
		return nil, err
	}
	game, err := Load(GameConfigPath(exePath))
	if err != nil {
		return nil, err
	}
	merged := Merge(global, game)
	applyDefaults(merged, exePath)
	return merged, nil
}

func Merge(base, override *Config) *Config {
	out := *base

	if override.ProtonVersion != nil {
		out.ProtonVersion = override.ProtonVersion
	}
	if override.PrefixPath != nil {
		out.PrefixPath = override.PrefixPath
	}
	if override.LaunchArgs != nil {
		out.LaunchArgs = override.LaunchArgs
	}
	if override.MangoHud != nil {
		out.MangoHud = override.MangoHud
	}
	if override.Gamescope != nil {
		out.Gamescope = override.Gamescope
	}
	if override.GameMode != nil {
		out.GameMode = override.GameMode
	}
	if override.Env != nil {
		merged := make(map[string]string)
		for k, v := range base.Env {
			merged[k] = v
		}
		for k, v := range override.Env {
			merged[k] = v
		}
		out.Env = merged
	}
	if override.GamescopeOpts != nil {
		if out.GamescopeOpts == nil {
			out.GamescopeOpts = override.GamescopeOpts
		} else {
			merged := *out.GamescopeOpts
			if override.GamescopeOpts.Width != nil {
				merged.Width = override.GamescopeOpts.Width
			}
			if override.GamescopeOpts.Height != nil {
				merged.Height = override.GamescopeOpts.Height
			}
			if override.GamescopeOpts.Fullscreen != nil {
				merged.Fullscreen = override.GamescopeOpts.Fullscreen
			}
			out.GamescopeOpts = &merged
		}
	}

	return &out
}

func ExpandPath(p string) string {
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return p
		}
		return filepath.Join(home, p[2:])
	}
	return p
}

func applyDefaults(cfg *Config, exePath string) {
	if cfg.MangoHud == nil {
		cfg.MangoHud = BoolPtr(false)
	}
	if cfg.Gamescope == nil {
		cfg.Gamescope = BoolPtr(false)
	}
	if cfg.GameMode == nil {
		cfg.GameMode = BoolPtr(false)
	}
	if cfg.PrefixPath == nil {
		dataDir, err := os.UserHomeDir()
		if err == nil {
			def := filepath.Join(dataDir, ".local", "share", "proton-launcher", "prefixes", "default")
			cfg.PrefixPath = &def
		}
	}
	if cfg.Env == nil {
		cfg.Env = make(map[string]string)
	}
}

func StringPtr(s string) *string { return &s }
func BoolPtr(b bool) *bool       { return &b }
func IntPtr(i int) *int          { return &i }

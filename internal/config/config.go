package config

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	ProtonVersion *string           `toml:"proton_version,omitempty"`
	PrefixPath    *string           `toml:"prefix_path,omitempty"`
	UseUmu        *bool             `toml:"use_umu,omitempty"`
	GameID        *string           `toml:"game_id,omitempty"`
	Locale        *string           `toml:"locale,omitempty"`
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

func GameConfigDir() (string, error) {
	dir, err := GlobalConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "games"), nil
}

func GameConfigPath(exePath string) (string, error) {
	dir, err := GameConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, SanitizeGameName(exePath)+".toml"), nil
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
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	f, err := os.CreateTemp(dir, ".proton-launcher-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp config file: %w", err)
	}
	tmpPath := f.Name()

	enc := toml.NewEncoder(f)
	if err := enc.Encode(cfg); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("encoding config: %w", err)
	}
	if err := f.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("closing config file: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("saving config file: %w", err)
	}
	return nil
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
	gamePath, err := GameConfigPath(exePath)
	if err != nil {
		return nil, err
	}
	game, err := Load(gamePath)
	if err != nil {
		return nil, err
	}

	// Save the global prefix_path as the base directory before merging,
	// so per-game prefix_path can fully override it.
	prefixBase := global.PrefixPath

	merged := Merge(global, game)
	applyDefaults(merged)

	// If the per-game config doesn't set its own prefix_path,
	// derive a per-game prefix under the base directory.
	if game.PrefixPath == nil || *game.PrefixPath == "" {
		base, err := DefaultPrefixBase()
		if err != nil {
			return nil, err
		}
		if prefixBase != nil {
			base = ExpandPath(*prefixBase)
		}
		name := SanitizeGameName(exePath)
		merged.PrefixPath = StringPtr(filepath.Join(base, name))
	} else {
		// Per-game config set an explicit prefix_path; expand ~ if used.
		merged.PrefixPath = StringPtr(ExpandPath(*merged.PrefixPath))
	}

	return merged, nil
}

func Merge(base, override *Config) *Config {
	out := *base

	// Deep-copy slices and maps from base so mutations to the merged
	// config cannot affect the originals.
	if base.LaunchArgs != nil {
		out.LaunchArgs = make([]string, len(base.LaunchArgs))
		copy(out.LaunchArgs, base.LaunchArgs)
	}
	if base.Env != nil {
		out.Env = make(map[string]string, len(base.Env))
		for k, v := range base.Env {
			out.Env[k] = v
		}
	}

	if override.ProtonVersion != nil {
		out.ProtonVersion = override.ProtonVersion
	}
	if override.PrefixPath != nil {
		out.PrefixPath = override.PrefixPath
	}
	if override.UseUmu != nil {
		out.UseUmu = override.UseUmu
	}
	if override.GameID != nil {
		out.GameID = override.GameID
	}
	if override.Locale != nil {
		out.Locale = override.Locale
	}
	if override.LaunchArgs != nil {
		out.LaunchArgs = make([]string, len(override.LaunchArgs))
		copy(out.LaunchArgs, override.LaunchArgs)
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
		for k, v := range override.Env {
			out.Env[k] = v
		}
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

func DefaultPrefixBase() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".local", "share", "proton-launcher", "prefixes"), nil
}

// DefaultGameID is the default UMU game ID used when none is configured.
const DefaultGameID = "umu-default"

func applyDefaults(cfg *Config) {
	if cfg.UseUmu == nil {
		cfg.UseUmu = BoolPtr(true)
	}
	if cfg.GameID == nil {
		cfg.GameID = StringPtr(DefaultGameID)
	}
	if cfg.MangoHud == nil {
		cfg.MangoHud = BoolPtr(false)
	}
	if cfg.Gamescope == nil {
		cfg.Gamescope = BoolPtr(false)
	}
	if cfg.GameMode == nil {
		cfg.GameMode = BoolPtr(false)
	}
	if cfg.Env == nil {
		cfg.Env = make(map[string]string)
	}
}

var nonWordChar = regexp.MustCompile(`[^\p{L}\p{N}]+`)

func SanitizeGameName(exePath string) string {
	abs, err := filepath.Abs(exePath)
	if err != nil {
		abs = exePath
	}
	dir := filepath.Dir(abs)
	hash := sha256.Sum256([]byte(dir))
	shortHash := hex.EncodeToString(hash[:4])

	name := strings.TrimSuffix(filepath.Base(abs), filepath.Ext(abs))
	name = strings.ToLower(name)
	name = nonWordChar.ReplaceAllString(name, "-")
	name = strings.Trim(name, "-")
	if name == "" {
		return shortHash
	}
	return name + "-" + shortHash
}

// DefaultGlobalConfig returns a Config with the standard global defaults.
func DefaultGlobalConfig() *Config {
	return &Config{
		UseUmu:    BoolPtr(true),
		MangoHud:  BoolPtr(false),
		Gamescope: BoolPtr(false),
		GameMode:  BoolPtr(false),
	}
}

// ResetGlobalConfig overwrites the global config with defaults.
func ResetGlobalConfig() error {
	cfgPath, err := GlobalConfigPath()
	if err != nil {
		return err
	}
	return Save(cfgPath, DefaultGlobalConfig())
}

// DeleteGameConfig removes the per-game config file for the given executable.
func DeleteGameConfig(exePath string) error {
	p, err := GameConfigPath(exePath)
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// DeleteGamePrefix removes the Wine prefix directory for the given executable.
func DeleteGamePrefix(exePath string) error {
	cfg, err := Resolve(exePath)
	if err != nil {
		return err
	}
	if cfg.PrefixPath == nil {
		return nil
	}
	p := ExpandPath(*cfg.PrefixPath)
	if err := validateRemovePath(p); err != nil {
		return err
	}
	if err := os.RemoveAll(p); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// DeleteAllGameConfigs removes the entire per-game config directory.
func DeleteAllGameConfigs() error {
	dir, err := GameConfigDir()
	if err != nil {
		return err
	}
	if err := os.RemoveAll(dir); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// DeleteAllPrefixes removes the default prefix base directory.
// Custom per-game prefix paths that point elsewhere are not affected.
func DeleteAllPrefixes() error {
	globalPath, err := GlobalConfigPath()
	if err != nil {
		return err
	}
	global, err := Load(globalPath)
	if err != nil {
		return err
	}
	base, err := DefaultPrefixBase()
	if err != nil {
		return err
	}
	if global.PrefixPath != nil {
		base = ExpandPath(*global.PrefixPath)
	}
	if err := validateRemovePath(base); err != nil {
		return err
	}
	if err := os.RemoveAll(base); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func StringPtr(s string) *string { return &s }
func BoolPtr(b bool) *bool       { return &b }
func IntPtr(i int) *int          { return &i }

// validateRemovePath refuses to os.RemoveAll a path that looks dangerously
// broad (home directory, filesystem root, or fewer than 4 path components).
func validateRemovePath(p string) error {
	cleaned := filepath.Clean(p)
	home, _ := os.UserHomeDir()
	if cleaned == "/" || cleaned == "." || cleaned == home {
		return fmt.Errorf("refusing to delete dangerous path: %s", cleaned)
	}
	// Require at least 4 components, e.g. /home/user/.local/share
	parts := strings.Split(cleaned, string(filepath.Separator))
	// Remove empty parts from leading /
	var nonEmpty int
	for _, part := range parts {
		if part != "" {
			nonEmpty++
		}
	}
	if nonEmpty < 4 {
		return fmt.Errorf("refusing to delete path with fewer than 4 components: %s", cleaned)
	}
	return nil
}

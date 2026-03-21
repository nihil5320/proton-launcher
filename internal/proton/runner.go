package proton

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nihil5320/proton-launcher/internal/config"
)

// cleanEnv returns os.Environ() with all known Wine, Proton, Steam, and
// umu-launcher environment variables stripped so that stale values inherited
// from the parent process cannot interfere with prefix isolation or runner
// behaviour.
func cleanEnv() []string {
	remove := map[string]bool{
		"WINEPREFIX":                       true,
		"WINEDLLOVERRIDES":                 true,
		"WINESERVER":                       true,
		"WINELOADER":                       true,
		"WINEDEBUG":                        true,
		"WINE_LARGE_ADDRESS_AWARE":         true,
		"STEAM_COMPAT_DATA_PATH":           true,
		"STEAM_COMPAT_CLIENT_INSTALL_PATH": true,
		"STEAM_COMPAT_TOOL_PATHS":          true,
		"STEAM_COMPAT_MOUNTS":              true,
		"PROTONPATH":                       true,
		"PROTON_LOG":                       true,
		"PROTON_DUMP_DEBUG_COMMANDS":       true,
		"GAMEID":                           true,
		"UMU_ID":                           true,
		"STORE":                            true,
		"SteamAppId":                       true,
		"SteamGameId":                      true,
	}

	var out []string
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if !remove[key] {
			out = append(out, entry)
		}
	}
	return out
}

// reservedEnvKeys are environment variables that must not be overridden
// by user [env] config, as they control prefix isolation and runner behaviour.
var reservedEnvKeys = map[string]bool{
	"WINEPREFIX":                       true,
	"STEAM_COMPAT_DATA_PATH":           true,
	"PROTONPATH":                       true,
	"GAMEID":                           true,
	"STEAM_COMPAT_TOOL_PATHS":          true,
	"STEAM_COMPAT_CLIENT_INSTALL_PATH": true,
}

func Run(exePath string, cfg *config.Config) error {
	absExe, err := filepath.Abs(exePath)
	if err != nil {
		return fmt.Errorf("resolving exe path: %w", err)
	}
	if _, err := os.Stat(absExe); err != nil {
		return fmt.Errorf("exe not found: %s", absExe)
	}

	if cfg.ProtonVersion == nil || *cfg.ProtonVersion == "" {
		return fmt.Errorf("no proton version configured; run 'proton-launcher config' to set one")
	}

	version, err := FindVersion(*cfg.ProtonVersion)
	if err != nil {
		return err
	}

	prefixPath := config.ExpandPath(*cfg.PrefixPath)
	if err := os.MkdirAll(prefixPath, 0o755); err != nil {
		return fmt.Errorf("creating prefix dir: %w", err)
	}

	logDir, err := logDirectory()
	if err == nil {
		os.MkdirAll(logDir, 0o755)
	}

	var args []string
	var env []string
	var warnings []string
	if cfg.UseUmu != nil && *cfg.UseUmu {
		umuPath, err := exec.LookPath("umu-run")
		if err != nil {
			return fmt.Errorf("umu-run not found in PATH; install umu-launcher or set use_umu = false in config")
		}
		args, warnings = buildUmuCommand(umuPath, absExe, cfg)
		env = buildUmuEnv(version, prefixPath, cfg)
	} else {
		args, warnings = buildCommand(version, absExe, cfg)
		env = buildEnv(version, prefixPath, cfg)
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = filepath.Dir(absExe)
	cmd.Env = env

	logFile, logErr := openLogFile(absExe)
	if logErr == nil {
		for _, w := range warnings {
			fmt.Fprintf(logFile, "WARNING: %s\n", w)
		}
		cmd.Stdout = logFile
		cmd.Stderr = logFile
	}

	if err := cmd.Start(); err != nil {
		if logFile != nil {
			logFile.Close()
		}
		return fmt.Errorf("starting proton: %w", err)
	}

	go func() {
		waitErr := cmd.Wait()
		if logFile != nil {
			if waitErr != nil {
				fmt.Fprintf(logFile, "\n--- proton exited with error: %v\n", waitErr)
			} else {
				fmt.Fprintf(logFile, "\n--- proton exited successfully\n")
			}
			logFile.Close()
		}
	}()

	return nil
}

func buildWrapperArgs(cfg *config.Config) ([]string, []string) {
	var args []string
	var warnings []string

	if cfg.GameMode != nil && *cfg.GameMode {
		if _, err := exec.LookPath("gamemoderun"); err == nil {
			args = append(args, "gamemoderun")
		} else {
			warnings = append(warnings, "gamemode enabled in config but gamemoderun not found in PATH; skipping")
		}
	}

	if cfg.MangoHud != nil && *cfg.MangoHud {
		if _, err := exec.LookPath("mangohud"); err == nil {
			args = append(args, "mangohud")
		} else {
			warnings = append(warnings, "mangohud enabled in config but not found in PATH; skipping")
		}
	}

	if cfg.Gamescope != nil && *cfg.Gamescope {
		if _, err := exec.LookPath("gamescope"); err == nil {
			gsArgs := []string{"gamescope"}
			if cfg.GamescopeOpts != nil {
				if cfg.GamescopeOpts.Width != nil {
					gsArgs = append(gsArgs, "-w", fmt.Sprintf("%d", *cfg.GamescopeOpts.Width))
				}
				if cfg.GamescopeOpts.Height != nil {
					gsArgs = append(gsArgs, "-h", fmt.Sprintf("%d", *cfg.GamescopeOpts.Height))
				}
				if cfg.GamescopeOpts.Fullscreen != nil && *cfg.GamescopeOpts.Fullscreen {
					gsArgs = append(gsArgs, "-f")
				}
			}
			gsArgs = append(gsArgs, "--")
			args = append(args, gsArgs...)
		} else {
			warnings = append(warnings, "gamescope enabled in config but not found in PATH; skipping")
		}
	}

	return args, warnings
}

func buildCommand(version Version, exePath string, cfg *config.Config) ([]string, []string) {
	args, warnings := buildWrapperArgs(cfg)

	args = append(args, version.Path, "run", exePath)
	args = append(args, cfg.LaunchArgs...)
	return args, warnings
}

func buildUmuCommand(umuPath, exePath string, cfg *config.Config) ([]string, []string) {
	args, warnings := buildWrapperArgs(cfg)
	args = append(args, umuPath, exePath)
	args = append(args, cfg.LaunchArgs...)
	return args, warnings
}

func buildUmuEnv(version Version, prefixPath string, cfg *config.Config) []string {
	env := cleanEnv()
	set := func(key, val string) {
		env = append(env, key+"="+val)
	}

	gameID := "umu-default"
	if cfg.GameID != nil && *cfg.GameID != "" {
		gameID = *cfg.GameID
	}
	set("GAMEID", gameID)
	set("PROTONPATH", filepath.Dir(version.Path))
	set("STEAM_COMPAT_DATA_PATH", prefixPath)
	set("WINEPREFIX", filepath.Join(prefixPath, "pfx"))

	if cfg.Locale != nil && *cfg.Locale != "" {
		set("LANG", *cfg.Locale)
	}

	for k, v := range cfg.Env {
		if !reservedEnvKeys[k] {
			set(k, v)
		}
	}

	return env
}

func buildEnv(version Version, prefixPath string, cfg *config.Config) []string {
	env := cleanEnv()

	protonDir := filepath.Dir(version.Path)
	set := func(key, val string) {
		env = append(env, key+"="+val)
	}

	set("STEAM_COMPAT_DATA_PATH", prefixPath)
	if root := findSteamRoot(); root != "" {
		set("STEAM_COMPAT_CLIENT_INSTALL_PATH", root)
	}
	set("WINEPREFIX", filepath.Join(prefixPath, "pfx"))
	set("STEAM_COMPAT_TOOL_PATHS", protonDir)

	if cfg.Locale != nil && *cfg.Locale != "" {
		set("LANG", *cfg.Locale)
	}

	for k, v := range cfg.Env {
		if !reservedEnvKeys[k] {
			set(k, v)
		}
	}

	return env
}

func findSteamRoot() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	candidates := []string{
		filepath.Join(home, ".steam", "root"),
		filepath.Join(home, ".local", "share", "Steam"),
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c
		}
	}
	return ""
}

func logDirectory() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "proton-launcher", "logs"), nil
}

func openLogFile(exePath string) (*os.File, error) {
	logDir, err := logDirectory()
	if err != nil {
		return nil, err
	}
	name := config.SanitizeGameName(exePath)
	logPath := filepath.Join(logDir, name+".log")
	return os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
}

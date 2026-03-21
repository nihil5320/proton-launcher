package proton

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nihil5320/proton-launcher/internal/config"
)

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
	if cfg.UseUmu != nil && *cfg.UseUmu {
		umuPath, err := exec.LookPath("umu-run")
		if err != nil {
			return fmt.Errorf("umu-run not found in PATH; install umu-launcher or set use_umu = false in config")
		}
		args = buildUmuCommand(umuPath, absExe, cfg)
		env = buildUmuEnv(version, prefixPath, cfg)
	} else {
		args = buildCommand(version, absExe, cfg)
		env = buildEnv(version, prefixPath, cfg)
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = filepath.Dir(absExe)
	cmd.Env = env

	logFile, logErr := openLogFile(absExe)
	if logErr == nil {
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
		cmd.Wait()
		if logFile != nil {
			logFile.Close()
		}
	}()

	return nil
}

func buildWrapperArgs(cfg *config.Config) []string {
	var args []string

	if cfg.GameMode != nil && *cfg.GameMode {
		if _, err := exec.LookPath("gamemoderun"); err == nil {
			args = append(args, "gamemoderun")
		}
	}

	if cfg.MangoHud != nil && *cfg.MangoHud {
		if _, err := exec.LookPath("mangohud"); err == nil {
			args = append(args, "mangohud")
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
		}
	}

	return args
}

func buildCommand(version Version, exePath string, cfg *config.Config) []string {
	args := buildWrapperArgs(cfg)

	args = append(args, version.Path, "run", exePath)
	args = append(args, cfg.LaunchArgs...)
	return args
}

func buildUmuCommand(umuPath, exePath string, cfg *config.Config) []string {
	args := buildWrapperArgs(cfg)
	args = append(args, umuPath, exePath)
	args = append(args, cfg.LaunchArgs...)
	return args
}

func buildUmuEnv(version Version, prefixPath string, cfg *config.Config) []string {
	env := os.Environ()
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

	for k, v := range cfg.Env {
		set(k, v)
	}

	return env
}

func buildEnv(version Version, prefixPath string, cfg *config.Config) []string {
	env := os.Environ()

	protonDir := filepath.Dir(version.Path)
	set := func(key, val string) {
		env = append(env, key+"="+val)
	}

	set("STEAM_COMPAT_DATA_PATH", prefixPath)
	set("STEAM_COMPAT_CLIENT_INSTALL_PATH", findSteamRoot())
	set("WINEPREFIX", filepath.Join(prefixPath, "pfx"))
	set("STEAM_COMPAT_TOOL_PATHS", protonDir)

	for k, v := range cfg.Env {
		set(k, v)
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
	name := strings.TrimSuffix(filepath.Base(exePath), filepath.Ext(exePath))
	logPath := filepath.Join(logDir, name+".log")
	return os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
}

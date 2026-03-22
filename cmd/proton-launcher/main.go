package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nihil5320/proton-launcher/internal/config"
	"github.com/nihil5320/proton-launcher/internal/desktop"
	"github.com/nihil5320/proton-launcher/internal/gui"
	"github.com/nihil5320/proton-launcher/internal/proton"
)

const usage = `Usage: proton-launcher <command> [options]

Commands:
  run <exe>                Launch a game through Proton
  config [<exe>]           Open config GUI (global if no exe given)
  list                     List discovered Proton versions
  desktop <exe>            Create a .desktop shortcut for a game

Run 'proton-launcher <command> -h' for command-specific help.
`

func main() {
	if len(os.Args) < 2 {
		// No args: open global config (for .desktop launcher entry)
		cmdConfig([]string{})
		return
	}

	switch os.Args[1] {
	case "run":
		cmdRun(os.Args[2:])
	case "config":
		cmdConfig(os.Args[2:])
	case "list":
		cmdList()
	case "desktop":
		cmdDesktop(os.Args[2:])
	case "-h", "--help", "help":
		fmt.Print(usage)
	default:
		// If the argument looks like a path to an .exe, treat it as "run"
		if strings.HasSuffix(strings.ToLower(os.Args[1]), ".exe") {
			cmdRun(os.Args[1:])
		} else {
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n%s", os.Args[1], usage)
			os.Exit(1)
		}
	}
}

func cmdRun(args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: proton-launcher run <exe> [flags]")
		fs.PrintDefaults()
	}
	fs.Parse(args)

	if fs.NArg() < 1 {
		fs.Usage()
		os.Exit(1)
	}

	exePath := fs.Arg(0)
	absExe, err := filepath.Abs(exePath)
	if err != nil {
		showError(fmt.Sprintf("Invalid path: %s", err))
		os.Exit(1)
	}

	// Ensure global config exists (first-run bootstrap)
	ensureGlobalConfig()

	cfg, err := config.Resolve(absExe)
	if err != nil {
		showError(fmt.Sprintf("Config error: %s", err))
		os.Exit(1)
	}

	if err := proton.Run(absExe, cfg); err != nil {
		showError(fmt.Sprintf("Launch failed: %s", err))
		os.Exit(1)
	}
}

func cmdConfig(args []string) {
	fs := flag.NewFlagSet("config", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: proton-launcher config [<exe>]")
		fmt.Fprintln(os.Stderr, "  Opens global config if no exe is given.")
		fs.PrintDefaults()
	}
	fs.Parse(args)

	if fs.NArg() > 0 {
		// Per-game config
		exePath := fs.Arg(0)
		absExe, err := filepath.Abs(exePath)
		if err != nil {
			showError(fmt.Sprintf("Invalid path: %s", err))
			os.Exit(1)
		}
		cfgPath, err := config.GameConfigPath(absExe)
		if err != nil {
			showError(fmt.Sprintf("Config error: %s", err))
			os.Exit(1)
		}
		cfg, err := config.Load(cfgPath)
		if err != nil {
			showError(fmt.Sprintf("Config error: %s", err))
			os.Exit(1)
		}
		gui.ShowConfigForm(cfgPath, cfg, absExe)
	} else {
		// Global config
		cfgPath, err := config.GlobalConfigPath()
		if err != nil {
			showError(fmt.Sprintf("Config error: %s", err))
			os.Exit(1)
		}
		cfg, err := config.Load(cfgPath)
		if err != nil {
			showError(fmt.Sprintf("Config error: %s", err))
			os.Exit(1)
		}
		gui.ShowConfigForm(cfgPath, cfg, "")
	}
}

func cmdList() {
	versions := proton.Discover()
	if len(versions) == 0 {
		fmt.Println("No Proton versions found.")
		return
	}
	for _, v := range versions {
		fmt.Printf("%-30s  %s  (%s)\n", v.Name, v.Path, v.Source)
	}
}

func cmdDesktop(args []string) {
	fs := flag.NewFlagSet("desktop", flag.ExitOnError)
	name := fs.String("name", "", "Display name for the shortcut")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: proton-launcher desktop <exe> [-name \"Game Name\"]")
		fs.PrintDefaults()
	}
	fs.Parse(args)

	if fs.NArg() < 1 {
		fs.Usage()
		os.Exit(1)
	}

	path, err := desktop.CreateShortcut(fs.Arg(0), *name)
	if err != nil {
		showError(fmt.Sprintf("Failed to create shortcut: %s", err))
		os.Exit(1)
	}
	fmt.Printf("Created shortcut: %s\n", path)
}

// ensureGlobalConfig creates a default global config if none exists,
// using the first discovered Proton version.
func ensureGlobalConfig() {
	cfgPath, err := config.GlobalConfigPath()
	if err != nil {
		return
	}
	if _, err := os.Stat(cfgPath); err == nil {
		return // already exists
	}

	versions := proton.Discover()
	cfg := config.DefaultGlobalConfig()
	if len(versions) > 0 {
		cfg.ProtonVersion = config.StringPtr(versions[0].Name)
	}

	if err := config.Save(cfgPath, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to save default config: %v\n", err)
	}
}

// showError displays an error to the user. Tries kdialog (KDE), then zenity
// (GTK), then notify-send as a fallback. Always prints to stderr.
func showError(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	if _, err := exec.LookPath("kdialog"); err == nil {
		exec.Command("kdialog", "--error", msg, "--title", "Proton Launcher").Run()
	} else if _, err := exec.LookPath("zenity"); err == nil {
		exec.Command("zenity", "--error", "--text", msg, "--title", "Proton Launcher").Run()
	} else if _, err := exec.LookPath("notify-send"); err == nil {
		exec.Command("notify-send", "-a", "Proton Launcher", "Proton Launcher", msg).Run()
	}
}

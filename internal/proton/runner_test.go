package proton

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nihil5320/proton-launcher/internal/config"
)

func TestVersionPriority(t *testing.T) {
	tests := []struct {
		version  Version
		wantPrio int
	}{
		{Version{Name: "proton-cachyos-slr", Source: SourceSteamCompat}, 0},
		{Version{Name: "Proton 9.0", Source: SourceSteamBundled}, 1},
		{Version{Name: "GE-Proton10-32", Source: SourceSteamCompat}, 2},
		{Version{Name: "Proton - Experimental", Source: SourceSteamBundled}, 3},
		{Version{Name: "Proton Hotfix", Source: SourceSteamBundled}, 3},
		{Version{Name: "lutris-wine-runner", Source: SourceLutris}, 4},
		{Version{Name: "some-other", Source: SourceSteamCompat}, 5},
	}
	for _, tt := range tests {
		got := versionPriority(tt.version)
		if got != tt.wantPrio {
			t.Errorf("versionPriority(%q, src=%q) = %d, want %d", tt.version.Name, tt.version.Source, got, tt.wantPrio)
		}
	}
}

func TestDedup(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "proton-real")
	if err := os.WriteFile(target, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	link1 := filepath.Join(dir, "proton-link1")
	link2 := filepath.Join(dir, "proton-link2")
	os.Symlink(target, link1)
	os.Symlink(target, link2)

	versions := []Version{
		{Name: "A", Path: link1, Source: SourceSteamCompat},
		{Name: "B", Path: link2, Source: SourceSteamCompat},
		{Name: "C", Path: target, Source: SourceSteamCompat},
	}

	deduped := dedup(versions)
	if len(deduped) != 1 {
		t.Errorf("dedup returned %d versions, want 1", len(deduped))
	}
}

func TestDedupDistinct(t *testing.T) {
	dir := t.TempDir()
	f1 := filepath.Join(dir, "proton1")
	f2 := filepath.Join(dir, "proton2")
	os.WriteFile(f1, []byte("1"), 0o755)
	os.WriteFile(f2, []byte("2"), 0o755)

	versions := []Version{
		{Name: "A", Path: f1},
		{Name: "B", Path: f2},
	}

	deduped := dedup(versions)
	if len(deduped) != 2 {
		t.Errorf("dedup returned %d versions, want 2", len(deduped))
	}
}

func TestReadVDFDisplayName(t *testing.T) {
	dir := t.TempDir()
	vdfPath := filepath.Join(dir, "compatibilitytool.vdf")

	content := `"compatibilitytools"
{
  "compat_tools"
  {
    "proton_ge"
    {
      "install_path" "."
      "display_name" "GE-Proton10-32"
      "from_oslist"  "windows"
      "to_oslist"    "linux"
    }
  }
}`
	if err := os.WriteFile(vdfPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	got := readVDFDisplayName(vdfPath)
	if got != "GE-Proton10-32" {
		t.Errorf("readVDFDisplayName = %q, want GE-Proton10-32", got)
	}
}

func TestReadVDFDisplayNameMissing(t *testing.T) {
	got := readVDFDisplayName("/nonexistent/file.vdf")
	if got != "" {
		t.Errorf("readVDFDisplayName for missing file = %q, want empty", got)
	}
}

func TestCleanEnv(t *testing.T) {
	t.Setenv("WINEPREFIX", "/some/prefix")
	t.Setenv("GAMEID", "test")

	env := cleanEnv()
	for _, entry := range env {
		if entry == "WINEPREFIX=/some/prefix" || entry == "GAMEID=test" {
			t.Errorf("cleanEnv should have stripped %q", entry)
		}
	}
}

func TestBuildCommandIncludesLaunchArgs(t *testing.T) {
	ver := Version{Name: "Proton 9.0", Path: "/proton/proton"}
	cfg := &config.Config{
		LaunchArgs: []string{"-fullscreen"},
		MangoHud:   config.BoolPtr(false),
		Gamescope:  config.BoolPtr(false),
		GameMode:   config.BoolPtr(false),
	}
	args, _ := buildCommand(ver, "/game.exe", cfg)

	found := false
	for _, a := range args {
		if a == "-fullscreen" {
			found = true
		}
	}
	if !found {
		t.Errorf("buildCommand did not include launch args: %v", args)
	}
}

func TestBuildUmuCommandIncludesExe(t *testing.T) {
	cfg := &config.Config{
		MangoHud:  config.BoolPtr(false),
		Gamescope: config.BoolPtr(false),
		GameMode:  config.BoolPtr(false),
	}
	args, _ := buildUmuCommand("/usr/bin/umu-run", "/game.exe", cfg)

	if args[0] != "/usr/bin/umu-run" {
		t.Errorf("first arg should be umu-run path, got %q", args[0])
	}
	if args[1] != "/game.exe" {
		t.Errorf("second arg should be exe path, got %q", args[1])
	}
}

func TestReservedEnvKeysNotInUserEnv(t *testing.T) {
	ver := Version{Name: "Proton 9.0", Path: "/proton/proton"}
	cfg := &config.Config{
		GameID: config.StringPtr("test-game"),
		Env: map[string]string{
			"WINEPREFIX": "/override",
			"HOME":       "/evil",
			"CUSTOM_VAR": "allowed",
		},
	}

	env := buildUmuEnv(ver, "/prefix", cfg)

	for _, entry := range env {
		if entry == "WINEPREFIX=/override" {
			t.Error("user WINEPREFIX override should be blocked")
		}
		if entry == "HOME=/evil" {
			t.Error("user HOME override should be blocked")
		}
	}

	found := false
	for _, entry := range env {
		if entry == "CUSTOM_VAR=allowed" {
			found = true
		}
	}
	if !found {
		t.Error("non-reserved env var should be allowed")
	}
}

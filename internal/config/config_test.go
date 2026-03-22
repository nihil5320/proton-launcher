package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSanitizeGameName(t *testing.T) {
	tests := []struct {
		name     string
		exePath  string
		wantName string
	}{
		{"simple exe", "/home/user/Games/MyGame.exe", "mygame-"},
		{"uppercase extension", "/home/user/Games/Test.EXE", "test-"},
		{"spaces in name", "/home/user/Games/My Game.exe", "my-game-"},
		{"special chars", "/home/user/Games/game (v2.1).exe", "game-v2-1-"},
		{"no extension", "/home/user/Games/game", "game-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeGameName(tt.exePath)
			if len(got) <= len(tt.wantName) {
				t.Errorf("SanitizeGameName(%q) = %q, expected prefix %q + hash", tt.exePath, got, tt.wantName)
				return
			}
			prefix := got[:len(tt.wantName)]
			if prefix != tt.wantName {
				t.Errorf("SanitizeGameName(%q) = %q, expected prefix %q", tt.exePath, got, tt.wantName)
			}
		})
	}
}

func TestSanitizeGameNameDifferentDirs(t *testing.T) {
	a := SanitizeGameName("/home/user/Games/A/game.exe")
	b := SanitizeGameName("/home/user/Games/B/game.exe")
	if a == b {
		t.Errorf("expected different hashes for different dirs, got %q and %q", a, b)
	}
}

func TestSanitizeGameNameEmptyName(t *testing.T) {
	got := SanitizeGameName("/home/user/Games/.exe")
	if got == "" {
		t.Error("SanitizeGameName should return hash for empty name")
	}
}

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home dir")
	}
	tests := []struct {
		input string
		want  string
	}{
		{"~/foo/bar", filepath.Join(home, "foo/bar")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
		{"~nothome", "~nothome"},
	}
	for _, tt := range tests {
		got := ExpandPath(tt.input)
		if got != tt.want {
			t.Errorf("ExpandPath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMergeOverridesFields(t *testing.T) {
	base := &Config{
		ProtonVersion: StringPtr("base-ver"),
		UseUmu:        BoolPtr(true),
		MangoHud:      BoolPtr(false),
		LaunchArgs:    []string{"--base-arg"},
		Env:           map[string]string{"BASE": "1"},
	}
	override := &Config{
		ProtonVersion: StringPtr("override-ver"),
		MangoHud:      BoolPtr(true),
		LaunchArgs:    []string{"--override-arg"},
		Env:           map[string]string{"OVER": "2"},
	}

	merged := Merge(base, override)

	if *merged.ProtonVersion != "override-ver" {
		t.Errorf("ProtonVersion = %q, want override-ver", *merged.ProtonVersion)
	}
	if *merged.UseUmu != true {
		t.Error("UseUmu should inherit from base")
	}
	if *merged.MangoHud != true {
		t.Error("MangoHud should be overridden to true")
	}
	if len(merged.LaunchArgs) != 1 || merged.LaunchArgs[0] != "--override-arg" {
		t.Errorf("LaunchArgs = %v, want [--override-arg]", merged.LaunchArgs)
	}
	if merged.Env["BASE"] != "1" || merged.Env["OVER"] != "2" {
		t.Errorf("Env = %v, want merged map", merged.Env)
	}
}

func TestMergeDeepCopyIsolation(t *testing.T) {
	base := &Config{
		LaunchArgs: []string{"--base"},
		Env:        map[string]string{"K": "V"},
	}
	override := &Config{}
	merged := Merge(base, override)

	merged.LaunchArgs = append(merged.LaunchArgs, "--extra")
	if len(base.LaunchArgs) != 1 {
		t.Error("mutating merged.LaunchArgs affected base")
	}
	merged.Env["NEW"] = "val"
	if _, ok := base.Env["NEW"]; ok {
		t.Error("mutating merged.Env affected base")
	}
}

func TestLoadSaveRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.toml")

	cfg := &Config{
		ProtonVersion: StringPtr("test-proton"),
		UseUmu:        BoolPtr(true),
		MangoHud:      BoolPtr(false),
		LaunchArgs:    []string{"-arg1", "-arg2"},
		Env:           map[string]string{"KEY": "val"},
		GamescopeOpts: &GamescopeOpts{
			Width:      IntPtr(1920),
			Height:     IntPtr(1080),
			Fullscreen: BoolPtr(true),
		},
	}

	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if *loaded.ProtonVersion != "test-proton" {
		t.Errorf("ProtonVersion = %q", *loaded.ProtonVersion)
	}
	if len(loaded.LaunchArgs) != 2 {
		t.Errorf("LaunchArgs = %v", loaded.LaunchArgs)
	}
	if loaded.Env["KEY"] != "val" {
		t.Errorf("Env = %v", loaded.Env)
	}
	if loaded.GamescopeOpts == nil || *loaded.GamescopeOpts.Width != 1920 {
		t.Error("GamescopeOpts not loaded correctly")
	}
}

func TestLoadNonExistent(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.toml")
	if err != nil {
		t.Fatalf("Load should not error for missing file: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load should return empty config for missing file")
	}
}

func TestValidateRemovePath(t *testing.T) {
	tests := []struct {
		path    string
		wantErr bool
	}{
		{"/", true},
		{".", true},
		{"/home", true},
		{"/home/user", true},
		{"/home/user/data", true},
		{"/home/user/.local/share", false},
		{"/home/user/.local/share/proton/pfx", false},
	}
	home, _ := os.UserHomeDir()
	tests = append(tests, struct {
		path    string
		wantErr bool
	}{home, true})

	for _, tt := range tests {
		err := validateRemovePath(tt.path)
		if (err != nil) != tt.wantErr {
			t.Errorf("validateRemovePath(%q) err=%v, wantErr=%v", tt.path, err, tt.wantErr)
		}
	}
}

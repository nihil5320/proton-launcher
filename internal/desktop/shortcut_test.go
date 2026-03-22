package desktop

import "testing"

func TestEscapeExecPath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/home/user/game.exe", "/home/user/game.exe"},
		{"/home/user/my game.exe", "/home/user/my game.exe"},
		{`/home/user/ga"me.exe`, `/home/user/ga\"me.exe`},
		{"/home/user/$game.exe", "/home/user/\\$game.exe"},
		{"/home/user/`game`.exe", "/home/user/\\`game\\`.exe"},
		{"/home/user/100%.exe", "/home/user/100%%.exe"},
		{`C:\Games\test.exe`, `C:\\Games\\test.exe`},
	}
	for _, tt := range tests {
		got := escapeExecPath(tt.input)
		if got != tt.want {
			t.Errorf("escapeExecPath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"My Game", "My Game"},
		{"Game\nWith\nNewlines", "Game With Newlines"},
		{"Game\rWith\rCR", "Game With CR"},
		{"Game\t\tTabbed", "Game  Tabbed"},
		{"\x01\x02Hidden\x03", "Hidden"},
		{"  Spaces  ", "Spaces"},
		{"", ""},
	}
	for _, tt := range tests {
		got := sanitizeName(tt.input)
		if got != tt.want {
			t.Errorf("sanitizeName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"My Game", "my-game"},
		{"Game: Subtitle", "game--subtitle"},
		{"it's a game", "its-a-game"},
		{`"quoted"`, "quoted"},
		{"path/to\\file", "path-to-file"},
	}
	for _, tt := range tests {
		got := sanitizeFilename(tt.input)
		if got != tt.want {
			t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

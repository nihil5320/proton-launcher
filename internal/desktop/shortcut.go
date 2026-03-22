package desktop

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const shortcutTemplate = `[Desktop Entry]
Name={{.Name}}
Comment=Launch {{.Name}} with Proton
Exec=proton-launcher run "{{.ExePath}}"
Icon=applications-games
Terminal=false
Type=Application
Categories=Game;
`

// escapeExecPath escapes special characters inside a double-quoted Exec value
// per the freedesktop Desktop Entry spec.
func escapeExecPath(p string) string {
	r := strings.NewReplacer(
		"\\", "\\\\",
		"\"", "\\\"",
		"$", "\\$",
		"`", "\\`",
		"%", "%%",
	)
	return r.Replace(p)
}

// sanitizeName strips control characters from a display name to prevent
// injection of arbitrary keys into .desktop files.
func sanitizeName(name string) string {
	var b strings.Builder
	for _, r := range name {
		if r == '\n' || r == '\r' || r == '\t' {
			b.WriteRune(' ')
		} else if r >= 0x20 {
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}

func CreateShortcut(exePath, name string) (string, error) {
	absExe, err := filepath.Abs(exePath)
	if err != nil {
		return "", fmt.Errorf("resolving exe path: %w", err)
	}

	if name == "" {
		name = strings.TrimSuffix(filepath.Base(absExe), filepath.Ext(absExe))
	}
	name = sanitizeName(name)
	if name == "" {
		return "", fmt.Errorf("launcher entry name is empty after sanitization")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	appDir := filepath.Join(home, ".local", "share", "applications")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		return "", fmt.Errorf("creating applications dir: %w", err)
	}

	safeName := sanitizeFilename(name)
	destPath := filepath.Join(appDir, "proton-launcher-"+safeName+".desktop")

	if _, err := os.Stat(destPath); err == nil {
		return destPath, fmt.Errorf("launcher entry already exists: %s (delete it first to recreate)", destPath)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("creating desktop file: %w", err)
	}
	defer f.Close()

	tmpl := template.Must(template.New("shortcut").Parse(shortcutTemplate))
	data := struct {
		Name    string
		ExePath string
	}{
		Name:    name,
		ExePath: escapeExecPath(absExe),
	}
	if err := tmpl.Execute(f, data); err != nil {
		return "", fmt.Errorf("writing desktop file: %w", err)
	}

	return destPath, nil
}

func sanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		" ", "-",
		"/", "-",
		"\\", "-",
		":", "-",
		"'", "",
		"\"", "",
	)
	return strings.ToLower(replacer.Replace(name))
}

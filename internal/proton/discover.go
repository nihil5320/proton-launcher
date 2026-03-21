package proton

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Version struct {
	Name   string
	Path   string
	Source string
}

func Discover() []Version {
	var versions []Version

	home, err := os.UserHomeDir()
	if err != nil {
		return versions
	}

	steamCompatDirs := []string{
		filepath.Join(home, ".steam", "root", "compatibilitytools.d"),
		filepath.Join(home, ".local", "share", "Steam", "compatibilitytools.d"),
		"/usr/share/steam/compatibilitytools.d",
		filepath.Join(home, ".var", "app", "com.valvesoftware.Steam", "data", "Steam", "compatibilitytools.d"),
	}
	for _, dir := range steamCompatDirs {
		versions = append(versions, scanProtonDir(dir, "steam-compat")...)
	}

	steamApps := filepath.Join(home, ".local", "share", "Steam", "steamapps", "common")
	flatpakSteamApps := filepath.Join(home, ".var", "app", "com.valvesoftware.Steam", "data", "Steam", "steamapps", "common")
	for _, dir := range []string{steamApps, flatpakSteamApps} {
		if entries, err := os.ReadDir(dir); err == nil {
			for _, e := range entries {
				if !e.IsDir() || !strings.HasPrefix(e.Name(), "Proton") {
					continue
				}
				protonBin := filepath.Join(dir, e.Name(), "proton")
				if info, err := os.Stat(protonBin); err == nil && !info.IsDir() {
					versions = append(versions, Version{
						Name:   e.Name(),
						Path:   protonBin,
						Source: "steam-bundled",
					})
				}
			}
		}
	}

	lutrisDir := filepath.Join(home, ".local", "share", "lutris", "runners", "wine")
	if entries, err := os.ReadDir(lutrisDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			// Only pick up Proton binaries from Lutris; plain Wine
			// runners are incompatible with both execution paths.
			protonBin := filepath.Join(lutrisDir, e.Name(), "proton")
			if info, err := os.Stat(protonBin); err == nil && !info.IsDir() {
				versions = append(versions, Version{
					Name:   e.Name(),
					Path:   protonBin,
					Source: "lutris",
				})
			}
		}
	}

	versions = dedup(versions)
	sort.Slice(versions, func(i, j int) bool {
		pi, pj := versionPriority(versions[i]), versionPriority(versions[j])
		if pi != pj {
			return pi < pj
		}
		return versions[i].Name > versions[j].Name
	})
	return versions
}

func FindVersion(name string) (Version, error) {
	for _, v := range Discover() {
		if v.Name == name {
			return v, nil
		}
	}
	return Version{}, fmt.Errorf("proton version %q not found", name)
}

func scanProtonDir(dir, source string) []Version {
	var versions []Version
	entries, err := os.ReadDir(dir)
	if err != nil {
		return versions
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		protonBin := filepath.Join(dir, e.Name(), "proton")
		if info, err := os.Stat(protonBin); err == nil && !info.IsDir() {
			name := readVDFDisplayName(filepath.Join(dir, e.Name(), "compatibilitytool.vdf"))
			if name == "" {
				name = e.Name()
			}
			versions = append(versions, Version{
				Name:   name,
				Path:   protonBin,
				Source: source,
			})
		}
	}
	return versions
}

func readVDFDisplayName(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.Contains(line, "\"display_name\"") {
			continue
		}
		// Extract all quoted values from the line.
		var values []string
		rest := line
		for {
			idx := strings.IndexByte(rest, '"')
			if idx < 0 {
				break
			}
			rest = rest[idx+1:]
			end := strings.IndexByte(rest, '"')
			if end < 0 {
				break
			}
			values = append(values, rest[:end])
			rest = rest[end+1:]
		}
		// Expect ["display_name", <value>]
		if len(values) >= 2 && strings.EqualFold(values[0], "display_name") {
			return values[1]
		}
	}
	return ""
}

func versionPriority(v Version) int {
	low := strings.ToLower(v.Name)
	switch {
	case strings.Contains(low, "cachyos"):
		return 0
	case v.Source == "steam-bundled" && !strings.Contains(low, "experimental") && !strings.Contains(low, "hotfix"):
		return 1
	case strings.Contains(low, "ge-proton"):
		return 2
	case strings.Contains(low, "experimental") || strings.Contains(low, "hotfix"):
		return 3
	case v.Source == "lutris":
		return 4
	default:
		return 5
	}
}

func dedup(versions []Version) []Version {
	seen := make(map[string]bool)
	var out []Version
	for _, v := range versions {
		key := v.Path
		if resolved, err := filepath.EvalSymlinks(v.Path); err == nil {
			key = resolved
		} else {
			fmt.Fprintf(os.Stderr, "Warning: could not resolve symlink for %s: %v\n", v.Path, err)
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, v)
	}
	return out
}

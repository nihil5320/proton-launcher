package gui

import (
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/nihil5320/proton-launcher/internal/config"
	"github.com/nihil5320/proton-launcher/internal/proton"
)

func ShowConfigForm(cfgPath string, cfg *config.Config, exePath string) {
	a := app.New()

	title := "Proton Launcher — Global Config"
	if exePath != "" {
		name := strings.TrimSuffix(filepath.Base(exePath), filepath.Ext(exePath))
		title = "Proton Launcher — " + name
	}
	w := a.NewWindow(title)
	w.Resize(fyne.NewSize(500, 450))

	versions := proton.Discover()
	versionNames := make([]string, len(versions))
	for i, v := range versions {
		versionNames[i] = v.Name
	}

	currentVersion := ""
	if cfg.ProtonVersion != nil {
		currentVersion = *cfg.ProtonVersion
	}
	currentPrefix := ""
	if cfg.PrefixPath != nil {
		currentPrefix = *cfg.PrefixPath
	}
	currentMangoHud := cfg.MangoHud != nil && *cfg.MangoHud
	currentGamescope := cfg.Gamescope != nil && *cfg.Gamescope
	currentGameMode := cfg.GameMode != nil && *cfg.GameMode

	var envLines []string
	for k, v := range cfg.Env {
		envLines = append(envLines, k+"="+v)
	}
	currentEnv := strings.Join(envLines, "\n")
	currentArgs := strings.Join(cfg.LaunchArgs, " ")

	versionSelect := widget.NewSelect(versionNames, nil)
	versionSelect.SetSelected(currentVersion)

	prefixEntry := widget.NewEntry()
	prefixEntry.SetText(currentPrefix)
	prefixEntry.SetPlaceHolder("~/.local/share/proton-launcher/prefixes/default")

	envEntry := widget.NewMultiLineEntry()
	envEntry.SetText(currentEnv)
	envEntry.SetPlaceHolder("KEY=VALUE (one per line)")
	envEntry.SetMinRowsVisible(4)

	argsEntry := widget.NewEntry()
	argsEntry.SetText(currentArgs)
	argsEntry.SetPlaceHolder("-fullscreen -skipintro")

	mangoHudCheck := widget.NewCheck("MangoHud", nil)
	mangoHudCheck.SetChecked(currentMangoHud)

	gamescopeCheck := widget.NewCheck("Gamescope", nil)
	gamescopeCheck.SetChecked(currentGamescope)

	gameModeCheck := widget.NewCheck("GameMode", nil)
	gameModeCheck.SetChecked(currentGameMode)

	checks := container.NewHBox(mangoHudCheck, gamescopeCheck, gameModeCheck)

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Proton Version", Widget: versionSelect},
			{Text: "Prefix Path", Widget: prefixEntry},
			{Text: "Environment", Widget: envEntry},
			{Text: "Launch Args", Widget: argsEntry},
			{Text: "Options", Widget: checks},
		},
		OnSubmit: func() {
			newCfg := &config.Config{}

			if versionSelect.Selected != "" {
				newCfg.ProtonVersion = config.StringPtr(versionSelect.Selected)
			}
			if prefixEntry.Text != "" {
				newCfg.PrefixPath = config.StringPtr(prefixEntry.Text)
			}

			newCfg.MangoHud = config.BoolPtr(mangoHudCheck.Checked)
			newCfg.Gamescope = config.BoolPtr(gamescopeCheck.Checked)
			newCfg.GameMode = config.BoolPtr(gameModeCheck.Checked)

			if argsEntry.Text != "" {
				newCfg.LaunchArgs = strings.Fields(argsEntry.Text)
			}

			if envEntry.Text != "" {
				newCfg.Env = parseEnvLines(envEntry.Text)
			}

			config.Save(cfgPath, newCfg)
			a.Quit()
		},
		OnCancel: func() {
			a.Quit()
		},
		SubmitText: "Save",
		CancelText: "Cancel",
	}

	w.SetContent(form)
	w.ShowAndRun()
}

func parseEnvLines(text string) map[string]string {
	env := make(map[string]string)
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if k, v, ok := strings.Cut(line, "="); ok {
			env[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	return env
}

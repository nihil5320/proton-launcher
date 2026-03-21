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

	versions := proton.Discover()
	versionNames := make([]string, len(versions))
	for i, v := range versions {
		versionNames[i] = v.Name
	}

	currentVersion := ""
	if cfg.ProtonVersion != nil {
		currentVersion = *cfg.ProtonVersion
	} else if len(versionNames) > 0 {
		currentVersion = versionNames[0]
	}
	currentPrefix := ""
	if cfg.PrefixPath != nil {
		currentPrefix = *cfg.PrefixPath
	}
	currentMangoHud := cfg.MangoHud != nil && *cfg.MangoHud
	currentGamescope := cfg.Gamescope != nil && *cfg.Gamescope
	currentGameMode := cfg.GameMode != nil && *cfg.GameMode
	currentUseUmu := cfg.UseUmu == nil || (cfg.UseUmu != nil && *cfg.UseUmu)
	currentGameID := ""
	if cfg.GameID != nil {
		currentGameID = *cfg.GameID
	}
	currentLocale := ""
	if cfg.Locale != nil {
		currentLocale = *cfg.Locale
	}

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
	prefixEntry.SetPlaceHolder("Leave empty for automatic per-game prefix")

	envEntry := widget.NewMultiLineEntry()
	envEntry.SetText(currentEnv)
	envEntry.SetPlaceHolder("KEY=VALUE (one per line)")
	envEntry.SetMinRowsVisible(4)

	argsEntry := widget.NewEntry()
	argsEntry.SetText(currentArgs)
	argsEntry.SetPlaceHolder("-fullscreen -skipintro")

	useUmuCheck := widget.NewCheck("Use umu-run", nil)
	useUmuCheck.SetChecked(currentUseUmu)

	gameIDEntry := widget.NewEntry()
	gameIDEntry.SetText(currentGameID)
	gameIDEntry.SetPlaceHolder("umu-default")

	localeOptions := []string{
		"System Default",
		"ja_JP.UTF-8",
		"zh_CN.UTF-8",
		"zh_TW.UTF-8",
		"ko_KR.UTF-8",
		"th_TH.UTF-8",
		"ru_RU.UTF-8",
	}
	localeSelect := widget.NewSelect(localeOptions, nil)
	if currentLocale != "" {
		localeSelect.SetSelected(currentLocale)
	} else {
		localeSelect.SetSelected("System Default")
	}

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
			{Text: "Use umu-run", Widget: useUmuCheck},
			{Text: "Game ID", Widget: gameIDEntry},
			{Text: "Locale", Widget: localeSelect},
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
			newCfg.UseUmu = config.BoolPtr(useUmuCheck.Checked)

			if gameIDEntry.Text != "" {
				newCfg.GameID = config.StringPtr(gameIDEntry.Text)
			}

			if localeSelect.Selected != "" && localeSelect.Selected != "System Default" {
				newCfg.Locale = config.StringPtr(localeSelect.Selected)
			}

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

	w.SetContent(container.NewVBox(form))
	w.Resize(fyne.NewSize(500, w.Content().MinSize().Height))
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

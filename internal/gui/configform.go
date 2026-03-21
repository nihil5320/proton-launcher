package gui

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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
	currentArgs := strings.Join(cfg.LaunchArgs, "\n")

	versionSelect := widget.NewSelect(versionNames, nil)
	if len(versionNames) == 0 {
		versionSelect.PlaceHolder = "No Proton versions found — install Proton and reopen"
	}
	versionSelect.SetSelected(currentVersion)

	prefixEntry := widget.NewEntry()
	prefixEntry.SetText(currentPrefix)
	prefixEntry.SetPlaceHolder("Leave empty for automatic per-game prefix")

	envEntry := widget.NewMultiLineEntry()
	envEntry.SetText(currentEnv)
	envEntry.SetPlaceHolder("KEY=VALUE (one per line)")
	envEntry.SetMinRowsVisible(4)

	argsEntry := widget.NewMultiLineEntry()
	argsEntry.SetText(currentArgs)
	argsEntry.SetPlaceHolder("One argument per line")
	argsEntry.SetMinRowsVisible(3)

	useUmuCheck := widget.NewCheck("Use umu-run", nil)
	useUmuCheck.SetChecked(currentUseUmu)

	gameIDEntry := widget.NewEntry()
	gameIDEntry.SetText(currentGameID)
	gameIDEntry.SetPlaceHolder("umu-default")

	localeOptions := []string{
		"System Default",
		"English (en_US.UTF-8)",
		"French (fr_FR.UTF-8)",
		"German (de_DE.UTF-8)",
		"Spanish (es_ES.UTF-8)",
		"Italian (it_IT.UTF-8)",
		"Portuguese - Brazil (pt_BR.UTF-8)",
		"Japanese (ja_JP.UTF-8)",
		"Chinese - Simplified (zh_CN.UTF-8)",
		"Chinese - Traditional (zh_TW.UTF-8)",
		"Korean (ko_KR.UTF-8)",
		"Thai (th_TH.UTF-8)",
		"Russian (ru_RU.UTF-8)",
		"Polish (pl_PL.UTF-8)",
	}
	localeSelect := widget.NewSelect(localeOptions, nil)
	if currentLocale != "" {
		localeSelect.SetSelected(localeLabelFromCode(currentLocale))
	} else {
		localeSelect.SetSelected("System Default")
	}

	mangoHudCheck := widget.NewCheck("MangoHud", nil)
	mangoHudCheck.SetChecked(currentMangoHud)

	gamescopeCheck := widget.NewCheck("Gamescope", nil)
	gamescopeCheck.SetChecked(currentGamescope)

	gsWidthEntry := widget.NewEntry()
	gsWidthEntry.SetPlaceHolder("e.g. 1920")
	gsHeightEntry := widget.NewEntry()
	gsHeightEntry.SetPlaceHolder("e.g. 1080")
	gsFullscreenCheck := widget.NewCheck("Fullscreen", nil)

	updateGamescopeFields := func(enabled bool) {
		if enabled {
			gsWidthEntry.Enable()
			gsHeightEntry.Enable()
			gsFullscreenCheck.Enable()
		} else {
			gsWidthEntry.Disable()
			gsHeightEntry.Disable()
			gsFullscreenCheck.Disable()
		}
	}

	if cfg.GamescopeOpts != nil {
		if cfg.GamescopeOpts.Width != nil {
			gsWidthEntry.SetText(fmt.Sprintf("%d", *cfg.GamescopeOpts.Width))
		}
		if cfg.GamescopeOpts.Height != nil {
			gsHeightEntry.SetText(fmt.Sprintf("%d", *cfg.GamescopeOpts.Height))
		}
		if cfg.GamescopeOpts.Fullscreen != nil {
			gsFullscreenCheck.SetChecked(*cfg.GamescopeOpts.Fullscreen)
		}
	}
	updateGamescopeFields(currentGamescope)

	gamescopeCheck.OnChanged = func(checked bool) {
		updateGamescopeFields(checked)
	}

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
			{Text: "Gamescope Width", Widget: gsWidthEntry},
			{Text: "Gamescope Height", Widget: gsHeightEntry},
			{Text: "Gamescope Fullscreen", Widget: gsFullscreenCheck},
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

			if gsWidthEntry.Text != "" || gsHeightEntry.Text != "" || gsFullscreenCheck.Checked {
				opts := &config.GamescopeOpts{}
				if w, err := strconv.Atoi(gsWidthEntry.Text); err == nil && w > 0 {
					opts.Width = config.IntPtr(w)
				}
				if h, err := strconv.Atoi(gsHeightEntry.Text); err == nil && h > 0 {
					opts.Height = config.IntPtr(h)
				}
				if gsFullscreenCheck.Checked {
					opts.Fullscreen = config.BoolPtr(true)
				}
				newCfg.GamescopeOpts = opts
			}

			if gameIDEntry.Text != "" {
				newCfg.GameID = config.StringPtr(gameIDEntry.Text)
			}

			if localeSelect.Selected != "" && localeSelect.Selected != "System Default" {
				newCfg.Locale = config.StringPtr(localeCodeFromLabel(localeSelect.Selected))
			}

			if argsEntry.Text != "" {
				newCfg.LaunchArgs = parseLines(argsEntry.Text)
			}

			if envEntry.Text != "" {
				newCfg.Env = parseEnvLines(envEntry.Text)
			}

			if err := config.Save(cfgPath, newCfg); err != nil {
				dialog.ShowError(err, w)
				return
			}
			a.Quit()
		},
		OnCancel: func() {
			a.Quit()
		},
		SubmitText: "Save",
		CancelText: "Cancel",
	}

	w.SetContent(container.NewVBox(form, dangerZone(w, exePath)))
	w.Resize(fyne.NewSize(500, w.Content().MinSize().Height))
	w.ShowAndRun()
}

func dangerZone(w fyne.Window, exePath string) fyne.CanvasObject {
	// showDoneAndClose shows an informational dialog that closes the window
	// when dismissed, since the underlying data has changed.
	showDoneAndClose := func(title, msg string) {
		d := dialog.NewInformation(title, msg, w)
		d.SetOnClosed(func() { w.Close() })
		d.Show()
	}

	var buttons []fyne.CanvasObject

	if exePath != "" {
		// Per-game actions
		buttons = append(buttons,
			widget.NewButton("Clear Prefix", func() {
				dialog.ShowConfirm(
					"Clear Prefix?",
					"This will delete the Wine prefix directory for this game. All game-specific Wine data (saves, settings, installed components) will be lost.",
					func(ok bool) {
						if !ok {
							return
						}
						if err := config.DeleteGamePrefix(exePath); err != nil {
							dialog.ShowError(err, w)
						} else {
							showDoneAndClose("Done", "Prefix cleared. A fresh prefix will be created on next launch.")
						}
					}, w)
			}),
			widget.NewButton("Delete Config", func() {
				dialog.ShowConfirm(
					"Delete Config?",
					"This will delete the per-game configuration. The game will use global defaults on next launch.",
					func(ok bool) {
						if !ok {
							return
						}
						if err := config.DeleteGameConfig(exePath); err != nil {
							dialog.ShowError(err, w)
						} else {
							showDoneAndClose("Done", "Game config deleted.")
						}
					}, w)
			}),
		)
	} else {
		// Global actions
		buttons = append(buttons,
			widget.NewButton("Reset to Defaults", func() {
				dialog.ShowConfirm(
					"Reset Global Config?",
					"This will overwrite the global config with defaults.",
					func(ok bool) {
						if !ok {
							return
						}
						if err := config.ResetGlobalConfig(); err != nil {
							dialog.ShowError(err, w)
						} else {
							showDoneAndClose("Done", "Global config reset to defaults.")
						}
					}, w)
			}),
			widget.NewButton("Clear All Prefixes", func() {
				dialog.ShowConfirm(
					"Clear All Prefixes?",
					"This will delete ALL Wine prefix directories. All game saves and Wine data in the default prefix location will be lost.",
					func(ok bool) {
						if !ok {
							return
						}
						if err := config.DeleteAllPrefixes(); err != nil {
							dialog.ShowError(err, w)
						} else {
							showDoneAndClose("Done", "All prefixes cleared.")
						}
					}, w)
			}),
			widget.NewButton("Clear All Game Configs", func() {
				dialog.ShowConfirm(
					"Clear All Game Configs?",
					"This will delete ALL per-game configuration files. Every game will revert to global defaults.",
					func(ok bool) {
						if !ok {
							return
						}
						if err := config.DeleteAllGameConfigs(); err != nil {
							dialog.ShowError(err, w)
						} else {
							showDoneAndClose("Done", "All game configs deleted.")
						}
					}, w)
			}),
		)
	}

	header := widget.NewRichTextFromMarkdown("**Manage Data**")
	buttonRow := container.NewHBox(buttons...)
	return container.NewVBox(widget.NewSeparator(), header, buttonRow)
}

func parseEnvLines(text string) map[string]string {
	env := make(map[string]string)
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if k, v, ok := strings.Cut(line, "="); ok {
			v = strings.TrimSpace(v)
			v = stripQuotes(v)
			env[strings.TrimSpace(k)] = v
		}
	}
	return env
}

func parseLines(text string) []string {
	var out []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

func stripQuotes(s string) string {
	if len(s) >= 2 && ((s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'')) {
		return s[1 : len(s)-1]
	}
	return s
}

var localeLabels = []struct {
	code  string
	label string
}{
	{"en_US.UTF-8", "English (en_US.UTF-8)"},
	{"fr_FR.UTF-8", "French (fr_FR.UTF-8)"},
	{"de_DE.UTF-8", "German (de_DE.UTF-8)"},
	{"es_ES.UTF-8", "Spanish (es_ES.UTF-8)"},
	{"it_IT.UTF-8", "Italian (it_IT.UTF-8)"},
	{"pt_BR.UTF-8", "Portuguese - Brazil (pt_BR.UTF-8)"},
	{"ja_JP.UTF-8", "Japanese (ja_JP.UTF-8)"},
	{"zh_CN.UTF-8", "Chinese - Simplified (zh_CN.UTF-8)"},
	{"zh_TW.UTF-8", "Chinese - Traditional (zh_TW.UTF-8)"},
	{"ko_KR.UTF-8", "Korean (ko_KR.UTF-8)"},
	{"th_TH.UTF-8", "Thai (th_TH.UTF-8)"},
	{"ru_RU.UTF-8", "Russian (ru_RU.UTF-8)"},
	{"pl_PL.UTF-8", "Polish (pl_PL.UTF-8)"},
}

func localeLabelFromCode(code string) string {
	for _, l := range localeLabels {
		if l.code == code {
			return l.label
		}
	}
	return code
}

func localeCodeFromLabel(label string) string {
	for _, l := range localeLabels {
		if l.label == label {
			return l.code
		}
	}
	return label
}

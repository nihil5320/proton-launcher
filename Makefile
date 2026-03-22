PREFIX ?= $(HOME)/.local
BINDIR = $(PREFIX)/bin
APPDIR = $(PREFIX)/share/applications
SERVICEDIR = $(PREFIX)/share/kio/servicemenus
NAUTILUSDIR = $(PREFIX)/share/nautilus/scripts

BINARY = proton-launcher
BUILD_DIR = build

.PHONY: build install install-common install-kde install-gnome install-cosmic \
        uninstall clean

build:
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/proton-launcher

install: build install-common install-kde install-gnome install-cosmic
	@echo ""
	@echo "Installed proton-launcher to $(BINDIR)/$(BINARY)"
	@echo ""
	@echo "To set proton-launcher as the default handler for .exe files:"
	@echo "  xdg-mime default proton-launcher.desktop application/x-ms-dos-executable"
	@echo ""

install-common:
	install -Dm755 $(BUILD_DIR)/$(BINARY) $(DESTDIR)$(BINDIR)/$(BINARY)
	install -Dm644 assets/proton-launcher.desktop $(DESTDIR)$(APPDIR)/proton-launcher.desktop
	sed -i 's|Exec=proton-launcher |Exec=$(BINDIR)/$(BINARY) |g' $(DESTDIR)$(APPDIR)/proton-launcher.desktop
	install -Dm644 assets/proton-launcher-config.desktop $(DESTDIR)$(APPDIR)/proton-launcher-config.desktop
	sed -i 's|Exec=proton-launcher |Exec=$(BINDIR)/$(BINARY) |g' $(DESTDIR)$(APPDIR)/proton-launcher-config.desktop
	@if command -v update-desktop-database >/dev/null 2>&1 && [ -z "$(DESTDIR)" ]; then \
		update-desktop-database $(APPDIR); \
	fi

install-kde:
	install -Dm755 assets/proton-launcher-service.desktop $(DESTDIR)$(SERVICEDIR)/proton-launcher-service.desktop
	sed -i 's|Exec=proton-launcher |Exec=$(BINDIR)/$(BINARY) |g' $(DESTDIR)$(SERVICEDIR)/proton-launcher-service.desktop

install-gnome:
	install -Dm755 assets/nautilus/proton-launcher-configure $(DESTDIR)$(NAUTILUSDIR)/proton-launcher-configure
	sed -i 's|exec proton-launcher |exec $(BINDIR)/$(BINARY) |g' $(DESTDIR)$(NAUTILUSDIR)/proton-launcher-configure
	install -Dm755 assets/nautilus/proton-launcher-shortcut $(DESTDIR)$(NAUTILUSDIR)/proton-launcher-shortcut
	sed -i 's|exec proton-launcher |exec $(BINDIR)/$(BINARY) |g' $(DESTDIR)$(NAUTILUSDIR)/proton-launcher-shortcut

install-cosmic:
	@# COSMIC Files follows XDG standards; the MIME handler and .desktop entries
	@# installed by install-common are sufficient. Custom context menu actions
	@# will be added when cosmic-files supports them upstream.

uninstall:
	rm -f $(DESTDIR)$(BINDIR)/$(BINARY)
	rm -f $(DESTDIR)$(APPDIR)/proton-launcher.desktop
	rm -f $(DESTDIR)$(APPDIR)/proton-launcher-config.desktop
	rm -f $(DESTDIR)$(SERVICEDIR)/proton-launcher-service.desktop
	rm -f $(DESTDIR)$(NAUTILUSDIR)/proton-launcher-configure
	rm -f $(DESTDIR)$(NAUTILUSDIR)/proton-launcher-shortcut
	@if command -v update-desktop-database >/dev/null 2>&1 && [ -z "$(DESTDIR)" ]; then \
		update-desktop-database $(APPDIR); \
	fi
	@if [ -z "$(DESTDIR)" ]; then \
		rm -rf $(HOME)/.config/proton-launcher; \
		rm -rf $(HOME)/.local/share/proton-launcher; \
		echo "Uninstalled proton-launcher (binary, desktop files, config, and data)"; \
	else \
		echo "Uninstalled proton-launcher (binary and desktop files only; DESTDIR set, skipping user data)"; \
	fi

clean:
	@if [ "$(BUILD_DIR)" = "/" ] || [ "$(BUILD_DIR)" = "." ] || [ -z "$(BUILD_DIR)" ]; then \
		echo "Error: BUILD_DIR is empty, root, or current directory. Refusing to clean."; \
		exit 1; \
	fi
	rm -rf $(BUILD_DIR)

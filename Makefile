PREFIX ?= $(HOME)/.local
BINDIR = $(PREFIX)/bin
APPDIR = $(PREFIX)/share/applications
SERVICEDIR = $(PREFIX)/share/kio/servicemenus

BINARY = proton-launcher
BUILD_DIR = build

.PHONY: build install uninstall clean

build:
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/proton-launcher

install: build
	install -Dm755 $(BUILD_DIR)/$(BINARY) $(DESTDIR)$(BINDIR)/$(BINARY)
	install -Dm644 assets/proton-launcher.desktop $(DESTDIR)$(APPDIR)/proton-launcher.desktop
	sed -i 's|Exec=proton-launcher |Exec=$(BINDIR)/$(BINARY) |g' $(DESTDIR)$(APPDIR)/proton-launcher.desktop
	install -Dm644 assets/proton-launcher-config.desktop $(DESTDIR)$(APPDIR)/proton-launcher-config.desktop
	sed -i 's|Exec=proton-launcher |Exec=$(BINDIR)/$(BINARY) |g' $(DESTDIR)$(APPDIR)/proton-launcher-config.desktop
	install -Dm755 assets/proton-launcher-service.desktop $(DESTDIR)$(SERVICEDIR)/proton-launcher-service.desktop
	sed -i 's|Exec=proton-launcher |Exec=$(BINDIR)/$(BINARY) |g' $(DESTDIR)$(SERVICEDIR)/proton-launcher-service.desktop
	@if command -v update-desktop-database >/dev/null 2>&1 && [ -z "$(DESTDIR)" ]; then \
		update-desktop-database $(APPDIR); \
	fi
	@echo ""
	@echo "Installed proton-launcher to $(BINDIR)/$(BINARY)"
	@echo ""
	@echo "To set proton-launcher as the default handler for .exe files:"
	@echo "  xdg-mime default proton-launcher.desktop application/x-ms-dos-executable"
	@echo ""

uninstall:
	rm -f $(DESTDIR)$(BINDIR)/$(BINARY)
	rm -f $(DESTDIR)$(APPDIR)/proton-launcher.desktop
	rm -f $(DESTDIR)$(APPDIR)/proton-launcher-config.desktop
	rm -f $(DESTDIR)$(SERVICEDIR)/proton-launcher-service.desktop
	@if command -v update-desktop-database >/dev/null 2>&1 && [ -z "$(DESTDIR)" ]; then \
		update-desktop-database $(APPDIR); \
	fi
	@echo "Uninstalled proton-launcher"

clean:
	@if [ "$(BUILD_DIR)" = "/" ] || [ "$(BUILD_DIR)" = "." ] || [ -z "$(BUILD_DIR)" ]; then \
		echo "Error: BUILD_DIR is empty, root, or current directory. Refusing to clean."; \
		exit 1; \
	fi
	rm -rf $(BUILD_DIR)

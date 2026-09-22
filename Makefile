.DEFAULT_GOAL := build

GO ?= go
MAKEPKG ?= makepkg
PREFIX ?= /usr/local
DESTDIR ?=
VERSION ?= dev
BINARY ?= ocrecent
LDFLAGS ?= -s -w -X main.version=$(VERSION)

.PHONY: build package install raw-build raw-install clean

build: package

package:
	@if command -v "$(MAKEPKG)" >/dev/null 2>&1; then \
		cd packaging && "$(MAKEPKG)" -f --noconfirm -p PKGBUILD; \
	else \
		$(MAKE) raw-build; \
	fi

raw-build:
	CGO_ENABLED=0 $(GO) build -buildvcs=false -trimpath \
		-ldflags "$(LDFLAGS)" -o "$(BINARY)" ./cmd/ocrecent

install:
	@if command -v "$(MAKEPKG)" >/dev/null 2>&1; then \
		cd packaging && "$(MAKEPKG)" -si -p PKGBUILD; \
	else \
		$(MAKE) raw-install; \
	fi

raw-install: raw-build
	install -Dm755 "$(BINARY)" "$(DESTDIR)$(PREFIX)/bin/$(BINARY)"
	install -Dm644 packaging/ocrecent.1 "$(DESTDIR)$(PREFIX)/share/man/man1/ocrecent.1"
	install -Dm644 completions/bash/ocrecent "$(DESTDIR)$(PREFIX)/share/bash-completion/completions/ocrecent"
	install -Dm644 completions/zsh/_ocrecent "$(DESTDIR)$(PREFIX)/share/zsh/site-functions/_ocrecent"
	install -Dm644 LICENSE "$(DESTDIR)$(PREFIX)/share/licenses/$(BINARY)/LICENSE"

clean:
	rm -f "$(BINARY)"
	rm -rf dist packaging/pkg packaging/src
	rm -f packaging/*.tar.gz packaging/*.pkg.tar.*

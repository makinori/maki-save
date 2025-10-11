default:
	@just --list

# for linux
[group("desktop")]
build:
	CGO_ENABLED=0 GOOS=linux go build -o maki-immich ./desktop

# for linux
[group("desktop")]
install:
	#!/usr/bin/env bash
	set -euo pipefail
	cp maki-immich ~/maki-immich
	echo "installed to ~/maki-immich"
	echo "sample config for nautilus:"
	echo ""
	echo "~/.local/share/actions-for-nautilus/config.json"
	echo ""
	echo '{
		"debug": false,
		"actions": [
			{
				"type": "command",
				"label": "Maki Immich",
				"use_shell": true,
				"command_line": "NAUTILUS=1 ~/maki-immich \"%U\"",
				"min_items": 1,
				"filetypes": ["file"]
			}
		]
	}'


# for desktop without emulator
[group("android")]
start-android:
	INTENT_TEST=1 go run ./mobile 

[group("android")]
[working-directory("mobile")]
build-android:
	go tool fyne package -os android -app-id cafe.maki.immich \
	-icon icon.png -name "maki immich" -release
	mv maki_immich.apk ../maki-immich.apk

[group("android")]
install-android:
	adb install maki-immich.apk

[group("mobile-on-desktop")]
start-mobile-on-desktop +args:
	go run ./mobile {{args}}

# for windows using mingw. drag images into a single binary
[group("mobile-on-desktop")]
build-mobile-on-desktop:
	CGO_ENABLED=1 GOOS=windows \
	CXX=x86_64-w64-mingw32-g++ CC=x86_64-w64-mingw32-gcc \
	go build -ldflags -H=windowsgui -o maki-immich-mobile.exe ./mobile

# for scraping on desktop
[group("webext")]
build-webext:
	cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" ./webext

	CGO_ENABLED=0 GOOS=js GOARCH=wasm \
	go build -o ./webext/maki-immich-scrape.wasm ./webext

	cd webext && zip -r ../maki-immich-scrape.zip \
	*.js *.wasm icon.png manifest.json
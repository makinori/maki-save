default:
	@just --list

[group("desktop")]
build:
	CGO_ENABLED=0 go build -o maki-immich ./desktop

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

[group("desktop-mobile")]
start-desktop-mobile +args:
	go run ./mobile {{args}}

[group("desktop-mobile")]
build-desktop-mobile:
	go build -o maki-immich-mobile ./mobile
default:
	@just --list

[group("desktop")]
build:
	CGO_ENABLED=0 GOOS=linux go build -o maki-immich ./desktop

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
[working-directory("mobile")]
build-apk:
	go tool fyne package -os android -app-id cafe.maki.immich \
	-icon icon.png -name "maki immich"
	mv maki_immich.apk ../maki-immich.apk

[group("android")]
install-apk:
	adb install maki-immich.apk
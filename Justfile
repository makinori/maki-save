default:
	@just --list

[group("desktop")]
build-linux:
	CGO_ENABLED=0 GOOS=linux go build -o maki-save ./desktop

[group("desktop")]
install-linux msg="1":
	#!/usr/bin/env bash
	set -euo pipefail
	cp maki-save ~/maki-save
	if [[ {{msg}} == "1" ]]; then
	echo "installed to ~/maki-save"
	echo "sample config for nautilus:"
	echo "~/.local/share/actions-for-nautilus/config.json"
	echo '{
		"debug": false,
		"actions": [
			{
				"type": "command",
				"label": "maki save",
				"use_shell": true,
				"command_line": "NAUTILUS=1 ~/maki-save \"%U\"",
				"min_items": 1,
				"filetypes": ["file"]
			}
		]
	}'
	fi

# on desktop without emulator
[group("android")]
start-android:
	INTENT_TEST=1 go run ./mobile 

[group("android")]
[working-directory("mobile")]
build-android:
	go tool fyne package -os android -app-id cafe.maki.save \
	-icon ../icon/icon128.png -name "maki save" -release
	mv maki_save.apk ../maki-save.apk

[group("android")]
install-android:
	adb install maki-save.apk

[group("mobile-on-desktop")]
start-mobile-on-desktop +args:
	go run ./mobile {{args}}

# for windows using mingw
[group("mobile-on-desktop")]
build-mobile-on-desktop:
	CGO_ENABLED=1 GOOS=windows \
	CXX=x86_64-w64-mingw32-g++ CC=x86_64-w64-mingw32-gcc \
	go build -ldflags -H=windowsgui -o maki-save.exe ./mobile

# for scraping on desktop
[group("firefox")]
build-web:
	cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" ./webext

	CGO_ENABLED=0 GOOS=js GOARCH=wasm \
	go build -o ./webext/maki-save.wasm ./webext

	zip -jrFS maki-save.xpi \
	webext/*.js webext/*.wasm webext/manifest.json \
	icon/icon48.png icon/icon128.png

# xpinstall.signatures.required to false
[group("firefox")]
install-web:
	firefox-developer-edition maki-save.xpi 

all:
	#!/usr/bin/env -S parallel --shebang --ungroup --jobs {{ num_cpus() }}
	just build-linux install-linux "0"
	just build-android
	just build-mobile-on-desktop
	just build-web install-web
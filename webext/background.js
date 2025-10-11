// dev tools in about:debugger

async function webExtAlert(tab, message) {
	// const tabs = await browser.tabs.query({
	// 	active: true,
	// 	currentWindow: true,
	// });
	browser.tabs.executeScript(tab.id, {
		code: `alert(${JSON.stringify(message)});`,
	});
}

function sanitizeDirName(name) {
	return name
		.toLowerCase()
		.replaceAll(/\s+/g, "_")
		.replaceAll(/[^a-z_]/g, "");
}

// https://github.com/sindresorhus/escape-string-regexp/blob/main/index.js
function escapeStringRegexp(string) {
	return string.replace(/[|\\{}()[\]^$+*?.]/g, "\\$&").replace(/-/g, "\\x2d");
}

async function saveFile(filename, content) {
	const items = await browser.downloads.search({
		filenameRegex: "/" + escapeStringRegexp(filename) + "$",
	});

	if (items.some(item => item.exists)) {
		return;
	}

	// TODO: if deleted from disk but still in Downloads, will save with a (1)

	const blob = new Blob([content]);
	const url = URL.createObjectURL(blob);

	await browser.downloads.download({
		url: url,
		filename,
		saveAs: false,
	});

	setTimeout(() => {
		URL.revokeObjectURL(url);
	}, 1000 * 10);
}

async function scrapeURL(tab) {
	try {
		const go = new Go();
		const { instance } = await WebAssembly.instantiateStreaming(
			fetch("maki-immich-scrape.wasm"),
			go.importObject,
		);
		go.run(instance);

		const { name, files } = await wasmScrapeURL(tab.url);
		const dirName = "maki_" + sanitizeDirName(name);

		let error = "";
		let promises = [];

		for (const file of files) {
			if (file.uiErr != "") {
				error += file.name + ": " + file.uiErr + "\n";
				continue;
			}
			const promise = (async () => {
				try {
					await saveFile(dirName + "/" + file.name, file.data);
				} catch (error) {
					error += file.name + ": " + (error.message ?? error) + "\n";
				}
			})();
			promises.push(promise);
		}

		await Promise.all(promises);

		if (error != "") {
			throw new Error(error);
		}
	} catch (error) {
		webExtAlert(tab, error.message ?? error);
	}
}

browser.browserAction.onClicked.addListener(tab => {
	scrapeURL(tab);
});

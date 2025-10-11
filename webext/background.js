// dev tools in about:debugging

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
	const blob = new Blob([content]);
	const url = URL.createObjectURL(blob);

	await browser.downloads.download({
		url: url,
		filename,
		saveAs: false,
		conflictAction: "overwrite",
	});

	setTimeout(() => {
		URL.revokeObjectURL(url);
	}, 1000 * 10);
}

async function scrapeURL(tab, url) {
	try {
		const go = new Go();
		const { instance } = await WebAssembly.instantiateStreaming(
			fetch("maki-immich-scrape.wasm"),
			go.importObject,
		);
		go.run(instance);

		const { name, files } = await wasmScrapeURL(url);
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

browser.contextMenus.onClicked.addListener((info, tab) => {
	if (info.menuItemId == "maki-immich-page") {
		scrapeURL(tab, tab.url);
	} else if (info.menuItemId == "maki-immich-link") {
		scrapeURL(tab, info.linkUrl);
	}
});

browser.runtime.onInstalled.addListener(() => {
	browser.contextMenus.create({
		id: "maki-immich-page",
		title: "scrape page",
		contexts: ["page"],
		documentUrlPatterns: ["<all_urls>"],
	});
	browser.contextMenus.create({
		id: "maki-immich-link",
		title: "scrape link",
		contexts: ["link"],
		documentUrlPatterns: ["<all_urls>"],
	});
});

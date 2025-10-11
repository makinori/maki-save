# maki immich

Tiny program to upload to Immich like an image board

-   Select an album to upload one or more images to
-   Date and time gets reset on those images
-   Many go routines. It's very fast and optimized
-   GNOME nautilus support
-   Android support that scrapes links

## how

-   Add `immich.txt` to `immich/` folder<br/>
    **Line 1:** URL to instance<br/>
    **Line 2:** API key

<!-- -   Add `nitter.txt` (url) to `scrape/` folder<br/>
    Recommend using a private instance -->

-   Add `mastofedi.txt` to `scrape/` folder<br/>
    **Line 1:** URL to your instance<br/>
    **Line 2:** access token with `read:search` scope

| platform | scraping | uploading | how                  | build                          |
| -------- | -------- | --------- | -------------------- | ------------------------------ |
| linux    |          | ✔️        | command or nautilus  | `just build install`           |
| windows  |          | ✔️        | drag images onto exe | `just build-mobile-on-desktop` |
| android  | ✔️       | ✔️        | share with app       | `just build-apkinstall-apk`    |
| firefox  | ✔️       |           | click extension      | `just build-webext`            |

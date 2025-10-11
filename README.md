# maki immich

Tiny program to upload to Immich like an image board

-   Select an album to upload one or more images to
-   Date and time gets reset on those images
-   Many go routines. It's very fast and optimized
-   GNOME nautilus support
-   Android support that scrapes links

## how

### requirements

-   Add `immich.txt` to `immich/` folder<br/>
    **Line 1:** URL to instance<br/>
    **Line 2:** API key

<!-- -   Add `nitter.txt` (url) to `scrape/` folder<br/>
    Recommend using a private instance -->

-   Add `mastofedi.txt` to `scrape/` folder<br/>
    **Line 1:** URL to your instance<br/>
    **Line 2:** access token with `read:search` scope

### building

-   **For Linux desktop** make sure `kdialog` is installed<br/>
    `just build install` will write to `~/maki-immich`<br/>
    and print GNOME nautilus instructions

-   **For Windows desktop,** to drag files onto an exe<br/>
    `just build mobile-on-desktop`

-   **For Android**, use share to scrape urls/files and upload<br/>
    `just build-apk install-apk`

-   **For Firefox**, to scrape urls to your downloads folder<br/>
    `just build-webext`

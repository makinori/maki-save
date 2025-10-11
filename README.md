# maki save

scrape and save images to upload to immich

-   select an album to upload one or more images to
-   date gets reset so images appear on top
-   many go routines. it's very fast
-   gnome nautilus support to upload many
-   android support that scrapes and uploads
-   web extension that also scrapes

this program is for my own leisure. you're on your own

<table>
    <td valign="top">
        linux
        <br/><br/>
        <img height="312" src="screenshots/linux-nautilus.png" />
        <img height="176" src="screenshots/linux-dialog.png" />
    </td>
    <td valign="top">
        windows and android
        <br/><br/>
        <img height="312" src="screenshots/android.jpg" />
    </td>
</table>

## how

-   add `immich.txt` to `immich/` folder<br/>
    **line 1:** url to instance<br/>
    **line 2:** api key

<!-- -   Add `nitter.txt` (url) to `scrape/` folder<br/>
    Recommend using a private instance -->

-   add `mastofedi.txt` to `scrape/` folder<br/>
    required for pulling from any activitypub source<br/>
    **line 1:** url to your mastodon instance<br/>
    **line 2:** access token with `read:search` scope

| platform | scrape | upload | how                  | build                            |
| -------- | ------ | ------ | -------------------- | -------------------------------- |
| linux    |        | ✔️     | command or nautilus  | `just build-linux install-linux` |
| windows  |        | ✔️     | drag images onto exe | `just build-mobile-on-desktop`   |
| android  | ✔️     | ✔️     | share with app       | `just build-apk install-apk`     |
| firefox  | ✔️     |        | click extension      | `just build-webext`              |

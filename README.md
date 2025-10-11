# maki immich

tiny program to upload to immich like an image board

-   select an album to upload one or more images to
-   date and time gets reset on those images
-   many go routines. it's very fast
-   gnome nautilus support to upload many
-   android support that scrapes and uploads
-   web extension that also scrapes

this program is for my own leisure. you're on your own

<table>
   <td valign="top">
      linux
      <br/><br/>
      <img height="312" src="https://github.com/user-attachments/assets/a5565173-9914-44a3-94a2-af23a0db9002" />
      <img height="176" src="https://github.com/user-attachments/assets/79644502-8db9-4ebd-a1b7-799113ced909" />
   </td>
   <td valign="top">
      windows and android
      <br/><br/>
      <img height="312" alt="Screenshot_20251010-205642" src="https://github.com/user-attachments/assets/f51bc743-3c81-47f0-9dd8-4910da652850" />
   </td>
</table>

## how

-   add `immich.txt` to `immich/` folder<br/>
    **line 1:** url to instance<br/>
    **line 2:** api key

<!-- -   Add `nitter.txt` (url) to `scrape/` folder<br/>
    Recommend using a private instance -->

-   add `mastofedi.txt` to `scrape/` folder<br/>
    **line 1:** url to your mastodon instance<br/>
    **line 2:** access token with `read:search` scope

| platform | scrape | upload | how                  | build                          |
| -------- | ------ | ------ | -------------------- | ------------------------------ |
| linux    |        | ✔️     | command or nautilus  | `just build install`           |
| windows  |        | ✔️     | drag images onto exe | `just build-mobile-on-desktop` |
| android  | ✔️     | ✔️     | share with app       | `just build-apkinstall-apk`    |
| firefox  | ✔️     |        | click extension      | `just build-webext`            |


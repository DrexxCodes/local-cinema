# Changelog

## 1.0.0

### Control panel (Windows)
- Opening Local Cinema on Windows now shows a small native control panel: server status, the address for this computer and the address guests use on phones and TVs, Start and Stop buttons, Open admin, Copy guest link, Open logs folder, and a live log of what the server prints (including crash messages).
- The server runs as a hidden copy of the program (`local-cinema.exe serve`), so Stop and Start really restart it. A Windows job object makes sure it never keeps running after the panel closes, even after a crash. Stopping also ends FFmpeg jobs it started.
- Closing the panel stops the cinema (it asks first when the server is running). Minimize the window to keep it running. Starting the program again brings the open panel to the front.
- On a fresh install the setup page opens in the browser by itself once the server is up. Starting with Windows opens the panel minimized (`--background`).
- Every start up problem now shows a message box and is written to `local-cinema.log` (for example the port is already taken). Fixes: the installed program started but showed nothing and gave no sign of why.
- Replaces the tray icon and app window tried earlier in this release, which did not show up. No tray icon or WebView2 is needed, so `fyne.io/systray` and `go-webview2` are not dependencies. 
- Installer: shortcuts start the program itself and the setup page is opened by the program, not the installer.
- Linux and macOS are unchanged. Set `CINEMA_NO_DESKTOP=1` to run the Windows build as a plain server.
- New: `internal/desktop/`, `flags.go`. Changed: `main.go`, `paths.go`.


## 0.5.1

### Local Windows build
- Fix: running Inno Setup by hand failed with `Source file ...\dist\local-cinema-gui.exe does not exist`. That file is only made by the release workflow.
- New: `scripts/build-windows.ps1` builds the UI, embeds the icon and version info, builds `local-cinema.exe` and `dist\local-cinema-gui.exe`, and with `-Installer` runs Inno Setup. Local installers skip the FFmpeg download (install it from Settings or with winget).
- README: added the one-command build steps and a note about the build order.

## 0.5.0

### Installer, releases and updates
- Windows installer (`LocalCinema-Setup-x.y.z.exe`, built with Inno Setup). It installs the program, keeps data in `C:\ProgramData\Local Cinema` (updates and uninstalls leave it alone), opens the Windows firewall for the cinema (TCP 8765) and discovery (UDP 5353), offers start with Windows and a desktop shortcut, adds Start Menu entries (Open Local Cinema, Stop Local Cinema) and opens the setup page when finished.
- FFmpeg is not bundled. During setup a page reading "We are installing some necessary components for Local Cinema" shows a progress bar while FFmpeg downloads from this project's own release, checked against a SHA-256 checksum. If the download fails you can retry, or continue and install it later from Settings.
- Uninstalling asks whether to delete your data. Installed copies run without a console window and write a log to `C:\ProgramData\Local Cinema\logs`.
- Portable archives for Windows, Linux (amd64, arm64) and macOS (amd64, arm64), with `checksums.txt`.
- GitHub Actions: CI on every pull request, a Release workflow on version tags (builds the UI, every platform, the installer and publishes the release using this changelog as notes), an FFmpeg bundle workflow, and a monthly link check.
- App icon: placeholder icons in `packaging/icon/` (replace with your logo), favicon and Apple touch icon for the web UI, icon and version info embedded in the Windows program.
- New: `.github/workflows/`, `packaging/`, `internal/version/`, `paths.go`.

### Update checks
- On start (and once a day) the server asks GitHub for the latest release. Settings, About and updates shows the current version, a Check for updates button, release notes, and a download button for the right file for the computer. A chip in the admin header says when an update exists. Development builds don't check. Pre-releases are not offered.

### FFmpeg
- Found in this order: `CINEMA_FFMPEG_DIR`, the `bin` folder next to the data folder, the program's folder, the PATH.
- Settings, FFmpeg shows whether it is installed and its version, with an Install FFmpeg button (Windows and Linux) that downloads and verifies it with a progress bar.

### Sign in
- Show and hide button on password fields: guest sign in, admin sign in, setup, and the admin account form.

### Project files
- `LICENSE.md` (MIT), `SECURITY.md`, `CONTRIBUTING.md`, `THIRD_PARTY_NOTICES.md`.
- Not compiled or run here. New Go tests cover version comparison, subtitle timing and SRT conversion.

## 0.4.1

### First-time player guide
- When a movie starts, the browser value `isFirstTime` is created as `true` if it does not exist. While it is `true` the guide shows; finishing or skipping it sets it to `false`. Clear the value in the browser (or set it to `true`) to see the guide again.
- The movie pauses during the guide and carries on afterwards. Each step lights up one control and dims the rest: play and pause, back and forward 10 seconds, volume, the seek bar, speed, subtitles, rotate, full screen and More, then touch shortcuts on phones. Previous and Next arrows, step bars, "Skip tour", and the arrow and Escape keys all work.
- New: `Tour.tsx`, `constants/tour.ts`, `hooks/useFirstTimeGuide.ts`. Controls carry `data-tour` markers.

### Fixed: stream dialog misbehaving while a movie plays
- The dialog lived inside the player's menu, whose blur effect made `position: fixed` relative to the small menu instead of the screen, so the dialog was cut off and misplaced. It now renders into the player itself (so it also shows in full screen and rotates with the picture), fits small screens, closes with Escape, and the movie pauses while it is open and resumes afterwards.
- Player shortcut keys no longer fire while a dialog is open.

### iPhone full screen
- iPhone full screen switches to Apple's own controls. Before entering it, iPhone users now see a note: turn off Portrait Orientation Lock in Control Center to watch sideways. They can continue, or use Rotate to keep the custom controls, and can tick "Don't show this again" (`iosFullscreenHintDismissed`). The guide's full screen step mentions it too.

### Scroll bars
- Scroll bars use the accent colour (thin, with a faint track), with a fallback for older browsers. The movie carousels still hide theirs.

## 0.4.0

### Custom video player
- Replaces the browser's built-in controls. The seek bar, buttons and spinner use your accent colour; the lighter part of the seek bar shows what has buffered.
- Back and forward 10 seconds (buttons, double-tap the left or right side on a phone, arrow keys), playback speed, volume with mute, a subtitles menu, full screen, picture in picture and AirPlay where the device supports them.
- Rotate: turns the picture sideways to fill the screen, so it works on iPhones too. Tap again for the other direction, and once more to go back. On Android, full screen also tries to lock to landscape.
- Controls hide while playing and return on a tap or mouse move. Keyboard: space or K, arrows, J and L, M, F, C.
- Shows a buffering spinner. When a movie can't be played, the player offers "Stream on an external app".
- New: `ui/components/player/` (Player, Controls, PlayerMenu, SeekBar), `ui/hooks/usePlayer.ts`.

### Stream URLs for other apps (VLC, Infuse, nPlayer and so on)
- "Copy stream URL" on the movie page, in the player's More menu, and when a movie can't be played. It shows the address with a Copy button (works on plain http addresses too), plus a link for each subtitle track.
- Admin chooses per room whether this is allowed: a switch when creating rooms (off by default) and in the room editor afterwards. Turning it off ends that room's open stream URLs.
- Every link carries its own session id and counts against the room's connection limit. One device keeps one stream URL session, so asking again returns the same link. These sessions show in Admin, Sessions as "Stream URL". They stay open while a player is reading the file and for 30 minutes after it stops.
- Links use the original file (other apps play most formats) and the host's IP address, because `.local` names often don't resolve in other apps.
- At the limit the guest sees "You have reached the maximum connection for this profile" with the connected devices (they can disconnect one) and a Request Upgrade button. They enter how many extra connections they need.
- Admin, new Requests tab (with a count badge): see each request with the room's current usage, change the number, then Approve (adds that many connections) or Decline. Earlier requests are listed.
- New: `external.go`; new `upgrade_requests` collection; `allow_external` on rooms and `kind` on sessions are added automatically.

### Subtitles
- Several tracks per movie. Library tab, Edit movie, Subtitles: upload .srt, .vtt, .ass or .ssa files (up to 2 MB, old Windows encodings are converted) with a language and a name, rename them, choose which track starts selected (or none), shift timing by half a second steps, delete.
- "Extract from video" reads the text subtitle tracks out of MKV and MP4 files (picture based DVD and Blu-ray tracks can't be converted).
- Sidecar files next to the movie still work and are listed. They are never modified.
- Uploaded and extracted tracks live in `pb_data/cache/subs/`.
- New: `admin_subs.go`, `ui/components/admin/SubtitleManager.tsx`. `subtitles.go` was rewritten.

### API
- `POST /api/movies/{id}/stream-url`, `GET /api/sessions`, `DELETE /api/sessions/{id}`, `POST /api/upgrade-requests`
- `GET /api/movies/{id}/subtitles` now returns `key`, `lang`, `label`, `default`, `source`. `GET /api/movies/{id}/subtitles/{key}`.
- Admin: `GET|POST /api/admin/movies/{id}/subtitles`, `POST .../subtitles/extract`, `PATCH|DELETE .../subtitles/{key}`, `GET /api/admin/upgrade-requests`, `PATCH /api/admin/upgrade-requests/{id}`.
- `GET /api/admin/stats` adds `subs`, `subs_msg`, `pending_upgrades`.
- Not compiled or run here. No new dependencies.

## 0.3.2

### Fixed: Library tab said "Could not list movies."
- The movie list sorted with `title COLLATE NOCASE`, which the query builder quotes as one column name, so the database rejected the query. The list now uses PocketBase's own filter and sort. Failures now show the real database error and are logged, for both the movie and user lists.

### Every streamable file is listed
- The scanner now also lists WMV, FLV, MPG/MPEG, TS, M2TS, 3GP and OGV files. They can't play directly, but "Make compatible" turns them into playable MP4 copies.
- If a scan skips a file (unreadable, or it can't be saved), or finds no videos at all, the Library tab says which file and why.

### Edit movie details
- Library tab, Edit under any movie: title, year, rating, genres, synopsis, poster and banner upload (JPEG, PNG or WebP up to 10 MB), and a "Look up on TMDB" button that uses the current title and year, so a wrong title can be fixed first.
- Edited movies are protected: rescans never overwrite your text or pictures with TMDB data.

### Remove movies from the stream
- Select movies, then "Hide from guests" or "Show to guests". Hidden movies disappear from the guest catalog and can't be opened or streamed. A rescan never un-hides them.
- Filter: All movies, Hidden from guests, File missing. Movies whose file is gone can be removed from the list for good ("Remove missing from list"). Movies with a file on disk can only be hidden, since a rescan would add them back.

### API
- `PATCH /api/admin/movies/{id}`, `DELETE /api/admin/movies/{id}` (missing files only), `POST /api/admin/movies/{id}/tmdb`, `PUT /api/admin/movies/{id}/art/{poster|backdrop}` (raw image body), `POST /api/admin/movies/visibility`.
- `GET /api/admin/movies` takes `show=hidden|missing` and returns `hidden`, `edited`, `available`, `tmdb_id`, `file_name`. `GET /api/admin/stats` returns `scan_issues`.
- New movie fields `hidden` and `edited` are added automatically. Not compiled here.

## 0.3.1

### Make compatible (fixes "This device can't play this format")
- Admin, Library tab: movies that are not H.264 video in an MP4 file are marked "May not play". Select them (or use "Select unsupported") and press "Make compatible", or "Make all compatible" for everything.
- The server saves a universal MP4 copy in `pb_data/cache/compat/` and guests are served that copy automatically. Your original file is never touched and the copy is not picked up by the scanner.
- H.264 video is copied untouched (seconds to minutes, no quality loss) and only the container and audio change. Other video (HEVC and so on) is re-encoded to H.264, which is slow. Audio that is not AAC becomes stereo AAC.
- Live progress with a percentage and a Cancel button. Jobs run in the background, one movie at a time, and thumbnails or scans can run alongside.
- "Remove converted copy" switches a movie back to its original file. Replacing a source file discards its stale copy. Half written copies are cleaned up at startup.
- Player message for unsupported formats now points to the host.
- New: `internal/cinema/compat.go`, `admin_compat.go`; UI `CompatBadge.tsx`, `JobProgress.tsx`. No schema change, no new dependencies.
- API: `POST /api/admin/compat` with `{"ids":[...]}` or `{"all":true}`, `POST /api/admin/compat/cancel`, `DELETE /api/admin/compat/{id}`. `GET /api/admin/movies` items carry `compat`, `GET /api/admin/stats` carries `convert` and `convert_error`.
- Not compiled here, like 0.3.0.

## 0.3.0

### Thumbnails
- Movies without TMDB art now get a poster (2:3) and a wide banner extracted from the video with `ffmpeg`. One decode produces both, the `thumbnail` filter skips black or faded frames, and portrait phone videos are centre cropped. Frame position defaults to about 12% into the film.
- Runs during every scan, and once at startup for movies that already exist without art. A failed extraction is not retried on every scan.
- TMDB art is never overwritten by a frame, and a failed TMDB download no longer hides an extracted frame (`has_poster` is only set when a download works).
- Admin, Library tab: pick movies, optionally enter the moment to use (`90`, `1:30`, `1:02:03`), and regenerate. "Create missing thumbnails" fills the rest. Progress shows live.
- New `thumb_version` field on movies busts the browser cache when art changes. Images are cached for 7 days.
- New `internal/cinema/thumbs.go`. Requires `ffmpeg` on the host (the Library tab warns when it is missing).

### Guest UI (Netflix style layout)
- Full width hero carousel (5 featured movies, auto rotates, Watch and Add list buttons, indicator bars).
- Trends Now with Popular, Premieres and Recently Added tabs, multi select genre pills and an arrow carousel.
- Movies section with genre pills, sort by Latest, Year or A-Z (tap the active one to reverse), minimum rating slider, responsive grid, "+" button and load more on scroll.
- Wide "Watching" banner with progress for the most recent unfinished movie, plus a Continue watching row.
- Home, Movies and My list views. My list is stored on the device, one list per room.
- Transparent header that turns to frosted glass on scroll, search, signal indicator and account menu.
- Cards show year, heart, watched eye and rating.
- Bold sans headings. Button text switches between black and white to stay readable on any accent colour (neon green was white on green before).

### Admin
- New Users tab: every room profile, 20 at a time, loads more as you scroll, with search. Shows devices in use against the limit, edit (name, devices, notes, new password, enable or disable) and delete.
- New Settings tab: everything from the setup wizard is editable (name, tagline, accent colour, movie folders with validation, folder watching, device limit, network name, TMDB key, admin email and password). Folder changes trigger a rescan, name or network name changes re-announce mDNS.
- Footer on every page except the player: copyright, current year and cinema name, "All rights reserved".

### Signal strength monitor
- Each device measures its round trip to the server every 4 seconds (`GET /api/ping`) and turns the last 12 results into latency, jitter and loss, shown as 1 to 4 bars.
- Bars appear in the guest header and on the player. Tap them for a latency graph.
- The measurement is sent with the heartbeat (now every 15 seconds), so Admin, Sessions shows the signal and latency of every device that is watching, plus the average and the admin's own link.
- New `rtt_ms` and `signal` fields on sessions are added automatically to existing databases.

### Structure
- UI split into `types/`, `constants/`, `lib/` (api, session, device, format, movies, adminApi, signal), `hooks/` (useGuest, useCatalog, useMyList, useSignal, useCinema, usePagedList, useSentinel, useDebounced) and `components/` (layout, home, signal, admin, ui).
- The admin page is now a thin shell over one component per tab.
- Server side admin code split into `admin_users.go`, `admin_library.go` and `admin_settings.go`.

### API additions
- `GET /api/ping`
- `GET /api/admin/users?limit&offset&q`, `PATCH` and `DELETE /api/admin/users/{id}`
- `GET /api/admin/movies?limit&offset&q`, `POST /api/admin/thumbnails`
- `GET` and `PUT /api/admin/settings`, `PUT /api/admin/account`, `POST /api/admin/validate-folder`
- `GET /api/admin-img/{posters|backdrops}/{id}?token=`
- `GET /api/admin/stats` now includes `thumbs` and `ffmpeg`.


## 0.2.0

- Folder watching (fsnotify) with 10 second debounce, auto-watching new sub-folders, re-sync when `movie_folders` changes
- Sidecar subtitles (.srt and .vtt, language from file name), SRT converted to WebVTT, tracks added to the player
- Admin page at `/admin/`: superuser sign in, live sessions with End session, stats, bulk room creation, printable QR room cards, manual rescan
- Login rate limit of 10 attempts per minute per IP (`rate_limited` error, separate from the device cap modal)
- `stream_start` audit events
- Setup wizard now links to the admin page
- README: Go, Node and ffprobe setup per OS, firewall notes, subtitles, troubleshooting
- New dependencies: `github.com/fsnotify/fsnotify` (run `go mod tidy`), `qrcode` (npm)

## 0.1.0 (initial scaffold)

- Go server with embedded PocketBase and an embedded Next.js static export
- Collections: settings, movies, users (rooms), sessions, playback_state, audit_log
- Login with per-room device cap, global cap, device_id session reuse, heartbeat, 2 minute reaper
- Direct Play streaming via `http.ServeContent` (range requests) with symlink-safe path checks
- Library scanner: filename parsing, optional ffprobe, optional TMDB metadata with poster and backdrop cache
- mDNS advertising (`_localcinema._tcp`) plus `<name>.local` host record
- Admin API: rescan, list and end sessions, stats, bulk room creation
- Setup wizard, login, home (hero, continue watching, recent, genres, search), movie detail, player with resume



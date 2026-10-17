# internal/cinema

One folder per piece of logic. Feature packages expose a `Service` with `Mount(httpx.Routes)`;
`cinema.go` and `routes.go` in this folder create them and connect them.

| Layer | Package | What it holds |
|---|---|---|
| Pure helpers | `util`, `progress`, `ratelimit`, `netinfo`, `cache`, `media`, `tmdb`, `ffmpeg` | No PocketBase. Plain Go, fully unit tested. |
| Shared base | `httpx` | JSON error shape, `Routes` (public / admin / guest), `Guard` type. |
| | `store` | The app, the settings record, cache layout, FFmpeg tools, audit log, collections. |
| Features | `subtitles` | Index, sidecar files, SRT to VTT, upload, extract, guest endpoints. |
| | `thumbs` | Poster and backdrop frames, thumbnail job. |
| | `compat` | "Make compatible" MP4 conversion job. |
| | `scanner` | Library scan, folder watcher, issues, TMDB metadata. |
| | `sessions` | Device sessions, limits, stream URLs, upgrade requests, reaper. |
| | `auth` | Login, logout, heartbeat, `/me`, the guest guard. |
| | `rooms` | Room accounts: bulk create, list, edit, delete. |
| | `settings` | Admin settings, validation, admin account. |
| | `setup` | First run wizard and public cinema info. |
| | `catalog` | Movie list for guests, movie editing and artwork for the admin. |
| | `stream` | Range streaming and playback position. |
| | `updates` | Release check against GitHub. |
| | `tools` | FFmpeg status and installer for the Settings page. |
| | `mdns` | `cinema.local` announcement. |
| | `admin` | Dashboard stats. |

Dependencies point downwards only (features use helpers and `store`, never the other way).
Where two features need each other the link is a callback set in `newApp`
(`thumbs.Busy`, `settings.OnIdentityChanged`).

## Tests

`go test ./...`. Tests cover pure logic, files, `httptest` servers, and FFmpeg through a fake
program (`CINEMA_FAKE_FFMPEG`). Handlers that need a real PocketBase record are exercised by
running the app.

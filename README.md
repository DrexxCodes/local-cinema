# Local Cinema

Local Cinema is a Self-hosted, LAN-only movie streaming. It is direct Play only, no transcoding.

## Install (no tools needed)

Download the latest release from the [Releases page](../../releases).

- **Windows:** run `LocalCinema-Setup-x.y.z.exe`. It installs the program, opens the firewall for your network, offers to start with Windows, and downloads the video tools (FFmpeg) while installing. Then open the setup page and follow the wizard.
- **Linux / macOS:** unpack the archive for your system and run `./local-cinema`. Open Settings to install FFmpeg on Linux, or run `brew install ffmpeg` on macOS.
- **Updates:** the admin Settings page shows the current version and checks GitHub for a newer one. Run the new installer over the old one, your data is kept (Windows keeps it in `C:\ProgramData\Local Cinema`).

Building from source is described below. See [CONTRIBUTING.md](CONTRIBUTING.md), [SECURITY.md](SECURITY.md) and [LICENSE.md](LICENSE.md).

## Build from source

## 1. Install the tools

### Go (1.23 or newer)

**Windows**
1. Download the MSI installer from https://go.dev/dl/ and run it. It adds Go to your PATH.
2. Close and reopen your terminal (PowerShell or Command Prompt).

**macOS**
- Download the `.pkg` from https://go.dev/dl/ and run it, or with Homebrew: `brew install go`

**Linux (Debian/Ubuntu and others)**
```bash
# Pick the current version number from https://go.dev/dl/
wget https://go.dev/dl/go1.XX.X.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.XX.X.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
source ~/.profile
```
(Use the `arm64` archive on ARM machines.)

**Verify**
```bash
go version      # should print go1.23 or newer
```

Notes:
- No C compiler is needed. PocketBase uses a pure Go SQLite driver.
- If `go.mod` asks for a newer Go than you have, Go 1.21+ downloads the right toolchain automatically (internet needed on the build machine only). The finished binary needs no internet.
- If `go` is "not found" after installing, your PATH was not refreshed. Restart the terminal or log out and in.
- Behind a slow network or proxy: `go env -w GOPROXY=https://proxy.golang.org,direct`

### Node.js and npm

Install the current LTS from https://nodejs.org (or `brew install node`, or your Linux package manager). Check with `node -v` and `npm -v`.

### Optional

- `ffmpeg` and `ffprobe` on PATH. ffprobe gives duration and codec info, ffmpeg extracts thumbnails from the videos: `winget install Gyan.FFmpeg`, `brew install ffmpeg`, or `sudo apt install ffmpeg`
- A free TMDB API key (https://www.themoviedb.org/settings/api) for posters and synopses

## 2. Build and run

```bash
make build      # installs UI deps, builds the static export, then builds the Go binary
./local-cinema  # serves on 0.0.0.0:8765 (override with CINEMA_PORT)
```

On Windows without `make`, run the steps by hand:

```powershell 
cd ui
npm install
npm run build
cd ..
go mod tidy
go build -o local-cinema.exe .
.\local-cinema.exe
```

To build the Windows program and installer in one go, run this from the project folder (needs Go, Node.js or pnpm, and [Inno Setup 6](https://jrsoftware.org/isdl.php) for the installer):

```powershell
.\scripts\build-windows.ps1 -Installer
```

It writes `local-cinema.exe` (with a console), `dist\local-cinema-gui.exe` (no console, packed into the installer) and `dist\installer\LocalCinema-Setup-<version>.exe`. Do not run `ISCC.exe` on its own first, it needs `dist\local-cinema-gui.exe` to exist.

On Windows the program opens a small control panel that starts the cinema and shows its status, the address for this computer, the address guests type on their phones and TVs, and a live log. Use Stop server and Start server there, and Open admin to reach the admin page in your browser (on a fresh install the setup page opens by itself). Closing the panel stops the cinema, so minimize it to keep it running. The log is also saved in `C:\ProgramData\Local Cinema\logs` when installed (a `logs` folder next to the program otherwise). Set `CINEMA_NO_DESKTOP=1` to run it as a plain server, and use `--background` to start the panel minimized.

On Linux and macOS, open `http://localhost:8765/setup/` and follow the wizard.

The UI uses the latest stable Next.js (`"next": "latest"`). After the first `npm install`, commit `package-lock.json` to pin what you got.

**Firewall:** allow inbound TCP 8765 (and UDP 5353 for mDNS) on the host, or other devices will not connect. On Windows, accept the firewall prompt the first time the binary starts.

## 3. Everyday use

- Guests: `http://cinema.local:8765` or `http://<host-ip>:8765` (Android often needs the IP). Username is the room name (`room204`) or the full email.
- Admin page: `http://<host-ip>:8765/admin/` with the superuser login from the wizard. It has live sessions with signal strength (and End session), all rooms (20 at a time, edit or delete), room creation with printable QR cards, library rescan with thumbnail generation, and a Settings tab for everything the setup wizard asked.
- Database dashboard: `http://<host-ip>:8765/_/` for editing settings, users and everything else directly.

### Rooms and QR cards

In the admin page, open Rooms, choose a prefix and range, and create. Passwords are shown once. Use Print room cards for a sheet with a QR code, username, password and address per room.

### Subtitles

Put `.srt` or `.vtt` files next to the movie, named like the video:

```
Inception (2010).mp4
Inception (2010).srt          shown as "Subtitles"
Inception (2010).en.srt       shown as "en"
Inception (2010).fr.vtt       shown as "fr"
```

SRT is converted to WebVTT on the fly. Use UTF-8 files. Pick a track from the player's captions menu.

### Folder watching

With `watch_folders` on (default), adding, removing or renaming videos triggers a rescan about 10 seconds after the changes stop. Changing `movie_folders` in the dashboard re-syncs watches and rescans.

## 4. API quick reference

Admin calls need a superuser token in the `Authorization` header.

- `POST /api/admin/scan`, `GET /api/admin/stats`
- `GET /api/admin/users?limit=20&offset=0&q=`, `PATCH|DELETE /api/admin/users/{id}`
- `GET /api/admin/movies`, `POST /api/admin/thumbnails` with `{"ids":["..."],"at":90,"force":true}` or `{"all":true}`
- `GET|PUT /api/admin/settings`, `PUT /api/admin/account`
- `GET /api/ping` (public, used by the signal monitor)
- `GET /api/admin/sessions`, `DELETE /api/admin/sessions/{id}`
- `POST /api/admin/rooms/bulk` with `{"prefix":"room","from":101,"to":120,"max_connections":2,"password":""}`

## 5. Development

```bash
go run . serve --http=0.0.0.0:8765      # terminal 1
cd ui && npm install && npm run dev     # terminal 2, proxies /api and /stream to :8765
```

## 6. Troubleshooting

- **Black screen or "can't play this format":** the browser does not support the codec. H.264 video with AAC audio in MP4 works everywhere. MKV, HEVC, AC3 and DTS are hit and miss.
- **Seeking is slow on an MP4:** the file has its index at the end. Fix with `ffmpeg -i in.mp4 -c copy -movflags +faststart out.mp4`.
- **`cinema.local` does not resolve:** use the IP address.
- **Port in use:** set `CINEMA_PORT=9000` before starting.
- **Reset everything:** stop the app and delete the `pb_data` folder.


## Thats a wrap...

We sure hope you have as much fun using this as we had building it. Cheers.

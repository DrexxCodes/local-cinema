# Third-party notices

Local Cinema is released under the MIT License (see `LICENSE.md`). It is built on
open source software, each part under its own licence.

## Included in Local Cinema

| Component | Licence |
| --- | --- |
| [PocketBase](https://github.com/pocketbase/pocketbase) and [dbx](https://github.com/pocketbase/dbx) | MIT |
| [zeroconf](https://github.com/grandcat/zeroconf) | MIT |
| [fsnotify](https://github.com/fsnotify/fsnotify) | BSD-3-Clause |
| [modernc.org/sqlite](https://gitlab.com/cznic/sqlite) and related libraries | BSD-3-Clause |
| [Next.js](https://github.com/vercel/next.js), [React](https://github.com/facebook/react), [Tailwind CSS](https://github.com/tailwindlabs/tailwindcss), [qrcode](https://github.com/soldair/node-qrcode) | MIT |

The full list of Go and JavaScript dependencies, with versions, is in `go.mod`
and `ui/package.json`. Each dependency keeps its own licence and copyright notice.

## Downloaded separately: FFmpeg

Local Cinema uses [FFmpeg](https://ffmpeg.org) for thumbnails, converting movies and
reading subtitles. FFmpeg is **not part of Local Cinema and is not included in its
files**. The Windows installer and the Settings page download an unmodified build
from this project's releases (release name `ffmpeg-<bundle>`).

FFmpeg is licensed under the GNU General Public License (the build used here includes
GPL components such as x264). Local Cinema runs FFmpeg as a separate program and
does not link to it. The licence text, source links and exact download details are
in `SOURCE.txt` and `LICENSE.txt` inside each downloaded FFmpeg zip, and in the
installed `bin` folder next to it.

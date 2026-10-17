# Contributing to Local Cinema

Thanks for helping. This guide covers setting up, how the code is organised, and what a
good pull request looks like. By contributing you agree that your work is released under
the project's [MIT License](LICENSE.md).

## Ground rules

- Be kind and patient. Questions from beginners are welcome.
- For anything bigger than a small fix, open an issue first so we can agree on the idea.
- Report security problems privately, see [SECURITY.md](SECURITY.md).

## Set up

You need:

- **Go** (the version in `go.mod`)
- **Node.js 22 or newer** and **pnpm**
- **FFmpeg** (needed for thumbnails, Make compatible and subtitle extraction). The app can
  download it for you from Settings once you have a release build, or install it yourself
  (`winget install Gyan.FFmpeg`, `brew install ffmpeg`, `sudo apt install ffmpeg`).

```bash
git clone <your fork>
cd local-cinema
cd ui && pnpm install && cd ..
```

## Run it while developing

Use two terminals so both parts reload.

```bash
# terminal 1, the server on port 8765
go run . serve --http=0.0.0.0:8765

# terminal 2, the UI on port 3000 (it forwards /api and /stream to the server)
cd ui && pnpm dev
```

Open http://localhost:3000. The first start shows the setup wizard. The database lives
in `pb_data/` next to where you run the program. Delete that folder to start over.

To build the single program with the UI inside:

```bash
cd ui && pnpm build && cd ..
go build -o local-cinema .
```

## How the code is organised

To keep things easy to maintain keep the code modular: small components, logic in hooks and `lib`, shared shapes in
`types`, fixed values in `constants`. Please do not grow one file to hold everything. Mauled up code will be rejected

## Before you open a pull request

1. Run `go vet ./...` and `go test ./...`.
2. Run `pnpm build` in `ui` (it type-checks).
3. Try the change in a browser, on a phone if it touches the player or layout.
4. Add an entry to `CHANGELOG.md` under a new or the newest heading, in plain words:
   what changed for the person using the app.
5. Describe the change and how you tested it in the pull request.

Small, focused pull requests are reviewed fastest.

## Releases (maintainers)

1. Make sure `packaging/ffmpeg-bundle.txt` names an FFmpeg bundle that exists. To publish a
   new one, run the **FFmpeg bundle** workflow, then update that file.
2. Add the version's section to `CHANGELOG.md` (`## 1.2.0`). It becomes the release notes.
3. Tag and push: `git tag v1.2.0 && git push origin v1.2.0`.
4. The **Release** workflow builds the program for every platform and the Windows
   installer, then publishes the GitHub Release. Tags with a dash (`v1.2.0-beta.1`) are
   published as pre-releases and are not offered by the in-app update check.

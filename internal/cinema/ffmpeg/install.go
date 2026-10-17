package ffmpeg

import (
	"archive/zip"
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// The download lives in a release of this project (tag ffmpeg-<bundle>) with one zip per
// platform, named ffmpeg-<goos>-<goarch>.zip, and a SHA256SUMS file.
var platforms = map[string]bool{"windows/amd64": true, "linux/amd64": true, "linux/arm64": true}

// archiveFiles are the only names taken from the zip. They are written flat into Dir.
var archiveFiles = map[string]bool{
	"ffmpeg": true, "ffmpeg.exe": true, "ffprobe": true, "ffprobe.exe": true,
	"LICENSE.txt": true, "SOURCE.txt": true, "README.txt": true,
}

// Installer downloads FFmpeg from this project's own GitHub release and checks it
// against the SHA-256 list published next to it.
type Installer struct {
	Repo   string // GitHub "owner/name"
	Bundle string // release tag is ffmpeg-<Bundle>
	Dir    string // where the programs are unpacked

	GOOS, GOARCH string       // default to the running system
	BaseURL      string       // overrides the release download address (used by tests)
	Client       *http.Client // defaults to http.DefaultClient
	UserAgent    string

	OnStep    func(msg string) // "Downloading FFmpeg", and so on
	OnPercent func(pct int)    // download progress
}

func (i *Installer) goos() string {
	if i.GOOS != "" {
		return i.GOOS
	}
	return runtime.GOOS // exit :)
}

func (i *Installer) goarch() string {
	if i.GOARCH != "" {
		return i.GOARCH
	}
	return runtime.GOARCH
}

// Supported reports whether this build knows where to download FFmpeg for this computer.
func (i *Installer) Supported() bool {
	return i.Repo != "" && i.Bundle != "" && platforms[i.goos()+"/"+i.goarch()]
}

// Asset is the zip name for this computer.
func (i *Installer) Asset() string {
	return fmt.Sprintf("ffmpeg-%s-%s.zip", i.goos(), i.goarch())
}

func (i *Installer) base() string {
	if i.BaseURL != "" {
		return strings.TrimRight(i.BaseURL, "/") + "/"
	}
	return fmt.Sprintf("https://github.com/%s/releases/download/ffmpeg-%s/", i.Repo, i.Bundle)
}

func (i *Installer) step(msg string) {
	if i.OnStep != nil {
		i.OnStep(msg)
	}
}

func (i *Installer) get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if i.UserAgent != "" {
		req.Header.Set("User-Agent", i.UserAgent)
	}
	c := i.Client
	if c == nil {
		c = http.DefaultClient
	}
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		res.Body.Close()
		return nil, fmt.Errorf("the download server answered with status %d", res.StatusCode)
	}
	return res, nil
}

// ExpectedHash finds the checksum of asset in a SHA256SUMS file.
func ExpectedHash(sums io.Reader, asset string) (string, error) {
	sc := bufio.NewScanner(io.LimitReader(sums, 64<<10))
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		if len(f) == 2 && strings.TrimPrefix(f[1], "*") == asset {
			return strings.ToLower(f[0]), nil
		}
	}
	return "", errors.New("the checksum list has no entry for this computer")
}

type countingWriter struct {
	n, total int64
	onPct    func(int)
}

func (c *countingWriter) Write(p []byte) (int, error) {
	c.n += int64(len(p))
	if c.total > 0 && c.onPct != nil {
		c.onPct(int(c.n * 100 / c.total))
	}
	return len(p), nil
}

// Install downloads, verifies and unpacks FFmpeg into Dir. Existing files are replaced.
func (i *Installer) Install(ctx context.Context) error {
	if !i.Supported() && i.BaseURL == "" {
		return errors.New("this build has no download for this computer. Install FFmpeg yourself (for example brew install ffmpeg) and restart")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()
	asset := i.Asset()
	base := i.base()

	i.step("Checking the download")
	sumsRes, err := i.get(ctx, base+"SHA256SUMS")
	if err != nil {
		return err
	}
	want, err := ExpectedHash(sumsRes.Body, asset)
	sumsRes.Body.Close()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(i.Dir, 0o755); err != nil {
		return err
	}
	tmp := filepath.Join(i.Dir, asset+".download")
	defer os.Remove(tmp)

	i.step("Downloading FFmpeg")
	res, err := i.get(ctx, base+asset)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	h := sha256.New()
	cw := &countingWriter{total: res.ContentLength, onPct: i.OnPercent}
	_, err = io.Copy(io.MultiWriter(f, h, cw), res.Body)
	f.Close()
	if err != nil {
		return err
	}
	if hex.EncodeToString(h.Sum(nil)) != want {
		return errors.New("the download is damaged (checksum does not match)")
	}

	i.step("Unpacking")
	return unpack(tmp, i.Dir)
}

func unpack(zipPath, dir string) error {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer zr.Close()
	found := 0
	for _, zf := range zr.File {
		name := filepath.Base(zf.Name) // flat names only, no paths from the archive are used
		if zf.FileInfo().IsDir() || !archiveFiles[name] {
			continue
		}
		if err := extractFile(zf, filepath.Join(dir, name)); err != nil {
			return err
		}
		found++
	}
	if found == 0 {
		return errors.New("the download did not contain ffmpeg")
	}
	return nil
}

func extractFile(zf *zip.File, dest string) error {
	rc, err := zf.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	part := dest + ".new"
	out, err := os.OpenFile(part, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, rc); err != nil {
		out.Close()
		os.Remove(part)
		return err
	}
	out.Close()
	if err := os.Rename(part, dest); err != nil {
		os.Remove(part)
		return fmt.Errorf("could not replace %s (is it in use?): %w", filepath.Base(dest), err)
	}
	return nil
}

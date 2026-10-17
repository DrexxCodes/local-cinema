package catalog

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/pocketbase/pocketbase/core"

	"local-cinema/internal/cinema/httpx"
	"local-cinema/internal/cinema/util"
)

// ConvertImage turns an uploaded picture into a JPEG no wider than maxWidth, using ffmpeg.
func ConvertImage(bin, src, dst string, maxWidth int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "-v", "error", "-y", "-i", src, "-frames:v", "1",
		"-vf", fmt.Sprintf("scale='min(%d,iw)':-2", maxWidth), "-q:v", "3", dst)
	if _, err := cmd.CombinedOutput(); err != nil || !util.FileExists(dst) {
		return fmt.Errorf("that image could not be read")
	}
	return nil
}

// handleAdminArt sets a poster or backdrop from an uploaded image (raw body, 10 MB at most).
// The image becomes a size limited JPEG when ffmpeg is present.
func (s *Service) handleAdminArt(e *core.RequestEvent) error {
	dir, maxW, ok := ArtKind(e.Request.PathValue("kind"))
	if !ok {
		return httpx.Err(e, 404, "not_found", "Unknown image type.")
	}
	r, err := s.st.App.FindRecordById("movies", e.Request.PathValue("id"))
	if err != nil {
		return httpx.Err(e, 404, "not_found", "Movie not found.")
	}
	body, err := io.ReadAll(http.MaxBytesReader(e.Response, e.Request.Body, 10<<20))
	if err != nil {
		return httpx.Err(e, 413, "too_large", "That image is too large (10 MB at most).")
	}
	ctype := http.DetectContentType(body)
	ext, ok := ImageExt(ctype)
	if !ok {
		return httpx.Err(e, 400, "bad_image", "Use a JPEG, PNG or WebP image.")
	}

	out := s.st.Cache.Image(dir, r.Id)
	if bin := s.st.Tools.FFmpeg(); bin != "" {
		in, tmp := out+".upload"+ext, out+".tmp.jpg"
		defer os.Remove(in)
		defer os.Remove(tmp)
		if err := os.WriteFile(in, body, 0o644); err != nil {
			return httpx.Err(e, 500, "internal", "Could not save the image.")
		}
		if err := ConvertImage(bin, in, tmp, maxW); err != nil {
			return httpx.Err(e, 400, "bad_image", "That image could not be read.")
		}
		if err := os.Rename(tmp, out); err != nil {
			return httpx.Err(e, 500, "internal", "Could not save the image.")
		}
	} else if ctype == "image/jpeg" {
		if err := os.WriteFile(out, body, 0o644); err != nil {
			return httpx.Err(e, 500, "internal", "Could not save the image.")
		}
	} else {
		return httpx.Err(e, 400, "bad_image", "Only JPEG images are accepted while FFmpeg is not installed.")
	}

	if dir == "posters" {
		r.Set("has_poster", true)
	} else {
		r.Set("has_backdrop", true)
	}
	r.Set("thumb_version", time.Now().Unix())
	r.Set("edited", true) // keep TMDB from replacing the admin's picture on a rescan
	if err := s.st.App.Save(r); err != nil {
		return httpx.Err(e, 500, "internal", "Could not save: "+err.Error())
	}
	return e.JSON(200, s.AdminDTO(r))
}

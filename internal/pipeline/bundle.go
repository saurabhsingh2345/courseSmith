package pipeline

// Offline distribution: build the static site and zip it — videos included —
// so a course runs from a USB stick via file:// with no CDN or server.

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/enfec/coursesmith/internal/project"
)

// BuildSite runs the site build: site/build.sh when present (hugo +
// pagefind), plain hugo otherwise. Returns the publish dir.
func BuildSite(ctx context.Context, e *Env) (string, error) {
	siteDir := e.siteDir()
	publicDir := filepath.Join(siteDir, "public")

	if _, err := os.Stat(filepath.Join(siteDir, "build.sh")); err == nil {
		cmd := exec.CommandContext(ctx, "sh", "build.sh")
		cmd.Dir = siteDir
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("site build.sh failed: %w\n%s", err, tailLines(out.String(), 12))
		}
		return publicDir, nil
	}

	if _, err := exec.LookPath("hugo"); err != nil {
		return "", fmt.Errorf("hugo not found — install it (macOS: brew install hugo) to build the site")
	}
	cmd := exec.CommandContext(ctx, "hugo", "--source", siteDir)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("hugo build failed: %w\n%s", err, tailLines(out.String(), 12))
	}
	return publicDir, nil
}

// BundleCourse builds the site and zips the publish dir into
// dist/<slug>-bundle.zip. Everything in the zip is relative, so extracting
// and opening index.html works from file://.
func BundleCourse(ctx context.Context, e *Env, course *project.Course) (string, error) {
	publicDir, err := BuildSite(ctx, e)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll("dist", 0o755); err != nil {
		return "", err
	}
	zipPath := filepath.Join("dist", course.Slug+"-bundle.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	zw := zip.NewWriter(f)

	var files, bytesTotal int64
	err = filepath.WalkDir(publicDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(publicDir, path)
		if err != nil {
			return err
		}
		w, err := zw.Create(filepath.ToSlash(rel))
		if err != nil {
			return err
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		n, err := io.Copy(w, in)
		files++
		bytesTotal += n
		return err
	})
	if err != nil {
		zw.Close()
		return "", fmt.Errorf("zipping site: %w", err)
	}
	if err := zw.Close(); err != nil {
		return "", err
	}
	fmt.Fprintf(e.out(), "  → bundle    %d file(s), %.1f MB → %s\n", files, float64(bytesTotal)/(1<<20), zipPath)
	return zipPath, nil
}

package pipeline

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"

	"github.com/enfec/coursesmith/internal/config"
)

// Title cards are rendered as PNGs in pure Go rather than with ffmpeg's
// drawtext filter: many ffmpeg builds (including Homebrew's) ship without
// libfreetype, so drawtext is not portable.

const (
	slideWidth  = 1920
	slideHeight = 1080
)

// fontCandidates are common sans-serif font locations per platform.
var fontCandidates = []string{
	"/System/Library/Fonts/Helvetica.ttc",                             // macOS
	"/System/Library/Fonts/Supplemental/Arial.ttf",                    // macOS
	"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",                 // Debian/Ubuntu
	"/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf", // Debian/Ubuntu
	"/usr/share/fonts/TTF/DejaVuSans.ttf",                             // Arch
	`C:\Windows\Fonts\arial.ttf`,                                      // Windows
}

// findFont locates and parses a usable system font for title cards.
func findFont() (*sfnt.Font, error) {
	for _, path := range fontCandidates {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if coll, err := sfnt.ParseCollection(data); err == nil && coll.NumFonts() > 0 {
			if f, err := coll.Font(0); err == nil {
				return f, nil
			}
		}
		if f, err := sfnt.Parse(data); err == nil {
			return f, nil
		}
	}
	return nil, fmt.Errorf(
		"no usable system font found for slides mode (looked in %s) — install DejaVu Sans or provide a %s instead",
		strings.Join(fontCandidates, ", "), RecordingFileName,
	)
}

// slideSpec describes one title card.
type slideSpec struct {
	Heading  string
	Subtitle string
	Colors   config.Colors
}

// renderTitleCard draws a 1920x1080 branded title card and writes it as PNG.
func renderTitleCard(fnt *sfnt.Font, spec slideSpec, outPath string) error {
	img := image.NewRGBA(image.Rect(0, 0, slideWidth, slideHeight))
	draw.Draw(img, img.Bounds(), image.NewUniform(parseHexColor(spec.Colors.Background, color.RGBA{255, 255, 255, 255})), image.Point{}, draw.Src)

	if err := drawCenteredText(img, fnt, spec.Heading, 88, slideHeight/2-40, parseHexColor(spec.Colors.Primary, color.RGBA{0, 0, 0, 255})); err != nil {
		return fmt.Errorf("drawing heading: %w", err)
	}
	if err := drawCenteredText(img, fnt, spec.Subtitle, 40, slideHeight/2+80, parseHexColor(spec.Colors.Accent, color.RGBA{85, 85, 85, 255})); err != nil {
		return fmt.Errorf("drawing subtitle: %w", err)
	}

	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("creating %s: %w", outPath, err)
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		return fmt.Errorf("encoding %s: %w", outPath, err)
	}
	return f.Close()
}

// drawCenteredText draws one line of text horizontally centered with its
// baseline at y.
func drawCenteredText(dst *image.RGBA, fnt *sfnt.Font, text string, size float64, y int, c color.Color) error {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	face, err := opentype.NewFace(fnt, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
	if err != nil {
		return err
	}
	defer face.Close()

	d := font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(c),
		Face: face,
	}
	width := d.MeasureString(text)
	d.Dot = fixed.Point26_6{
		X: (fixed.I(slideWidth) - width) / 2,
		Y: fixed.I(y),
	}
	d.DrawString(text)
	return nil
}

// parseHexColor parses "#rrggbb" or "#rgb"; anything else yields fallback.
func parseHexColor(s string, fallback color.RGBA) color.RGBA {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "#") {
		return fallback
	}
	hex := s[1:]
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}
	if len(hex) != 6 {
		return fallback
	}
	var r, g, b uint8
	if _, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b); err != nil {
		return fallback
	}
	return color.RGBA{r, g, b, 255}
}

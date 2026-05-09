package mchart

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers/rasterizer"
	xdraw "golang.org/x/image/draw"
)

//go:embed fonts/OpenSans.ttf
var openSansTTF []byte

var (
	canvasInitOnce sync.Once
	canvasInitErr  error

	canvasFontFamily = "Open Sans"
	canvasFontNames  = []string{"Open Sans", "OpenSans", "open sans", "opensans"}

	reTSpanBoundary = regexp.MustCompile(`(?is)</tspan>\s*<tspan\b[^>]*>`)
	reTSpanOpen     = regexp.MustCompile(`(?is)<tspan\b[^>]*>`)
	reTSpanClose    = regexp.MustCompile(`(?is)</tspan>`)
	reFontDecl      = regexp.MustCompile(`(?i)font-family\s*:\s*[^;}{]+`)
	reFontAttrDQ    = regexp.MustCompile(`(?i)font-family\s*=\s*"[^"]*"`)
	reFontAttrSQ    = regexp.MustCompile(`(?i)font-family\s*=\s*'[^']*'`)
)

func initCanvasFonts() error {
	canvasInitOnce.Do(func() {
		tmpDir, err := os.MkdirTemp("", "mchart-canvas-fonts-*")
		if err != nil {
			canvasInitErr = err
			return
		}

		regularPath := filepath.Join(tmpDir, "OpenSans.ttf")
		cachePath := filepath.Join(tmpDir, "system-fonts.cache")

		if err := os.WriteFile(regularPath, openSansTTF, 0o644); err != nil {
			canvasInitErr = err
			return
		}

		if err := canvas.CacheSystemFonts(cachePath, []string{tmpDir}); err != nil {
			canvasInitErr = err
			return
		}

		found := false
		for _, name := range canvasFontNames {
			if _, ok := canvas.FindSystemFont(name, canvas.FontRegular); ok {
				canvasFontFamily = name
				found = true
				break
			}
		}
		if !found {
			canvasInitErr = fmt.Errorf("embedded font could not be loaded into canvas system font cache")
		}
	})

	return canvasInitErr
}

func prepareSVGForCanvas(svgText, fontFamily string) string {
	sanitized := reTSpanBoundary.ReplaceAllString(svgText, "\n")
	sanitized = reTSpanOpen.ReplaceAllString(sanitized, "")
	sanitized = reTSpanClose.ReplaceAllString(sanitized, "")

	fontDecl := "font-family:" + fontFamily
	fontAttr := `font-family="` + fontFamily + `"`
	sanitized = reFontDecl.ReplaceAllString(sanitized, fontDecl)
	sanitized = reFontAttrDQ.ReplaceAllString(sanitized, fontAttr)
	sanitized = reFontAttrSQ.ReplaceAllString(sanitized, fontAttr)

	return sanitized
}

func parseSVGWithCanvas(svgText string) (_ *canvas.Canvas, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("canvas parse panic: %v", r)
		}
	}()
	return canvas.ParseSVG(strings.NewReader(svgText))
}

func rasterizeCanvas(svgText string, opts PNGOptions) ([]byte, error) {
	if err := initCanvasFonts(); err != nil {
		return nil, fmt.Errorf("canvas font init failed: %w", err)
	}

	prepared := prepareSVGForCanvas(svgText, canvasFontFamily)

	c, err := parseSVGWithCanvas(prepared)
	if err != nil {
		return nil, err
	}
	if c == nil || c.W <= 0 || c.H <= 0 {
		return nil, fmt.Errorf("%w: could not parse canvas dimensions", ErrInvalidDimensions)
	}

	dpmm := float64(opts.Width) / c.W
	if dpmm <= 0 {
		return nil, fmt.Errorf("%w: invalid canvas resolution", ErrInvalidDimensions)
	}

	rendered := rasterizer.Draw(c, canvas.DPMM(dpmm), canvas.DefaultColorSpace)
	finalImg := image.Image(rendered)

	if rendered.Bounds().Dx() != opts.Width || rendered.Bounds().Dy() != opts.Height {
		target := image.NewRGBA(image.Rect(0, 0, opts.Width, opts.Height))
		xdraw.CatmullRom.Scale(target, target.Bounds(), rendered, rendered.Bounds(), xdraw.Over, nil)
		finalImg = target
	}

	if bg, ok := parseHexColor(opts.Background); ok {
		withBG := image.NewRGBA(image.Rect(0, 0, opts.Width, opts.Height))
		draw.Draw(withBG, withBG.Bounds(), image.NewUniform(bg), image.Point{}, draw.Src)
		draw.Draw(withBG, withBG.Bounds(), finalImg, image.Point{}, draw.Over)
		finalImg = withBG
	}

	var out bytes.Buffer
	if err := png.Encode(&out, finalImg); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

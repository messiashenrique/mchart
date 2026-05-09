package mchart

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
)

func TestRenderPNGFromSVGCanvasBackend(t *testing.T) {
	svg := `<svg viewBox="0 0 120 60" xmlns="http://www.w3.org/2000/svg"><style>text{font-family:"Inter","Arial",sans-serif;font-size:16px;}</style><rect width="120" height="60" fill="#ffffff"/><text x="10" y="30">Olá</text></svg>`
	pngData, err := RenderPNGFromSVG(svg, PNGOptions{Backend: PNGBackendCanvas})
	if err != nil {
		t.Fatalf("RenderPNGFromSVG canvas backend failed: %v", err)
	}
	assertPNGSize(t, pngData, 120, 60)
}

func TestRenderPNGFromSVGScale(t *testing.T) {
	svg := `<svg viewBox="0 0 100 50" xmlns="http://www.w3.org/2000/svg"><rect width="100" height="50" fill="#ffffff"/></svg>`
	pngData, err := RenderPNGFromSVG(svg, PNGOptions{Backend: PNGBackendCanvas, Scale: 2})
	if err != nil {
		t.Fatalf("RenderPNGFromSVG failed: %v", err)
	}
	assertPNGSize(t, pngData, 200, 100)
}

func TestRenderPNGFromSVGAutoPrefersRSVG(t *testing.T) {
	origRSVG := rasterizeRSVGFunc
	origCanvas := rasterizeCanvasFunc
	t.Cleanup(func() {
		rasterizeRSVGFunc = origRSVG
		rasterizeCanvasFunc = origCanvas
	})

	rsvgCalls := 0
	canvasCalls := 0

	rasterizeRSVGFunc = func(svg string, opts PNGOptions) ([]byte, error) {
		rsvgCalls++
		return solidPNG(t, opts.Width, opts.Height), nil
	}
	rasterizeCanvasFunc = func(svg string, opts PNGOptions) ([]byte, error) {
		canvasCalls++
		return nil, fmt.Errorf("canvas should not be called when rsvg succeeds")
	}

	svg := `<svg viewBox="0 0 60 40" xmlns="http://www.w3.org/2000/svg"></svg>`
	pngData, err := RenderPNGFromSVG(svg, PNGOptions{Backend: PNGBackendAuto})
	if err != nil {
		t.Fatalf("RenderPNGFromSVG with auto backend failed: %v", err)
	}
	assertPNGSize(t, pngData, 60, 40)
	if rsvgCalls != 1 {
		t.Fatalf("expected rsvg to be called once, got %d", rsvgCalls)
	}
	if canvasCalls != 0 {
		t.Fatalf("expected canvas to not be called when rsvg succeeds, got %d calls", canvasCalls)
	}
}

func TestRenderPNGFromSVGAutoFallsBackToCanvas(t *testing.T) {
	origRSVG := rasterizeRSVGFunc
	origCanvas := rasterizeCanvasFunc
	t.Cleanup(func() {
		rasterizeRSVGFunc = origRSVG
		rasterizeCanvasFunc = origCanvas
	})

	canvasCalls := 0

	rasterizeRSVGFunc = func(svg string, opts PNGOptions) ([]byte, error) {
		return nil, errors.New("rsvg failed")
	}
	rasterizeCanvasFunc = func(svg string, opts PNGOptions) ([]byte, error) {
		canvasCalls++
		return solidPNG(t, opts.Width, opts.Height), nil
	}

	svg := `<svg viewBox="0 0 60 40" xmlns="http://www.w3.org/2000/svg"></svg>`
	pngData, err := RenderPNGFromSVG(svg, PNGOptions{Backend: PNGBackendAuto})
	if err != nil {
		t.Fatalf("RenderPNGFromSVG with auto backend failed: %v", err)
	}
	assertPNGSize(t, pngData, 60, 40)
	if canvasCalls != 1 {
		t.Fatalf("expected canvas to be called once, got %d", canvasCalls)
	}
}

func TestRenderPNGFromSVGAutoDoubleFailure(t *testing.T) {
	origRSVG := rasterizeRSVGFunc
	origCanvas := rasterizeCanvasFunc
	t.Cleanup(func() {
		rasterizeRSVGFunc = origRSVG
		rasterizeCanvasFunc = origCanvas
	})

	errRSVG := errors.New("rsvg boom")
	errCanvas := errors.New("canvas boom")

	rasterizeRSVGFunc = func(svg string, opts PNGOptions) ([]byte, error) {
		return nil, errRSVG
	}
	rasterizeCanvasFunc = func(svg string, opts PNGOptions) ([]byte, error) {
		return nil, errCanvas
	}

	svg := `<svg viewBox="0 0 30 20" xmlns="http://www.w3.org/2000/svg"></svg>`
	_, err := RenderPNGFromSVG(svg, PNGOptions{Backend: PNGBackendAuto})
	if err == nil {
		t.Fatal("expected error")
	}

	var renderErr *PNGRenderError
	if !errors.As(err, &renderErr) {
		t.Fatalf("expected PNGRenderError, got %T", err)
	}
	if len(renderErr.Attempts) != 2 {
		t.Fatalf("expected 2 attempts, got %d", len(renderErr.Attempts))
	}
	if !strings.Contains(err.Error(), "rsvg") || !strings.Contains(err.Error(), "canvas") {
		t.Fatalf("error should mention all backends: %v", err)
	}
}

func assertPNGSize(t *testing.T, pngData []byte, wantW, wantH int) {
	t.Helper()
	img, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		t.Fatalf("invalid png: %v", err)
	}
	if img.Bounds().Dx() != wantW || img.Bounds().Dy() != wantH {
		t.Fatalf("unexpected size: got %dx%d want %dx%d", img.Bounds().Dx(), img.Bounds().Dy(), wantW, wantH)
	}
}

func solidPNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}

	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return b.Bytes()
}

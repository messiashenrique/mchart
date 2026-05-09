package mchart

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
)

func TestRenderPNGFromSVGBasicNative(t *testing.T) {
	chart := buildSpiderChartFixture()
	svg, err := chart.RenderSVG()
	if err != nil {
		t.Fatalf("RenderSVG failed: %v", err)
	}

	pngData, err := RenderPNGFromSVG(svg, PNGOptions{Backend: PNGBackendNative, Width: 480, Height: 320})
	if err != nil {
		t.Fatalf("RenderPNGFromSVG failed: %v", err)
	}
	assertPNGSize(t, pngData, 480, 320)
}

func TestRenderPNGFromSVGScale(t *testing.T) {
	svg := `<svg viewBox="0 0 100 50" xmlns="http://www.w3.org/2000/svg"><rect width="100" height="50" fill="#ffffff"/></svg>`
	pngData, err := RenderPNGFromSVG(svg, PNGOptions{Backend: PNGBackendNative, Scale: 2})
	if err != nil {
		t.Fatalf("RenderPNGFromSVG failed: %v", err)
	}
	assertPNGSize(t, pngData, 200, 100)
}

func TestRenderPNGFromSVGAutoFallback(t *testing.T) {
	origNative := rasterizeNativeFunc
	origRSVG := rasterizeRSVGFunc
	t.Cleanup(func() {
		rasterizeNativeFunc = origNative
		rasterizeRSVGFunc = origRSVG
	})

	rasterizeNativeFunc = func(svg string, opts PNGOptions) ([]byte, error) {
		return nil, errors.New("native failed")
	}
	rasterizeRSVGFunc = func(svg string, opts PNGOptions) ([]byte, error) {
		return solidPNG(t, opts.Width, opts.Height), nil
	}

	svg := `<svg viewBox="0 0 60 40" xmlns="http://www.w3.org/2000/svg"></svg>`
	pngData, err := RenderPNGFromSVG(svg, PNGOptions{Backend: PNGBackendAuto})
	if err != nil {
		t.Fatalf("RenderPNGFromSVG with auto fallback failed: %v", err)
	}
	assertPNGSize(t, pngData, 60, 40)
}

func TestRenderPNGFromSVGAutoDoubleFailure(t *testing.T) {
	origNative := rasterizeNativeFunc
	origRSVG := rasterizeRSVGFunc
	t.Cleanup(func() {
		rasterizeNativeFunc = origNative
		rasterizeRSVGFunc = origRSVG
	})

	errNative := errors.New("native boom")
	errRSVG := errors.New("rsvg boom")

	rasterizeNativeFunc = func(svg string, opts PNGOptions) ([]byte, error) {
		return nil, errNative
	}
	rasterizeRSVGFunc = func(svg string, opts PNGOptions) ([]byte, error) {
		return nil, errRSVG
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
	if !strings.Contains(err.Error(), "native") || !strings.Contains(err.Error(), "rsvg") {
		t.Fatalf("error should mention both backends: %v", err)
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

package mchart

import (
	"encoding/xml"
	"errors"
	"fmt"
	"image/color"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type PNGBackend string

const (
	PNGBackendAuto   PNGBackend = "auto"
	PNGBackendRSVG   PNGBackend = "rsvg"
	PNGBackendCanvas PNGBackend = "canvas"
)

type PNGOptions struct {
	Width      int
	Height     int
	Scale      float64
	Background string
	Backend    PNGBackend
}

type BackendAttemptError struct {
	Backend PNGBackend
	Err     error
}

type PNGRenderError struct {
	Attempts []BackendAttemptError
}

func (e *PNGRenderError) Error() string {
	if e == nil || len(e.Attempts) == 0 {
		return "png render failed"
	}

	parts := make([]string, 0, len(e.Attempts))
	for _, attempt := range e.Attempts {
		parts = append(parts, fmt.Sprintf("%s: %v", attempt.Backend, attempt.Err))
	}

	return "png render failed (" + strings.Join(parts, "; ") + ")"
}

// Unwrap returns the first backend error to enable errors.Is/errors.As checks.
func (e *PNGRenderError) Unwrap() error {
	if e == nil || len(e.Attempts) == 0 {
		return nil
	}
	return e.Attempts[0].Err
}

type commandRunner interface {
	Run() error
}

type execCommand func(name string, args ...string) commandRunner

var (
	rasterizeRSVGFunc               = rasterizeRSVG
	rasterizeCanvasFunc             = rasterizeCanvas
	execLookPath                    = exec.LookPath
	execCommandFactory  execCommand = func(name string, args ...string) commandRunner {
		return exec.Command(name, args...)
	}
)

type svgMeta struct {
	Width  float64
	Height float64
}

func RenderPNGFromSVG(svg string, opts PNGOptions) ([]byte, error) {
	normalizedSVG, err := normalizeSVG(svg)
	if err != nil {
		return nil, err
	}

	meta, err := readSVGMeta(normalizedSVG)
	if err != nil {
		return nil, err
	}

	resolved, err := resolvePNGOptions(meta, opts)
	if err != nil {
		return nil, err
	}

	backend := resolved.Backend
	if backend == "" {
		backend = PNGBackendAuto
	}

	switch backend {
	case PNGBackendAuto:
		attempts := make([]BackendAttemptError, 0, 2)
		pngData, rsvgErr := rasterizeRSVGFunc(normalizedSVG, resolved)
		if rsvgErr == nil {
			return pngData, nil
		}
		attempts = append(attempts, BackendAttemptError{Backend: PNGBackendRSVG, Err: rsvgErr})

		pngData, canvasErr := rasterizeCanvasFunc(normalizedSVG, resolved)
		if canvasErr == nil {
			return pngData, nil
		}
		attempts = append(attempts, BackendAttemptError{Backend: PNGBackendCanvas, Err: canvasErr})

		return nil, &PNGRenderError{Attempts: attempts}

	case PNGBackendRSVG:
		pngData, err := rasterizeRSVGFunc(normalizedSVG, resolved)
		if err != nil {
			return nil, &PNGRenderError{Attempts: []BackendAttemptError{{Backend: PNGBackendRSVG, Err: err}}}
		}
		return pngData, nil

	case PNGBackendCanvas:
		pngData, err := rasterizeCanvasFunc(normalizedSVG, resolved)
		if err != nil {
			return nil, &PNGRenderError{Attempts: []BackendAttemptError{{Backend: PNGBackendCanvas, Err: err}}}
		}
		return pngData, nil

	default:
		return nil, fmt.Errorf("unsupported png backend: %q", backend)
	}
}

func resolvePNGOptions(meta svgMeta, opts PNGOptions) (PNGOptions, error) {
	resolved := opts
	if resolved.Backend == "" {
		resolved.Backend = PNGBackendAuto
	}
	if resolved.Scale <= 0 {
		resolved.Scale = 1
	}

	width := resolved.Width
	height := resolved.Height

	baseW := meta.Width
	baseH := meta.Height

	if baseW <= 0 || baseH <= 0 {
		return PNGOptions{}, fmt.Errorf("%w: could not determine svg size", ErrInvalidDimensions)
	}

	switch {
	case width > 0 && height > 0:
		// Keep provided dimensions.
	case width > 0 && height == 0:
		height = int(float64(width) * (baseH / baseW))
	case width == 0 && height > 0:
		width = int(float64(height) * (baseW / baseH))
	default:
		width = int(baseW)
		height = int(baseH)
	}

	width = int(float64(width) * resolved.Scale)
	height = int(float64(height) * resolved.Scale)

	if width <= 0 || height <= 0 {
		return PNGOptions{}, fmt.Errorf("%w: width=%d height=%d", ErrInvalidDimensions, width, height)
	}

	resolved.Width = width
	resolved.Height = height
	return resolved, nil
}

func readSVGMeta(svg string) (svgMeta, error) {
	type attrMap map[string]string

	decoder := xml.NewDecoder(strings.NewReader(svg))
	for {
		tok, err := decoder.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return svgMeta{}, err
		}

		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if strings.ToLower(start.Name.Local) != "svg" {
			continue
		}

		attrs := make(attrMap, len(start.Attr))
		for _, a := range start.Attr {
			attrs[strings.ToLower(a.Name.Local)] = strings.TrimSpace(a.Value)
		}

		if vb, ok := attrs["viewbox"]; ok {
			parts := strings.Fields(vb)
			if len(parts) == 4 {
				w, wErr := strconv.ParseFloat(parts[2], 64)
				h, hErr := strconv.ParseFloat(parts[3], 64)
				if wErr == nil && hErr == nil && w > 0 && h > 0 {
					return svgMeta{Width: w, Height: h}, nil
				}
			}
		}

		w, wErr := parseLength(attrs["width"])
		h, hErr := parseLength(attrs["height"])
		if wErr == nil && hErr == nil && w > 0 && h > 0 {
			return svgMeta{Width: w, Height: h}, nil
		}
		break
	}

	return svgMeta{}, fmt.Errorf("%w: missing viewBox/width/height", ErrInvalidDimensions)
}

func parseLength(v string) (float64, error) {
	t := strings.TrimSpace(v)
	if t == "" {
		return 0, ErrInvalidDimensions
	}

	if strings.HasSuffix(t, "px") {
		t = strings.TrimSpace(strings.TrimSuffix(t, "px"))
	}

	for len(t) > 0 {
		last := t[len(t)-1]
		if (last >= '0' && last <= '9') || last == '.' {
			break
		}
		t = strings.TrimSpace(t[:len(t)-1])
	}

	if t == "" {
		return 0, ErrInvalidDimensions
	}
	return strconv.ParseFloat(t, 64)
}

func rasterizeRSVG(svg string, opts PNGOptions) ([]byte, error) {
	if _, err := execLookPath("rsvg-convert"); err != nil {
		return nil, fmt.Errorf("rsvg-convert not available: %w", err)
	}

	tmpSVG, err := os.CreateTemp("", "graphics-*.svg")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpSVG.Name())

	if _, err := tmpSVG.WriteString(svg); err != nil {
		tmpSVG.Close()
		return nil, err
	}
	if err := tmpSVG.Close(); err != nil {
		return nil, err
	}

	tmpPNG, err := os.CreateTemp("", "graphics-*.png")
	if err != nil {
		return nil, err
	}
	tmpPNGPath := tmpPNG.Name()
	tmpPNG.Close()
	defer os.Remove(tmpPNGPath)

	args := []string{
		"-w", strconv.Itoa(opts.Width),
		"-h", strconv.Itoa(opts.Height),
		tmpSVG.Name(),
		"-o", tmpPNGPath,
	}
	cmd := execCommandFactory("rsvg-convert", args...)
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return os.ReadFile(tmpPNGPath)
}

func parseHexColor(input string) (color.Color, bool) {
	value := strings.TrimSpace(input)
	if value == "" || strings.EqualFold(value, "transparent") {
		return color.RGBA{}, false
	}

	value = strings.TrimPrefix(value, "#")

	switch len(value) {
	case 6:
		r, errR := strconv.ParseUint(value[0:2], 16, 8)
		g, errG := strconv.ParseUint(value[2:4], 16, 8)
		b, errB := strconv.ParseUint(value[4:6], 16, 8)
		if errR != nil || errG != nil || errB != nil {
			return color.RGBA{}, false
		}
		return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 0xFF}, true
	case 8:
		r, errR := strconv.ParseUint(value[0:2], 16, 8)
		g, errG := strconv.ParseUint(value[2:4], 16, 8)
		b, errB := strconv.ParseUint(value[4:6], 16, 8)
		a, errA := strconv.ParseUint(value[6:8], 16, 8)
		if errR != nil || errG != nil || errB != nil || errA != nil {
			return color.RGBA{}, false
		}
		return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)}, true
	default:
		return color.RGBA{}, false
	}
}

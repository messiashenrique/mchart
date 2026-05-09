package mchart

import (
	"errors"
	"os"
	"strings"
)

var (
	ErrEmptySVG          = errors.New("svg output is empty")
	ErrInvalidDimensions = errors.New("invalid output dimensions")
)

type svgRenderable interface {
	RenderSVG() (string, error)
}

func normalizeSVG(svg string) (string, error) {
	trimmed := strings.TrimSpace(svg)
	if trimmed == "" {
		return "", ErrEmptySVG
	}
	return trimmed, nil
}

func writeTextFile(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}

func renderPNGViaSVG(renderable svgRenderable, opts PNGOptions) ([]byte, error) {
	svg, err := renderable.RenderSVG()
	if err != nil {
		return nil, err
	}
	return RenderPNGFromSVG(svg, opts)
}

func writePNGViaSVG(path string, renderable svgRenderable, opts PNGOptions) error {
	pngData, err := renderPNGViaSVG(renderable, opts)
	if err != nil {
		return err
	}
	return os.WriteFile(path, pngData, 0o644)
}

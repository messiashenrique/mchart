package mchart

import (
	"fmt"
	"html"
	"strings"
)

type FunnelTheme string

const (
	FunnelThemeLight FunnelTheme = "light"
	FunnelThemeDark  FunnelTheme = "dark"
)

type FunnelValueMode string

const (
	FunnelValueFloat   FunnelValueMode = "float"
	FunnelValueInteger FunnelValueMode = "integer"
)

// FunnelSection represents one stage of a funnel chart.
type FunnelSection struct {
	Label string
	Value float64
	Color string
}

// FunnelChart renders stacked trapezoid funnel sections (SVG/PNG).
type FunnelChart struct {
	Title             string
	Sections          []FunnelSection
	Width             int
	Height            int
	Padding           int
	Palette           []string
	Theme             FunnelTheme
	ValueMode         FunnelValueMode
	ShowLegendConnect bool
}

// NewFunnelChart creates a FunnelChart with sensible defaults.
func NewFunnelChart(title string, sections []FunnelSection) *FunnelChart {
	return &FunnelChart{
		Title:             title,
		Sections:          sections,
		Width:             820,
		Height:            540,
		Padding:           24,
		Theme:             FunnelThemeLight,
		ValueMode:         FunnelValueFloat,
		ShowLegendConnect: true,
	}
}

func (fc *FunnelChart) RenderSVG() (string, error) {
	if len(fc.Sections) == 0 {
		return "", ErrEmptySVG
	}

	canvasW := fc.resolveWidth()
	canvasH := fc.resolveHeight()
	padding := fc.resolvePadding()
	colors := fc.resolveStyleColors()

	titleY := float64(padding + 28)
	titleGap := 26.0
	chartTop := titleY + titleGap
	chartBottom := float64(canvasH - padding - 22)
	if chartBottom <= chartTop {
		return "", ErrInvalidDimensions
	}

	legendFontSize := 16.0
	maxLabelW := 0.0
	for i, section := range fc.Sections {
		w := estimateTextWidth(fc.sectionLabel(i, section), legendFontSize)
		if w > maxLabelW {
			maxLabelW = w
		}
	}

	legendW := maxLabelW + 36
	minLegendW := 150.0
	maxLegendW := float64(canvasW) * 0.42
	if legendW < minLegendW {
		legendW = minLegendW
	}
	if legendW > maxLegendW {
		legendW = maxLegendW
	}

	chartLeft := float64(padding) + legendW
	chartRight := float64(canvasW - padding - 20)
	chartW := chartRight - chartLeft
	if chartW < 150 {
		return "", ErrInvalidDimensions
	}
	centerX := chartLeft + chartW/2

	n := len(fc.Sections)
	segmentH := (chartBottom - chartTop) / float64(n)
	if segmentH < 12 {
		return "", ErrInvalidDimensions
	}

	topWidth := chartW * 0.95
	bottomWidth := chartW * 0.05
	if n == 1 {
		bottomWidth = topWidth
	}

	var b strings.Builder
	fmt.Fprintf(&b, `<svg viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg">`, canvasW, canvasH)
	fmt.Fprintf(&b, `<style>
		.title { fill: %s; font-size: 21px; font-weight: 600; }
		.section-value { fill: %s; font-size: 26px; font-weight: 600; }
		.legend-text { fill: %s; font-size: 16px; font-weight: 500; }
		.legend-link { fill: none; stroke-width: 1.5; opacity: %.2f; stroke-linecap: round; }
		text { font-family: "Inter", "Segoe UI", "Roboto", "Arial", sans-serif; }
	</style>`, colors.title, colors.sectionValue, colors.legendText, colors.legendLineOpacity)

	fmt.Fprintf(&b, `<text class="title" x="%d" y="%d">%s</text>`, padding, padding+28, html.EscapeString(fc.Title))

	for i, section := range fc.Sections {
		y0 := chartTop + float64(i)*segmentH
		y1 := y0 + segmentH

		t0 := float64(i) / float64(n)
		t1 := float64(i+1) / float64(n)
		w0 := lerp(topWidth, bottomWidth, t0)
		w1 := lerp(topWidth, bottomWidth, t1)

		x0L := centerX - (w0 / 2)
		x0R := centerX + (w0 / 2)
		x1L := centerX - (w1 / 2)
		x1R := centerX + (w1 / 2)

		sectionColor := fc.resolveSectionColor(i, section)
		fmt.Fprintf(&b, `<polygon points="%.2f,%.2f %.2f,%.2f %.2f,%.2f %.2f,%.2f" fill="%s" />`, x0L, y0, x0R, y0, x1R, y1, x1L, y1, sectionColor)

		valueY := (y0 + y1) / 2
		fmt.Fprintf(&b, `<text class="section-value" x="%.2f" y="%.2f" text-anchor="middle" dominant-baseline="central">%s</text>`, centerX, valueY, fc.formatValue(section.Value))

		label := fc.sectionLabel(i, section)
		labelX := float64(padding)
		fmt.Fprintf(&b, `<text class="legend-text" x="%.2f" y="%.2f" dominant-baseline="central">%s</text>`, labelX, valueY, html.EscapeString(label))

		if fc.ShowLegendConnect {
			labelW := estimateTextWidth(label, legendFontSize)
			lineStartX := labelX + labelW + 8
			maxLineStartX := chartLeft - 20
			if lineStartX > maxLineStartX {
				lineStartX = maxLineStartX
			}

			lineEndX := ((x0L + x1L) / 2) - 8
			if lineEndX < lineStartX+4 {
				lineEndX = lineStartX + 4
			}

			fmt.Fprintf(&b, `<line class="legend-link" x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f" stroke="%s" />`, lineStartX, valueY, lineEndX, valueY, sectionColor)
		}
	}

	fmt.Fprint(&b, `</svg>`)
	return normalizeSVG(b.String())
}

func (fc *FunnelChart) resolveWidth() int {
	if fc.Width > 0 {
		return fc.Width
	}
	return 820
}

func (fc *FunnelChart) resolveHeight() int {
	if fc.Height > 0 {
		return fc.Height
	}
	return 540
}

func (fc *FunnelChart) resolvePadding() int {
	if fc.Padding <= 0 {
		return 24
	}
	return fc.Padding
}

func (fc *FunnelChart) resolvedPalette() []string {
	if len(fc.Palette) == 0 {
		return defaultColumnPalette
	}
	return fc.Palette
}

func (fc *FunnelChart) resolveSectionColor(index int, section FunnelSection) string {
	if strings.TrimSpace(section.Color) != "" {
		return section.Color
	}
	palette := fc.resolvedPalette()
	if len(palette) == 0 {
		return "#C8C8C8"
	}
	return palette[index%len(palette)]
}

func (fc *FunnelChart) sectionLabel(index int, section FunnelSection) string {
	label := strings.TrimSpace(section.Label)
	if label == "" {
		return fmt.Sprintf("Etapa %d", index+1)
	}
	return label
}

func (fc *FunnelChart) resolvedTheme() FunnelTheme {
	if fc.Theme == "" {
		return FunnelThemeLight
	}
	switch fc.Theme {
	case FunnelThemeLight, FunnelThemeDark:
		return fc.Theme
	default:
		return FunnelThemeLight
	}
}

func (fc *FunnelChart) resolvedValueMode() FunnelValueMode {
	if fc.ValueMode == "" {
		return FunnelValueFloat
	}
	switch fc.ValueMode {
	case FunnelValueFloat, FunnelValueInteger:
		return fc.ValueMode
	default:
		return FunnelValueFloat
	}
}

func (fc *FunnelChart) formatValue(value float64) string {
	if fc.resolvedValueMode() == FunnelValueInteger {
		return fmt.Sprintf("%.0f", value)
	}
	return ptNumber(value)
}

type funnelStyleColors struct {
	title             string
	sectionValue      string
	legendText        string
	legendLineOpacity float64
}

func (fc *FunnelChart) resolveStyleColors() funnelStyleColors {
	if fc.resolvedTheme() == FunnelThemeDark {
		return funnelStyleColors{
			title:             "#f3f4f6",
			sectionValue:      "#f9fafb",
			legendText:        "#e5e7eb",
			legendLineOpacity: 0.62,
		}
	}
	return funnelStyleColors{
		title:             "#111111",
		sectionValue:      "#111111",
		legendText:        "#111111",
		legendLineOpacity: 0.55,
	}
}

func (fc *FunnelChart) RenderPNG(opts PNGOptions) ([]byte, error) {
	return renderPNGViaSVG(fc, opts)
}

func (fc *FunnelChart) WriteSVG(path string) error {
	svg, err := fc.RenderSVG()
	if err != nil {
		return err
	}
	return writeTextFile(path, svg)
}

func (fc *FunnelChart) WritePNG(path string, opts PNGOptions) error {
	return writePNGViaSVG(path, fc, opts)
}

// ToSVG is deprecated and kept for compatibility with previous versions.
func (fc *FunnelChart) ToSVG() string {
	svg, err := fc.RenderSVG()
	if err != nil {
		return ""
	}
	return svg
}

func lerp(start, end, t float64) float64 {
	return start + (end-start)*t
}

package mchart

import (
	"fmt"
	"html"
	"math"
	"strings"
)

type SplineTheme string

const (
	SplineThemeLight SplineTheme = "light"
	SplineThemeDark  SplineTheme = "dark"
)

// SplineSeries defines one line/area dataset on a spline chart.
type SplineSeries struct {
	Label     string
	Values    []float64
	Color     string
	FillColor string
}

// SplineChart renders smoothed line charts with area under the curve (SVG/PNG).
type SplineChart struct {
	Title      string
	Labels     []string
	Series     []SplineSeries
	Width      int
	Height     int
	Padding    int
	Max        *float64
	Palette    []string
	Theme      SplineTheme
	ShowPoints bool
}

type splinePoint struct {
	x float64
	y float64
}

// NewSplineChart creates a SplineChart with sensible defaults.
func NewSplineChart(title string, labels []string, series []SplineSeries) *SplineChart {
	return &SplineChart{
		Title:      title,
		Labels:     labels,
		Series:     series,
		Width:      820,
		Height:     520,
		Padding:    24,
		Theme:      SplineThemeLight,
		ShowPoints: false,
	}
}

func (sc *SplineChart) RenderSVG() (string, error) {
	if len(sc.Labels) == 0 || len(sc.Series) == 0 {
		return "", ErrEmptySVG
	}

	canvasW := sc.resolveWidth()
	canvasH := sc.resolveHeight()
	padding := sc.resolvePadding()
	colors := sc.resolveStyleColors()

	titleTop := float64(padding + 28)
	titleGap := 22.0
	chartLeft := float64(padding + 44)
	chartRight := float64(canvasW - padding - 20)
	chartTop := titleTop + titleGap
	chartBottom := float64(canvasH - padding - 92)
	if chartBottom <= chartTop {
		return "", ErrInvalidDimensions
	}
	chartW := chartRight - chartLeft
	chartH := chartBottom - chartTop

	maxValue := sc.resolveMaxValue()
	ticks := 6

	var b strings.Builder
	fmt.Fprintf(&b, `<svg viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg">`, canvasW, canvasH)
	fmt.Fprintf(&b, `<style>
		.title { fill: %s; font-size: 34px; font-weight: 600; }
		.grid { stroke: %s; stroke-width: 1; }
		.axis { stroke: %s; stroke-width: 1.2; }
		.axis-text { fill: %s; font-size: 12px; font-weight: 500; }
		.legend-text { fill: %s; font-size: 14px; font-weight: 500; }
		text { font-family: "Inter", "Segoe UI", "Roboto", "Arial", sans-serif; }
	</style>`, colors.titleText, colors.gridStroke, colors.axisStroke, colors.axisText, colors.legendText)

	fmt.Fprintf(&b, `<text class="title" x="%d" y="%d">%s</text>`, padding, padding+28, html.EscapeString(sc.Title))

	for i := 0; i < ticks; i++ {
		ratio := float64(i) / float64(ticks-1)
		y := chartBottom - ratio*chartH
		value := ratio * maxValue
		fmt.Fprintf(&b, `<line class="grid" x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f" />`, chartLeft, y, chartRight, y)
		fmt.Fprintf(&b, `<text class="axis-text" x="%.2f" y="%.2f" text-anchor="end" dominant-baseline="central">%s</text>`, chartLeft-12, y, sc.formatAxisValue(value))
	}

	fmt.Fprintf(&b, `<line class="axis" x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f" />`, chartLeft, chartBottom, chartRight, chartBottom)

	for i, label := range sc.Labels {
		x := chartLeft + sc.labelRatio(i)*chartW
		fmt.Fprintf(&b, `<text class="axis-text" x="%.2f" y="%.2f" text-anchor="middle">%s</text>`, x, chartBottom+24, html.EscapeString(label))
	}

	for si, series := range sc.Series {
		if len(series.Values) == 0 {
			continue
		}
		limit := len(series.Values)
		if limit > len(sc.Labels) {
			limit = len(sc.Labels)
		}
		if limit < 2 {
			continue
		}

		pts := make([]splinePoint, 0, limit)
		for i := 0; i < limit; i++ {
			value := clamp(series.Values[i], 0, maxValue)
			x := chartLeft + sc.labelRatio(i)*chartW
			y := chartBottom - (value/maxValue)*chartH
			pts = append(pts, splinePoint{x: x, y: y})
		}

		lineColor := sc.resolveSeriesColor(si, series)
		fillColor := lineColor
		if strings.TrimSpace(series.FillColor) != "" {
			fillColor = series.FillColor
		}

		linePath := buildSplinePath(pts)
		areaPath := buildAreaPath(pts, chartBottom)
		fmt.Fprintf(&b, `<path d="%s" fill="%s" fill-opacity="%.2f" stroke="none" />`, areaPath, fillColor, colors.areaOpacity)
		fmt.Fprintf(&b, `<path d="%s" fill="none" stroke="%s" stroke-width="4" stroke-linecap="round" stroke-linejoin="round" />`, linePath, lineColor)

		if sc.ShowPoints {
			for _, p := range pts {
				fmt.Fprintf(&b, `<circle cx="%.2f" cy="%.2f" r="3.5" fill="%s" />`, p.x, p.y, lineColor)
			}
		}
	}

	legendTop := chartBottom + 50
	legendDotR := 6.0
	legendTextGap := 8.0
	legendFontSize := 14.0
	legendRowGap := 18.0
	legendLineHeight := 24.0
	legendSidePad := 14.0
	availableLegendW := float64(canvasW) - 2*legendSidePad

	type legendEntry struct {
		label string
		color string
		width float64
	}
	entries := make([]legendEntry, 0, len(sc.Series))
	for i, series := range sc.Series {
		label := strings.TrimSpace(series.Label)
		if label == "" {
			label = fmt.Sprintf("series%d", i+1)
		}
		width := (legendDotR * 2) + legendTextGap + estimateTextWidth(label, legendFontSize)
		entries = append(entries, legendEntry{
			label: label,
			color: sc.resolveSeriesColor(i, series),
			width: width,
		})
	}

	type legendLayout struct {
		cols      int
		rows      int
		colWidths []float64
		colGap    float64
	}

	layout := legendLayout{cols: 1, rows: len(entries), colWidths: []float64{0}, colGap: 0}
	if len(entries) > 0 {
		totalOneRow := 0.0
		for i, entry := range entries {
			totalOneRow += entry.width
			if i > 0 {
				totalOneRow += legendRowGap
			}
		}
		if totalOneRow <= availableLegendW {
			layout.cols = len(entries)
			layout.rows = 1
			layout.colWidths = make([]float64, len(entries))
			for i, entry := range entries {
				layout.colWidths[i] = entry.width
			}
			layout.colGap = legendRowGap
		} else {
			maxCols := len(entries)
			if maxCols > 4 {
				maxCols = 4
			}
			found := false
			for cols := maxCols; cols >= 2; cols-- {
				rows := int(math.Ceil(float64(len(entries)) / float64(cols)))
				colWidths := make([]float64, cols)
				for idx, entry := range entries {
					col := idx % cols
					if entry.width > colWidths[col] {
						colWidths[col] = entry.width
					}
				}

				sum := 0.0
				for _, w := range colWidths {
					sum += w
				}
				minGap := 20.0
				if sum+minGap*float64(cols-1) > availableLegendW {
					continue
				}
				gap := minGap
				if cols > 1 {
					gap = (availableLegendW - sum) / float64(cols-1)
				}
				layout = legendLayout{cols: cols, rows: rows, colWidths: colWidths, colGap: gap}
				found = true
				break
			}
			if !found {
				maxW := 0.0
				for _, entry := range entries {
					if entry.width > maxW {
						maxW = entry.width
					}
				}
				layout = legendLayout{cols: 1, rows: len(entries), colWidths: []float64{maxW}, colGap: 0}
			}
		}
	}

	colX := make([]float64, layout.cols)
	if layout.cols > 0 {
		colX[0] = legendSidePad
		for c := 1; c < layout.cols; c++ {
			colX[c] = colX[c-1] + layout.colWidths[c-1] + layout.colGap
		}
	}

	for idx, entry := range entries {
		col := idx % layout.cols
		row := idx / layout.cols
		x := colX[col]
		y := legendTop + float64(row)*legendLineHeight
		fmt.Fprintf(&b, `<circle cx="%.2f" cy="%.2f" r="%.2f" fill="%s" />`, x+legendDotR, y, legendDotR, entry.color)
		fmt.Fprintf(&b, `<text class="legend-text" x="%.2f" y="%.2f" dominant-baseline="central">%s</text>`, x+(legendDotR*2)+legendTextGap, y, html.EscapeString(entry.label))
	}

	fmt.Fprint(&b, `</svg>`)
	return normalizeSVG(b.String())
}

func (sc *SplineChart) resolveWidth() int {
	if sc.Width > 0 {
		return sc.Width
	}
	return 820
}

func (sc *SplineChart) resolveHeight() int {
	if sc.Height > 0 {
		return sc.Height
	}
	return 520
}

func (sc *SplineChart) resolvePadding() int {
	if sc.Padding <= 0 {
		return 24
	}
	return sc.Padding
}

func (sc *SplineChart) resolveMaxValue() float64 {
	if sc.Max != nil && *sc.Max > 0 {
		return *sc.Max
	}
	maxV := 0.0
	for _, series := range sc.Series {
		for _, v := range series.Values {
			if v > maxV {
				maxV = v
			}
		}
	}
	if maxV <= 0 {
		return 1
	}
	return maxV
}

func (sc *SplineChart) resolvedPalette() []string {
	if len(sc.Palette) == 0 {
		return defaultColumnPalette
	}
	return sc.Palette
}

func (sc *SplineChart) resolveSeriesColor(index int, series SplineSeries) string {
	if strings.TrimSpace(series.Color) != "" {
		return series.Color
	}
	palette := sc.resolvedPalette()
	if len(palette) == 0 {
		return "#C8C8C8"
	}
	return palette[index%len(palette)]
}

func (sc *SplineChart) resolvedTheme() SplineTheme {
	if sc.Theme == "" {
		return SplineThemeLight
	}
	switch sc.Theme {
	case SplineThemeLight, SplineThemeDark:
		return sc.Theme
	default:
		return SplineThemeLight
	}
}

type splineStyleColors struct {
	titleText   string
	gridStroke  string
	axisStroke  string
	axisText    string
	legendText  string
	areaOpacity float64
}

func (sc *SplineChart) resolveStyleColors() splineStyleColors {
	if sc.resolvedTheme() == SplineThemeDark {
		return splineStyleColors{
			titleText:   "#f3f4f6",
			gridStroke:  "#4b5563",
			axisStroke:  "#9ca3af",
			axisText:    "#d1d5db",
			legendText:  "#e5e7eb",
			areaOpacity: 0.35,
		}
	}
	return splineStyleColors{
		titleText:   "#111111",
		gridStroke:  "#d1d5db",
		axisStroke:  "#94a3b8",
		axisText:    "#111111",
		legendText:  "#333333",
		areaOpacity: 0.24,
	}
}

func (sc *SplineChart) formatAxisValue(v float64) string {
	rounded := math.Round(v)
	if math.Abs(v-rounded) < 1e-9 {
		return fmt.Sprintf("%.0f", rounded)
	}
	return ptNumber(v)
}

func (sc *SplineChart) labelRatio(index int) float64 {
	if len(sc.Labels) <= 1 {
		return 0.5
	}
	return float64(index) / float64(len(sc.Labels)-1)
}

func buildSplinePath(points []splinePoint) string {
	if len(points) == 0 {
		return ""
	}
	if len(points) == 1 {
		return fmt.Sprintf("M %.2f %.2f", points[0].x, points[0].y)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "M %.2f %.2f", points[0].x, points[0].y)
	for i := 0; i < len(points)-1; i++ {
		p0 := points[maxInt(0, i-1)]
		p1 := points[i]
		p2 := points[i+1]
		p3 := points[minInt(len(points)-1, i+2)]

		cp1x := p1.x + (p2.x-p0.x)/6.0
		cp1y := p1.y + (p2.y-p0.y)/6.0
		cp2x := p2.x - (p3.x-p1.x)/6.0
		cp2y := p2.y - (p3.y-p1.y)/6.0

		fmt.Fprintf(&b, " C %.2f %.2f, %.2f %.2f, %.2f %.2f", cp1x, cp1y, cp2x, cp2y, p2.x, p2.y)
	}
	return b.String()
}

func buildAreaPath(points []splinePoint, baselineY float64) string {
	if len(points) == 0 {
		return ""
	}
	linePath := buildSplinePath(points)
	last := points[len(points)-1]
	first := points[0]
	return fmt.Sprintf("%s L %.2f %.2f L %.2f %.2f Z", linePath, last.x, baselineY, first.x, baselineY)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (sc *SplineChart) RenderPNG(opts PNGOptions) ([]byte, error) {
	return renderPNGViaSVG(sc, opts)
}

func (sc *SplineChart) WriteSVG(path string) error {
	svg, err := sc.RenderSVG()
	if err != nil {
		return err
	}
	return writeTextFile(path, svg)
}

func (sc *SplineChart) WritePNG(path string, opts PNGOptions) error {
	return writePNGViaSVG(path, sc, opts)
}

// ToSVG is deprecated and kept for compatibility with previous versions.
func (sc *SplineChart) ToSVG() string {
	svg, err := sc.RenderSVG()
	if err != nil {
		return ""
	}
	return svg
}

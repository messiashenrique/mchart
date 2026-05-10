package mchart

import (
	"fmt"
	"html"
	"math"
	"strings"
)

// SpiderDataset defines one series on the radar/spider chart.
type SpiderDataset struct {
	Values      []float64
	Label       string
	FillColor   string
	StrokeColor string
}

type SpiderTheme string

const (
	SpiderThemeLight SpiderTheme = "light"
	SpiderThemeDark  SpiderTheme = "dark"
)

// SpiderChart renders a radar/spider chart for static reports (SVG/PNG).
type SpiderChart struct {
	Title    string
	Labels   []string
	Datasets []SpiderDataset
	Max      *float64
	Levels   int
	Size     int
	Padding  int
	Theme    SpiderTheme
}

// NewSpiderChart creates a SpiderChart with sensible defaults.
func NewSpiderChart(labels []string, datasets []SpiderDataset) *SpiderChart {
	return &SpiderChart{
		Labels:   labels,
		Datasets: datasets,
		Levels:   5,
		Size:     600,
		Padding:  100,
		Theme:    SpiderThemeLight,
	}
}

func (sc *SpiderChart) getPoint(index int, ratio, cx, cy, r float64) (float64, float64) {
	n := float64(len(sc.Labels))
	if n == 0 {
		return cx, cy
	}
	angle := (2*math.Pi*float64(index))/n - math.Pi/2
	radius := ratio * r
	return cx + radius*math.Cos(angle), cy + radius*math.Sin(angle)
}

func (sc *SpiderChart) RenderSVG() (string, error) {
	n := float64(len(sc.Labels))
	if n == 0 {
		return "", ErrEmptySVG
	}

	r := float64(sc.Size/2 - sc.Padding)
	cx := float64(sc.Size / 2)
	cy := float64(sc.Size / 2)

	chartMax := 100.0
	if sc.Max != nil {
		chartMax = *sc.Max
	} else {
		maxVal := 1.0
		for _, ds := range sc.Datasets {
			for _, v := range ds.Values {
				if v > maxVal {
					maxVal = v
				}
			}
		}
		chartMax = maxVal
	}
	if chartMax == 0 {
		chartMax = 1
	}
	colors := sc.resolveStyleColors()

	defaultColors := []struct {
		fill        string
		stroke      string
		fillOpacity float64
	}{
		{fill: "#8B5CF6", stroke: "#8B5CF6", fillOpacity: 0.3},
		{fill: "#24bc86", stroke: "#24bc86", fillOpacity: 0.3},
		{fill: "#ef4444", stroke: "#ef4444", fillOpacity: 0.3},
		{fill: "#2646ff", stroke: "#2646ff", fillOpacity: 0.3},
		{fill: "#d3Ea08", stroke: "#d3Ea08", fillOpacity: 0.3},
		{fill: "#ff7d33", stroke: "#ff7d33", fillOpacity: 0.3},
	}

	type legendEntry struct {
		label string
		color string
		width float64
	}

	legendDotRadius := 5.0
	legendDotTextGap := 8.0
	legendFontSize := 14.0
	legendLineHeight := 22.0
	legendRowGap := 16.0
	legendColGapMin := 24.0
	legendTopY := float64(sc.Size) - 50.0
	legendSidePadding := 10.0

	entries := make([]legendEntry, 0, len(sc.Datasets))
	for i, ds := range sc.Datasets {
		stroke := ds.StrokeColor
		if stroke == "" {
			stroke = defaultColors[i%len(defaultColors)].stroke
		}
		label := strings.TrimSpace(ds.Label)
		if label == "" {
			label = fmt.Sprintf("Série %d", i+1)
		}

		width := (legendDotRadius * 2) + legendDotTextGap + estimateTextWidth(label, legendFontSize)
		entries = append(entries, legendEntry{
			label: label,
			color: stroke,
			width: width,
		})
	}

	availableLegendWidth := float64(sc.Size) - (2 * legendSidePadding)

	type legendLayout struct {
		cols      int
		rows      int
		colWidths []float64
		colGap    float64
	}

	layout := legendLayout{
		cols:      1,
		rows:      len(entries),
		colWidths: []float64{0},
		colGap:    0,
	}

	totalOneRowWidth := 0.0
	for i, entry := range entries {
		totalOneRowWidth += entry.width
		if i > 0 {
			totalOneRowWidth += legendRowGap
		}
	}

	if len(entries) > 0 && totalOneRowWidth <= availableLegendWidth {
		layout.cols = len(entries)
		layout.rows = 1
		layout.colWidths = make([]float64, len(entries))
		for i, entry := range entries {
			layout.colWidths[i] = entry.width
		}
		layout.colGap = legendRowGap
	} else if len(entries) > 0 {
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

			totalWidth := 0.0
			for _, w := range colWidths {
				totalWidth += w
			}
			minTotal := totalWidth + legendColGapMin*float64(cols-1)
			if minTotal > availableLegendWidth {
				continue
			}

			colGap := legendColGapMin
			if cols > 1 {
				colGap = (availableLegendWidth - totalWidth) / float64(cols-1)
			}

			layout = legendLayout{
				cols:      cols,
				rows:      rows,
				colWidths: colWidths,
				colGap:    colGap,
			}
			found = true
			break
		}

		if !found {
			maxWidth := 0.0
			for _, entry := range entries {
				if entry.width > maxWidth {
					maxWidth = entry.width
				}
			}
			layout = legendLayout{
				cols:      1,
				rows:      len(entries),
				colWidths: []float64{maxWidth},
				colGap:    0,
			}
		}
	}

	legendBottomY := legendTopY
	if len(entries) > 0 {
		legendBottomY = legendTopY + float64(layout.rows-1)*legendLineHeight
	}
	svgHeight := int(math.Ceil(legendBottomY + 28))
	minHeight := sc.Size + 10
	if svgHeight < minHeight {
		svgHeight = minHeight
	}

	var b strings.Builder

	fmt.Fprintf(&b, `<svg viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg" style="font-family: sans-serif;">`, sc.Size, svgHeight)
	fmt.Fprintf(&b, `<style>
		.title { fill: %s; font-size: 21px; font-weight: 600; }
		.grid { fill: none; stroke: %s; stroke-opacity: %g; stroke-width: 1; }
		.axis { stroke: %s; stroke-opacity: %g; stroke-width: 1; }
		.label { fill: %s; font-size: 14px; font-weight: bold; }
		.legend-text { fill: %s; font-size: 14px; }
	</style>`, colors.titleText, colors.gridStroke, colors.gridOpacity, colors.axisStroke, colors.axisOpacity, colors.labelFill, colors.legendText)

	if strings.TrimSpace(sc.Title) != "" {
		fmt.Fprintf(&b, `<text class="title" x="%.2f" y="34" text-anchor="middle">%s</text>`, float64(sc.Size)/2, html.EscapeString(sc.Title))
	}

	for l := 0; l < sc.Levels; l++ {
		ratio := float64(l+1) / float64(sc.Levels)
		points := make([]string, len(sc.Labels))
		for i := 0; i < len(sc.Labels); i++ {
			x, y := sc.getPoint(i, ratio, cx, cy, r)
			points[i] = fmt.Sprintf("%.2f,%.2f", x, y)
		}
		fmt.Fprintf(&b, `<polygon class="grid" points="%s" />`, strings.Join(points, " "))
	}

	for i := 0; i < len(sc.Labels); i++ {
		x, y := sc.getPoint(i, 1, cx, cy, r)
		fmt.Fprintf(&b, `<line class="axis" x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f" />`, cx, cy, x, y)
	}

	for di, ds := range sc.Datasets {
		fill := ds.FillColor
		stroke := ds.StrokeColor
		fillOpacity := defaultColors[di%len(defaultColors)].fillOpacity
		if fill == "" {
			fill = defaultColors[di%len(defaultColors)].fill
		}
		if stroke == "" {
			stroke = defaultColors[di%len(defaultColors)].stroke
		}

		type plotPoint struct {
			x float64
			y float64
		}
		points := make([]plotPoint, 0, len(ds.Values))
		limit := len(ds.Values)
		if limit > len(sc.Labels) {
			limit = len(sc.Labels)
		}

		for j := 0; j < limit; j++ {
			ratio := math.Min(ds.Values[j]/chartMax, 1.0)
			x, y := sc.getPoint(j, ratio, cx, cy, r)
			points = append(points, plotPoint{x: x, y: y})
		}

		pointPairs := make([]string, len(points))
		for pi, p := range points {
			pointPairs[pi] = fmt.Sprintf("%.2f,%.2f", p.x, p.y)
		}

		polygonPoints := strings.Join(pointPairs, " ")
		fmt.Fprintf(&b, `<polygon points="%s" fill="%s" fill-opacity="%.2f" stroke="%s" stroke-width="2" stroke-linejoin="round" />`, polygonPoints, fill, fillOpacity, stroke)

		for _, p := range points {
			fmt.Fprintf(&b, `<circle cx="%.2f" cy="%.2f" r="3" fill="%s" stroke="%s" stroke-width="2" />`, p.x, p.y, stroke, stroke)
		}
	}

	labelFontSize := 12.0
	labelLineHeight := 14.0
	labelEdgeMargin := 8.0
	minLabelWidth := 40.0
	preferredLabelWidth := float64(sc.Padding)

	for i, label := range sc.Labels {
		angle := (2*math.Pi*float64(i))/n - math.Pi/2
		x, y := sc.getPoint(i, 1.06, cx, cy, r)

		cos := math.Cos(angle)
		sin := math.Sin(angle)

		anchor := "middle"
		baseline := "middle"

		if cos > 0.1 {
			anchor = "start"
		} else if cos < -0.1 {
			anchor = "end"
		}

		if sin > 0.1 {
			baseline = "hanging"
		} else if sin < -0.1 {
			baseline = "baseline"
		}

		maxWidth := preferredLabelWidth
		switch anchor {
		case "start":
			maxWidth = float64(sc.Size) - x - labelEdgeMargin
		case "end":
			maxWidth = x - labelEdgeMargin
		default:
			left := x - labelEdgeMargin
			right := float64(sc.Size) - x - labelEdgeMargin
			maxWidth = 2 * math.Min(left, right)
		}

		if maxWidth > preferredLabelWidth {
			maxWidth = preferredLabelWidth
		}
		if maxWidth < minLabelWidth {
			maxWidth = minLabelWidth
		}

		lines := wrapLabel(label, maxWidth, labelFontSize)
		if len(lines) == 0 {
			lines = []string{label}
		}

		firstLineY := y
		totalHeight := labelLineHeight * float64(len(lines)-1)
		switch baseline {
		case "baseline":
			firstLineY = y - totalHeight
		case "middle":
			firstLineY = y - (totalHeight / 2)
		}

		fmt.Fprintf(&b, `<text x="%.2f" y="%.2f" text-anchor="%s" class="label">`, x, firstLineY, anchor)
		for li, line := range lines {
			dy := 0.0
			if li > 0 {
				dy = labelLineHeight
			}
			fmt.Fprintf(&b, `<tspan x="%.2f" dy="%.2f">%s</tspan>`, x, dy, html.EscapeString(line))
		}
		fmt.Fprint(&b, `</text>`)
	}

	colX := make([]float64, layout.cols)
	if layout.cols > 0 {
		colX[0] = legendSidePadding
		for c := 1; c < layout.cols; c++ {
			colX[c] = colX[c-1] + layout.colWidths[c-1] + layout.colGap
		}
	}

	for idx, entry := range entries {
		col := idx % layout.cols
		row := idx / layout.cols
		y := legendTopY + float64(row)*legendLineHeight
		x := colX[col]
		fmt.Fprintf(&b, `<circle cx="%.2f" cy="%.2f" r="%.2f" fill="%s" />`, x+legendDotRadius, y, legendDotRadius, entry.color)
		fmt.Fprintf(&b, `<text class="legend-text" x="%.2f" y="%.2f" dominant-baseline="central">%s</text>`, x+(legendDotRadius*2)+legendDotTextGap, y, html.EscapeString(entry.label))
	}

	fmt.Fprint(&b, `</svg>`)
	return normalizeSVG(b.String())
}

func (sc *SpiderChart) resolvedTheme() SpiderTheme {
	if sc.Theme == "" {
		return SpiderThemeLight
	}
	switch sc.Theme {
	case SpiderThemeLight, SpiderThemeDark:
		return sc.Theme
	default:
		return SpiderThemeLight
	}
}

type spiderStyleColors struct {
	titleText   string
	gridStroke  string
	gridOpacity float64
	axisStroke  string
	axisOpacity float64
	labelFill   string
	legendText  string
}

func (sc *SpiderChart) resolveStyleColors() spiderStyleColors {
	if sc.resolvedTheme() == SpiderThemeDark {
		return spiderStyleColors{
			titleText:   "#f3f4f6",
			gridStroke:  "#9ca3af",
			gridOpacity: 0.28,
			axisStroke:  "#9ca3af",
			axisOpacity: 0.28,
			labelFill:   "#e5e7eb",
			legendText:  "#e5e7eb",
		}
	}
	return spiderStyleColors{
		titleText:   "#111111",
		gridStroke:  "#000000",
		gridOpacity: 0.1,
		axisStroke:  "#000000",
		axisOpacity: 0.1,
		labelFill:   "#333",
		legendText:  "#333",
	}
}

func (sc *SpiderChart) RenderPNG(opts PNGOptions) ([]byte, error) {
	return renderPNGViaSVG(sc, opts)
}

func (sc *SpiderChart) WriteSVG(path string) error {
	svg, err := sc.RenderSVG()
	if err != nil {
		return err
	}
	return writeTextFile(path, svg)
}

func (sc *SpiderChart) WritePNG(path string, opts PNGOptions) error {
	return writePNGViaSVG(path, sc, opts)
}

// ToSVG is deprecated and kept for compatibility with previous versions.
func (sc *SpiderChart) ToSVG() string {
	svg, err := sc.RenderSVG()
	if err != nil {
		return ""
	}
	return svg
}

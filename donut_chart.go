package mchart

import (
	"fmt"
	"html"
	"math"
	"strings"
)

type DonutTheme string

const (
	DonutThemeLight DonutTheme = "light"
	DonutThemeDark  DonutTheme = "dark"
)

// DonutSlice represents one segment of a donut chart.
type DonutSlice struct {
	Label string
	Value float64
	Color string
}

// DonutChart renders donut/pie-like charts for static reports (SVG/PNG).
type DonutChart struct {
	Title   string
	Slices  []DonutSlice
	Width   int
	Height  int
	Padding int
	Palette []string
	Theme   DonutTheme
}

type donutLegendItem struct {
	lines []string
	color string
}

// NewDonutChart creates a DonutChart with sensible defaults.
func NewDonutChart(title string, slices []DonutSlice) *DonutChart {
	return &DonutChart{
		Title:   title,
		Slices:  slices,
		Width:   820,
		Height:  520,
		Padding: 24,
		Theme:   DonutThemeLight,
	}
}

func (dc *DonutChart) RenderSVG() (string, error) {
	if len(dc.Slices) == 0 {
		return "", ErrEmptySVG
	}

	canvasW := dc.resolveWidth()
	canvasH := dc.resolveHeight()
	baseCanvasH := canvasH
	padding := dc.resolvePadding()
	colors := dc.resolveStyleColors()

	total := dc.totalValue()
	if total <= 0 {
		return "", ErrEmptySVG
	}

	fontSizeLegend := 16.0
	maxLabelW := 0.0
	for i, slice := range dc.Slices {
		label := dc.sliceLabel(i, slice)
		w := estimateTextWidth(label, fontSizeLegend)
		if w > maxLabelW {
			maxLabelW = w
		}
	}

	legendBottom := maxLabelW > 210
	legendFontSize := 16.0
	legendLineH := 24.0
	legendItemGap := 6.0
	dotR := 8.0

	legendItems := make([]donutLegendItem, 0, len(dc.Slices))

	var cx, cy, outerR, legendTop float64
	innerRRatio := 0.50
	if legendBottom {
		cx = float64(canvasW) / 2
		outerR = math.Min(float64(canvasW)*0.24, float64(baseCanvasH)*0.26)
		titleY := float64(padding + 28)
		cy = titleY + 38 + outerR

		legendTextMaxW := float64(canvasW - 88)
		if legendTextMaxW < 120 {
			legendTextMaxW = 120
		}

		totalLegendH := 0.0
		for i, slice := range dc.Slices {
			label := dc.sliceLabel(i, slice)
			lines := wrapLabel(label, legendTextMaxW, legendFontSize)
			if len(lines) == 0 {
				lines = []string{label}
			}
			legendItems = append(legendItems, donutLegendItem{
				lines: lines,
				color: dc.resolveSliceColor(i, slice),
			})

			itemTextH := float64(len(lines)) * legendLineH
			itemH := math.Max(itemTextH, dotR*2)
			totalLegendH += itemH
			if i < len(dc.Slices)-1 {
				totalLegendH += legendItemGap
			}
		}

		legendTop = cy + outerR + 40
		canvasH = int(math.Ceil(legendTop + totalLegendH + 24))
	} else {
		cx = float64(canvasW) * 0.33
		cy = float64(baseCanvasH) * 0.52
		outerR = math.Min(float64(canvasW)*0.20, float64(baseCanvasH)*0.30)
	}
	innerR := outerR * innerRRatio

	var b strings.Builder
	fmt.Fprintf(&b, `<svg viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg">`, canvasW, canvasH)
	fmt.Fprintf(&b, `<style>
		.title { fill: %s; font-size: 21px; font-weight: 600; }
		.slice-value { fill: %s; font-size: 18px; font-weight: 600; }
		.legend-text { fill: %s; font-size: 16px; font-weight: 500; }
		.legend-dot { stroke: none; }
		.slice-label-line { fill: none; stroke-width: 1.5; stroke-linecap: round; }
		text { font-family: "Inter", "Segoe UI", "Roboto", "Arial", sans-serif; }
	</style>`, colors.title, colors.sliceValue, colors.legendText)

	fmt.Fprintf(&b, `<text class="title" x="%d" y="%d">%s</text>`, padding, padding+28, html.EscapeString(dc.Title))

	angle := -math.Pi / 2
	for i, slice := range dc.Slices {
		v := math.Max(slice.Value, 0)
		if v == 0 {
			continue
		}

		portion := v / total
		sweep := portion * 2 * math.Pi
		endAngle := angle + sweep
		color := dc.resolveSliceColor(i, slice)

		path := donutSegmentPath(cx, cy, outerR, innerR, angle, endAngle)
		fmt.Fprintf(&b, `<path d="%s" fill="%s" stroke="%s" stroke-width="3" />`, path, color, colors.sliceDivider)

		mid := angle + (sweep / 2)
		cosMid := math.Cos(mid)
		sinMid := math.Sin(mid)

		lineStartX := cx + outerR*cosMid
		lineStartY := cy + outerR*sinMid
		lineEndX := cx + (outerR+14)*cosMid
		lineEndY := cy + (outerR+14)*sinMid
		labelX := cx + (outerR+28)*cosMid
		labelY := cy + (outerR+28)*sinMid

		textAnchor := "middle"
		if cosMid > 0.12 {
			textAnchor = "start"
		} else if cosMid < -0.12 {
			textAnchor = "end"
		}

		fmt.Fprintf(&b, `<polyline class="slice-label-line" points="%.2f,%.2f %.2f,%.2f" stroke="%s" />`, lineStartX, lineStartY, lineEndX, lineEndY, color)
		fmt.Fprintf(&b, `<text class="slice-value" x="%.2f" y="%.2f" text-anchor="%s" dominant-baseline="middle">%s</text>`, labelX, labelY, textAnchor, ptPercent(portion*100))

		angle = endAngle
	}

	if legendBottom {
		fmt.Fprint(&b, `<g class="legend legend-bottom">`)
		dc.renderLegendBottom(&b, legendItems, legendTop, legendLineH, legendItemGap, dotR)
		fmt.Fprint(&b, `</g>`)
	} else {
		fmt.Fprint(&b, `<g class="legend legend-right">`)
		dc.renderLegendRight(&b, float64(canvasW), cy)
		fmt.Fprint(&b, `</g>`)
	}

	fmt.Fprint(&b, `</svg>`)
	return normalizeSVG(b.String())
}

func (dc *DonutChart) renderLegendRight(b *strings.Builder, canvasW, cy float64) {
	startX := canvasW * 0.72
	startY := cy - float64(len(dc.Slices)-1)*22
	lineH := 44.0
	dotR := 9.0

	for i, slice := range dc.Slices {
		y := startY + float64(i)*lineH
		x := startX
		color := dc.resolveSliceColor(i, slice)
		label := dc.sliceLabel(i, slice)

		fmt.Fprintf(b, `<circle class="legend-dot" cx="%.2f" cy="%.2f" r="%.2f" fill="%s" />`, x, y, dotR, color)
		fmt.Fprintf(b, `<text class="legend-text" x="%.2f" y="%.2f" dominant-baseline="central">%s</text>`, x+20, y, html.EscapeString(label))
	}
}

func (dc *DonutChart) renderLegendBottom(b *strings.Builder, items []donutLegendItem, startY, lineH, itemGap, dotR float64) {
	startX := 48.0
	y := startY
	for _, item := range items {
		itemTextH := float64(len(item.lines)) * lineH
		itemH := math.Max(itemTextH, dotR*2)
		dotY := y + itemH/2

		fmt.Fprintf(b, `<circle class="legend-dot" cx="%.2f" cy="%.2f" r="%.2f" fill="%s" />`, startX, dotY, dotR, item.color)
		fmt.Fprintf(b, `<text class="legend-text" x="%.2f" y="%.2f">`, startX+18, y+lineH*0.65)
		for li, line := range item.lines {
			if li == 0 {
				fmt.Fprintf(b, `<tspan x="%.2f">%s</tspan>`, startX+18, html.EscapeString(line))
				continue
			}
			fmt.Fprintf(b, `<tspan x="%.2f" dy="%.2f">%s</tspan>`, startX+18, lineH, html.EscapeString(line))
		}
		fmt.Fprint(b, `</text>`)

		y += itemH + itemGap
	}
}

func (dc *DonutChart) resolveWidth() int {
	if dc.Width > 0 {
		return dc.Width
	}
	return 820
}

func (dc *DonutChart) resolveHeight() int {
	if dc.Height > 0 {
		return dc.Height
	}
	return 520
}

func (dc *DonutChart) resolvePadding() int {
	if dc.Padding <= 0 {
		return 24
	}
	return dc.Padding
}

func (dc *DonutChart) resolvedPalette() []string {
	if len(dc.Palette) == 0 {
		return defaultColumnPalette
	}
	return dc.Palette
}

func (dc *DonutChart) resolveSliceColor(index int, slice DonutSlice) string {
	if strings.TrimSpace(slice.Color) != "" {
		return slice.Color
	}
	palette := dc.resolvedPalette()
	if len(palette) == 0 {
		return "#C8C8C8"
	}
	return palette[index%len(palette)]
}

func (dc *DonutChart) totalValue() float64 {
	total := 0.0
	for _, slice := range dc.Slices {
		if slice.Value > 0 {
			total += slice.Value
		}
	}
	return total
}

func (dc *DonutChart) sliceLabel(index int, slice DonutSlice) string {
	label := strings.TrimSpace(slice.Label)
	if label == "" {
		return fmt.Sprintf("series-%d", index+1)
	}
	return label
}

func (dc *DonutChart) resolvedTheme() DonutTheme {
	if dc.Theme == "" {
		return DonutThemeLight
	}
	switch dc.Theme {
	case DonutThemeLight, DonutThemeDark:
		return dc.Theme
	default:
		return DonutThemeLight
	}
}

type donutStyleColors struct {
	title        string
	sliceValue   string
	legendText   string
	sliceDivider string
}

func (dc *DonutChart) resolveStyleColors() donutStyleColors {
	if dc.resolvedTheme() == DonutThemeDark {
		return donutStyleColors{
			title:        "#f3f4f6",
			sliceValue:   "#f9fafb",
			legendText:   "#e5e7eb",
			sliceDivider: "#111827",
		}
	}
	return donutStyleColors{
		title:        "#111111",
		sliceValue:   "#111111",
		legendText:   "#333333",
		sliceDivider: "#ffffff",
	}
}

func donutSegmentPath(cx, cy, outerR, innerR, startAngle, endAngle float64) string {
	x1 := cx + outerR*math.Cos(startAngle)
	y1 := cy + outerR*math.Sin(startAngle)
	x2 := cx + outerR*math.Cos(endAngle)
	y2 := cy + outerR*math.Sin(endAngle)

	x3 := cx + innerR*math.Cos(endAngle)
	y3 := cy + innerR*math.Sin(endAngle)
	x4 := cx + innerR*math.Cos(startAngle)
	y4 := cy + innerR*math.Sin(startAngle)

	largeArc := 0
	if endAngle-startAngle > math.Pi {
		largeArc = 1
	}

	return fmt.Sprintf(
		"M %.2f %.2f A %.2f %.2f 0 %d 1 %.2f %.2f L %.2f %.2f A %.2f %.2f 0 %d 0 %.2f %.2f Z",
		x1, y1,
		outerR, outerR, largeArc, x2, y2,
		x3, y3,
		innerR, innerR, largeArc, x4, y4,
	)
}

func (dc *DonutChart) RenderPNG(opts PNGOptions) ([]byte, error) {
	return renderPNGViaSVG(dc, opts)
}

func (dc *DonutChart) WriteSVG(path string) error {
	svg, err := dc.RenderSVG()
	if err != nil {
		return err
	}
	return writeTextFile(path, svg)
}

func (dc *DonutChart) WritePNG(path string, opts PNGOptions) error {
	return writePNGViaSVG(path, dc, opts)
}

// ToSVG is deprecated and kept for compatibility with previous versions.
func (dc *DonutChart) ToSVG() string {
	svg, err := dc.RenderSVG()
	if err != nil {
		return ""
	}
	return svg
}

package mchart

import (
	"fmt"
	"html"
	"math"
	"strings"
)

type BarChartType = ColumnValueMode

const (
	BarTypePercent BarChartType = ValueModePercent
	BarTypeNumber  BarChartType = ValueModeNumber
)

type BarTheme string

const (
	BarThemeLight BarTheme = "light"
	BarThemeDark  BarTheme = "dark"
)

// BarItem represents one horizontal bar in a BarChart.
type BarItem struct {
	Label string
	Value float64
	Color string
}

// BarChart renders horizontal bars for static reports (SVG/PNG).
type BarChart struct {
	Title      string
	Bars       []BarItem
	Width      int
	MinWidth   int
	Height     int
	Padding    int
	RowGap     int
	ValueMode  BarChartType
	Tipo       BarChartType
	Palette    []string
	ShowZeroAx bool
	Theme      BarTheme
}

// NewBarChart creates a BarChart with sensible defaults.
func NewBarChart(title string, bars []BarItem) *BarChart {
	return &BarChart{
		Title:      title,
		Bars:       bars,
		Width:      0,
		MinWidth:   640,
		Height:     0,
		Padding:    24,
		RowGap:     18,
		ValueMode:  BarTypePercent,
		ShowZeroAx: true,
		Theme:      BarThemeLight,
	}
}

func (bc *BarChart) RenderSVG() (string, error) {
	if len(bc.Bars) == 0 {
		return "", ErrEmptySVG
	}

	canvasWidth := bc.resolveCanvasWidth()
	canvasHeight := bc.resolveCanvasHeight()
	padding := bc.resolvePadding()
	maxValue := bc.resolveMaxValue()
	colors := bc.resolveStyleColors()

	rowHeight := 58.0
	rowGap := float64(bc.resolveRowGap())
	rowsStartY := float64(padding + 50)
	trackH := 14.0
	trackYShift := 26.0
	trackX := float64(padding)
	trackW := float64(canvasWidth - (2 * padding))

	var b strings.Builder
	fmt.Fprintf(&b, `<svg viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg">`, canvasWidth, canvasHeight)
	fmt.Fprintf(&b, `<style>
		.title { fill: %s; font-size: 21px; font-weight: 600; }
		.row-label { fill: %s; font-size: 18px; font-weight: 500; }
		.row-value { fill: %s; font-size: 19px; font-weight: 600; }
		.track { fill: %s; }
		text { font-family: "Inter", "Segoe UI", "Roboto", "Arial", sans-serif; }
	</style>`, colors.title, colors.rowLabel, colors.rowValue, colors.track)

	fmt.Fprintf(&b, `<text class="title" x="%d" y="%d">%s</text>`, padding, padding+16, html.EscapeString(bc.Title))

	for i, bar := range bc.Bars {
		rowTop := rowsStartY + float64(i)*(rowHeight+rowGap)
		labelY := rowTop + 14
		trackY := rowTop + trackYShift

		label := strings.TrimSpace(bar.Label)
		if label == "" {
			label = fmt.Sprintf("Item %d", i+1)
		}

		valueText := bc.formatValue(bar.Value)

		fmt.Fprintf(&b, `<text class="row-label" x="%.2f" y="%.2f">%s</text>`, trackX, labelY, html.EscapeString(label))
		fmt.Fprintf(&b, `<text class="row-value" x="%.2f" y="%.2f" text-anchor="end">%s</text>`, trackX+trackW, labelY, html.EscapeString(valueText))

		fmt.Fprintf(&b, `<rect class="track" x="%.2f" y="%.2f" width="%.2f" height="%.2f" rx="6" />`, trackX, trackY, trackW, trackH)

		fillColor := bc.resolveBarColor(i, bar)
		ratio := bc.resolveBarRatio(bar.Value, maxValue)
		fillW := trackW * ratio
		if fillW > 0 {
			fmt.Fprintf(&b, `<rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" rx="6" fill="%s" />`, trackX, trackY, fillW, trackH, fillColor)
		} else if bc.ShowZeroAx && bar.Value <= 0 {
			// Keep a tiny rounded fill segment for zero values so the bar doesn't look missing.
			zeroStubW := 8.0
			if zeroStubW > trackW {
				zeroStubW = trackW
			}
			fmt.Fprintf(&b, `<rect class="zero-mark" x="%.2f" y="%.2f" width="%.2f" height="%.2f" rx="6" fill="%s" />`, trackX, trackY, zeroStubW, trackH, fillColor)
		}
	}

	fmt.Fprint(&b, `</svg>`)
	return normalizeSVG(b.String())
}

func (bc *BarChart) resolvedValueMode() BarChartType {
	mode := bc.Tipo
	if mode == "" {
		mode = bc.ValueMode
	}
	if mode == "" {
		return BarTypePercent
	}
	if mode != BarTypePercent && mode != BarTypeNumber {
		return BarTypePercent
	}
	return mode
}

func (bc *BarChart) resolvedTheme() BarTheme {
	if bc.Theme == "" {
		return BarThemeLight
	}
	switch bc.Theme {
	case BarThemeLight, BarThemeDark:
		return bc.Theme
	default:
		return BarThemeLight
	}
}

type barStyleColors struct {
	title    string
	rowLabel string
	rowValue string
	track    string
}

func (bc *BarChart) resolveStyleColors() barStyleColors {
	if bc.resolvedTheme() == BarThemeDark {
		return barStyleColors{
			title:    "#f3f4f6",
			rowLabel: "#d1d5db",
			rowValue: "#f9fafb",
			track:    "#374151",
		}
	}

	return barStyleColors{
		title:    "#111111",
		rowLabel: "#333333",
		rowValue: "#111111",
		track:    "#dfdfdf",
	}
}

func (bc *BarChart) resolveCanvasWidth() int {
	if bc.Width > 0 {
		return bc.Width
	}

	padding := bc.resolvePadding()
	maxLabelWidth := 0.0
	for _, bar := range bc.Bars {
		label := strings.TrimSpace(bar.Label)
		if label == "" {
			continue
		}
		w := estimateTextWidth(label, 20)
		if w > maxLabelWidth {
			maxLabelWidth = w
		}
	}

	width := int(math.Ceil(float64(2*padding) + math.Max(540, maxLabelWidth+220)))
	if bc.MinWidth > 0 && width < bc.MinWidth {
		width = bc.MinWidth
	}
	if width < 520 {
		width = 520
	}
	return width
}

func (bc *BarChart) resolveCanvasHeight() int {
	if bc.Height > 0 {
		return bc.Height
	}

	padding := bc.resolvePadding()
	rowGap := bc.resolveRowGap()
	if len(bc.Bars) == 0 {
		return (padding * 2) + 64
	}

	rowHeight := 58
	rowsStartY := padding + 50
	trackYShift := 26
	trackH := 14
	lastRowTop := rowsStartY + (len(bc.Bars)-1)*(rowHeight+rowGap)
	contentBottom := lastRowTop + trackYShift + trackH

	return contentBottom + padding
}

func (bc *BarChart) resolvePadding() int {
	if bc.Padding <= 0 {
		return 24
	}
	return bc.Padding
}

func (bc *BarChart) resolveRowGap() int {
	if bc.RowGap < 0 {
		return 0
	}
	if bc.RowGap == 0 {
		return 18
	}
	return bc.RowGap
}

func (bc *BarChart) resolvedPalette() []string {
	if len(bc.Palette) == 0 {
		return defaultColumnPalette
	}
	return bc.Palette
}

func (bc *BarChart) resolveBarColor(index int, bar BarItem) string {
	if strings.TrimSpace(bar.Color) != "" {
		return bar.Color
	}
	palette := bc.resolvedPalette()
	if len(palette) == 0 {
		return "#C8C8C8"
	}
	return palette[index%len(palette)]
}

func (bc *BarChart) resolveMaxValue() float64 {
	if bc.resolvedValueMode() == BarTypePercent {
		return 100
	}

	maxValue := 0.0
	for _, bar := range bc.Bars {
		if bar.Value > maxValue {
			maxValue = bar.Value
		}
	}
	if maxValue <= 0 {
		return 1
	}
	return maxValue
}

func (bc *BarChart) resolveBarRatio(value, maxValue float64) float64 {
	if bc.resolvedValueMode() == BarTypePercent {
		return clamp(value, 0, 100) / 100.0
	}
	if maxValue <= 0 {
		return 0
	}
	return clamp(value/maxValue, 0, 1)
}

func (bc *BarChart) formatValue(value float64) string {
	if bc.resolvedValueMode() == BarTypePercent {
		return ptPercent(value)
	}
	return ptNumber(value)
}

func (bc *BarChart) RenderPNG(opts PNGOptions) ([]byte, error) {
	return renderPNGViaSVG(bc, opts)
}

func (bc *BarChart) WriteSVG(path string) error {
	svg, err := bc.RenderSVG()
	if err != nil {
		return err
	}
	return writeTextFile(path, svg)
}

func (bc *BarChart) WritePNG(path string, opts PNGOptions) error {
	return writePNGViaSVG(path, bc, opts)
}

// ToSVG is deprecated and kept for compatibility with previous versions.
func (bc *BarChart) ToSVG() string {
	svg, err := bc.RenderSVG()
	if err != nil {
		return ""
	}
	return svg
}

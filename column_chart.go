package mchart

import (
	"fmt"
	"html"
	"math"
	"strings"
)

type ColumnValueMode string

const (
	ValueModePercent      ColumnValueMode = "percent"
	ValueModeNumber       ColumnValueMode = "number"
	ValueModePercentColor ColumnValueMode = "percent_color"
)

type ColumnLegendPolicy string

const (
	LegendAuto   ColumnLegendPolicy = "auto"
	LegendAlways ColumnLegendPolicy = "always"
	LegendNever  ColumnLegendPolicy = "never"
)

type ColumnTheme string

const (
	ColumnThemeLight ColumnTheme = "light"
	ColumnThemeDark  ColumnTheme = "dark"
)

// ColumnBar represents one vertical bar item inside a column card.
type ColumnBar struct {
	Label string
	Value float64
	Color string
}

// ColumnCard groups bars rendered in one panel of a column chart.
type ColumnCard struct {
	Title string
	Bars  []ColumnBar
}

// ColumnChart renders grouped vertical bars for static reports (SVG/PNG).
type ColumnChart struct {
	Title        string
	Cards        []ColumnCard
	Width        int
	MinWidth     int
	Height       int
	Padding      int
	CardGap      int
	ValueMode    ColumnValueMode
	LegendPolicy ColumnLegendPolicy
	Palette      []string
	Theme        ColumnTheme
}

// NewColumnChart creates a ColumnChart with sensible defaults.
func NewColumnChart(title string, cards []ColumnCard) *ColumnChart {
	return &ColumnChart{
		Title:        title,
		Cards:        cards,
		Width:        0,
		MinWidth:     520,
		Height:       0,
		Padding:      20,
		CardGap:      20,
		ValueMode:    ValueModePercent,
		LegendPolicy: LegendAuto,
		Theme:        ColumnThemeLight,
	}
}

func (cc *ColumnChart) RenderSVG() (string, error) {
	if len(cc.Cards) == 0 {
		return "", ErrEmptySVG
	}

	canvasWidth := cc.resolveCanvasWidth()
	canvasHeight := cc.resolveCanvasHeight(canvasWidth)
	colors := cc.resolveStyleColors()

	var b strings.Builder

	fmt.Fprintf(&b, `<svg viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg">`, canvasWidth, canvasHeight)
	fmt.Fprintf(&b, `<style>
			.title { fill: %s; font-size: 21px; font-weight: 600; }
			.card-title { fill: %s; font-size: 19px; font-weight: 600; }
			.grid { stroke: %s; stroke-width: 1; }
			.axis-text { fill: %s; font-size: 12px; font-weight: 400; }
			.bar-label { fill: %s; font-size: 13px; font-weight: 500; }
			.bar-value { fill: %s; font-size: 12px; font-weight: 700; }
			.legend-text { fill: %s; font-size: 16px; font-weight: 500; }
			text { font-family: "Inter", "Segoe UI", "Roboto", "Arial", sans-serif; }
		</style>`, colors.title, colors.cardTitle, colors.gridStroke, colors.axisText, colors.barLabel, colors.barValue, colors.legendText)

	fmt.Fprintf(&b, `<text class="title" x="%d" y="%d">%s</text>`, cc.Padding, cc.Padding+24, html.EscapeString(cc.Title))

	cardsCount := float64(len(cc.Cards))
	availableWidth := float64(canvasWidth - (2 * cc.Padding) - ((len(cc.Cards) - 1) * cc.CardGap))
	cardW := availableWidth / cardsCount
	cardY := float64(cc.Padding + 40)
	for i, card := range cc.Cards {
		cardX := float64(cc.Padding) + float64(i)*(cardW+float64(cc.CardGap))

		innerPad := 16.0
		headerY := cardY + 30
		fmt.Fprintf(&b, `<text class="card-title" x="%.2f" y="%.2f">%s</text>`, cardX+innerPad, headerY, html.EscapeString(card.Title))

		if len(card.Bars) == 0 {
			continue
		}

		plotTop := cardY + 70
		plotH := 180.0
		plotBottom := plotTop + plotH
		axisLabelW := 44.0
		plotLeft := cardX + innerPad + axisLabelW
		plotRight := cardX + cardW - innerPad
		barsAreaW := plotRight - plotLeft
		barsCount := float64(len(card.Bars))
		gap := 24.0
		barW := (barsAreaW - gap*(barsCount-1)) / barsCount
		maxBarW := 44.0
		minBarW := 20.0
		if barW > maxBarW {
			barW = maxBarW
		}
		if barW < minBarW {
			gap = 8
			barW = (barsAreaW - gap*(barsCount-1)) / barsCount
			if barW < minBarW {
				barW = minBarW
			}
		}

		groupW := barW*barsCount + gap*(barsCount-1)
		startX := plotLeft + (barsAreaW-groupW)/2
		minStartX := plotLeft
		if startX < minStartX {
			startX = minStartX
		}

		displayLabels := make([]string, len(card.Bars))
		showLegend := cc.shouldShowLegend(card.Bars, barW)

		for bi, bar := range card.Bars {
			label := strings.TrimSpace(bar.Label)
			if label == "" {
				label = fmt.Sprintf("Item %d", bi+1)
			}

			if showLegend {
				displayLabels[bi] = fmt.Sprintf("%d", bi+1)
			} else {
				displayLabels[bi] = label
			}
		}

		maxValue := cc.resolveMaxValue(card.Bars)
		for tick := 0; tick <= 4; tick++ {
			ratio := float64(tick) / 4.0
			y := plotBottom - ratio*plotH
			fmt.Fprintf(&b, `<line class="grid" x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f" />`, plotLeft, y, plotRight, y)
			fmt.Fprintf(&b, `<text class="axis-text" x="%.2f" y="%.2f" text-anchor="end" dominant-baseline="central">%s</text>`, plotLeft-8, y, cc.formatAxisValue(ratio*maxValue))
		}

		for bi, bar := range card.Bars {
			x := startX + float64(bi)*(barW+gap)
			value := bar.Value

			fillColor := cc.resolveBarColor(bi, bar)
			ratio := cc.resolveBarRatio(value, maxValue)
			fillH := plotH * ratio
			fillY := plotBottom - fillH
			fmt.Fprintf(&b, `<rect class="bar" x="%.2f" y="%.2f" width="%.2f" height="%.2f" rx="4" fill="%s" />`, x, fillY, barW, fillH, fillColor)

			labelX := x + (barW / 2)
			labelY := plotBottom + 24
			fmt.Fprintf(&b, `<text class="bar-label" x="%.2f" y="%.2f" text-anchor="middle">%s</text>`, labelX, labelY, html.EscapeString(displayLabels[bi]))
			valueY := fillY - 8
			if fillH == 0 {
				valueY = plotBottom - 8
			}
			fmt.Fprintf(&b, `<text class="bar-value" x="%.2f" y="%.2f" text-anchor="middle">%s</text>`, labelX, valueY, cc.formatValue(value))
		}

		if showLegend {
			legendY := plotBottom + 70
			legendX := cardX + innerPad
			legendMaxWidth := cardW - (innerPad * 2)
			lineHeight := 24.0
			swatchW := 10.0
			gapAfterSwatch := 6.0
			minColumnGap := 22.0
			rowGap := 14.0
			legendFontSize := 18.0

			type legendEntry struct {
				text  string
				color string
				width float64
			}

			entries := make([]legendEntry, 0, len(card.Bars))
			for bi, bar := range card.Bars {
				label := strings.TrimSpace(bar.Label)
				if label == "" {
					label = fmt.Sprintf("Item %d", bi+1)
				}
				itemText := fmt.Sprintf("%d: %s", bi+1, label)

				itemWidth := estimateTextWidth(itemText, legendFontSize) * 1.1
				if cc.shouldShowLegendSwatch() {
					itemWidth += swatchW + gapAfterSwatch
				}

				entries = append(entries, legendEntry{
					text:  itemText,
					color: cc.resolveBarColor(bi, bar),
					width: itemWidth,
				})
			}

			renderEntry := func(entry legendEntry, x, y float64) {
				textX := x
				if cc.shouldShowLegendSwatch() {
					fmt.Fprintf(&b, `<rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" rx="2" fill="%s" />`, x, y-10, swatchW, swatchW, entry.color)
					textX = x + swatchW + gapAfterSwatch
				}
				fmt.Fprintf(&b, `<text class="legend-text" x="%.2f" y="%.2f">%s</text>`, textX, y, html.EscapeString(entry.text))
			}

			totalOneRow := 0.0
			for i, entry := range entries {
				totalOneRow += entry.width
				if i > 0 {
					totalOneRow += rowGap
				}
			}

			if totalOneRow <= legendMaxWidth {
				for i, entry := range entries {
					x := legendX
					for j := 0; j < i; j++ {
						x += entries[j].width + rowGap
					}
					renderEntry(entry, x, legendY)
				}
				continue
			}

			maxCols := len(entries)
			if maxCols > 4 {
				maxCols = 4
			}

			type layout struct {
				cols      int
				colWidths []float64
				gap       float64
				ok        bool
			}

			best := layout{}
			for cols := maxCols; cols >= 2; cols-- {
				colWidths := make([]float64, cols)

				for idx, entry := range entries {
					col := idx % cols
					if entry.width > colWidths[col] {
						colWidths[col] = entry.width
					}
				}

				sumWidths := 0.0
				for _, w := range colWidths {
					sumWidths += w
				}

				minTotal := sumWidths + minColumnGap*float64(cols-1)
				if minTotal > legendMaxWidth {
					continue
				}

				gap := minColumnGap
				if cols > 1 {
					gap = (legendMaxWidth - sumWidths) / float64(cols-1)
				}

				best = layout{
					cols:      cols,
					colWidths: colWidths,
					gap:       gap,
					ok:        true,
				}
				break
			}

			if best.ok {
				colX := make([]float64, best.cols)
				colX[0] = legendX
				for c := 1; c < best.cols; c++ {
					colX[c] = colX[c-1] + best.colWidths[c-1] + best.gap
				}

				for idx, entry := range entries {
					col := idx % best.cols
					row := idx / best.cols
					y := legendY + float64(row)*lineHeight
					renderEntry(entry, colX[col], y)
				}
			} else {
				for i, entry := range entries {
					y := legendY + float64(i)*lineHeight
					renderEntry(entry, legendX, y)
				}
			}
		}
	}

	fmt.Fprint(&b, `</svg>`)
	return normalizeSVG(b.String())
}

func (cc *ColumnChart) shouldShowLegend(bars []ColumnBar, barW float64) bool {
	switch cc.resolvedLegendPolicy() {
	case LegendAlways:
		return true
	case LegendNever:
		return false
	}

	maxTextWidth := barW - 2
	for _, bar := range bars {
		label := strings.TrimSpace(bar.Label)
		if label == "" {
			continue
		}
		if estimateTextWidth(label, 19) > maxTextWidth {
			return true
		}
	}
	return false
}

func (cc *ColumnChart) resolvedValueMode() ColumnValueMode {
	if cc.ValueMode == "" {
		return ValueModePercent
	}
	switch cc.ValueMode {
	case ValueModePercent, ValueModeNumber, ValueModePercentColor:
		return cc.ValueMode
	default:
		return ValueModePercent
	}
}

func (cc *ColumnChart) resolveCanvasWidth() int {
	if cc.Width > 0 {
		return cc.Width
	}

	padding := cc.Padding
	if padding <= 0 {
		padding = 20
	}
	cardGap := cc.CardGap
	if cardGap < 0 {
		cardGap = 0
	}

	maxBars := 2
	for _, card := range cc.Cards {
		if len(card.Bars) > maxBars {
			maxBars = len(card.Bars)
		}
	}

	preferredBarW := 44.0
	preferredGap := 24.0
	innerPad := 16.0
	minCardW := 320.0
	axisLabelW := 44.0
	cardW := innerPad*2 + axisLabelW + float64(maxBars)*preferredBarW + float64(maxBars-1)*preferredGap
	if cardW < minCardW {
		cardW = minCardW
	}

	width := int(math.Ceil(float64(2*padding) + float64(len(cc.Cards))*cardW + float64(len(cc.Cards)-1)*float64(cardGap)))
	if cc.MinWidth > 0 && width < cc.MinWidth {
		width = cc.MinWidth
	}
	if width < 420 {
		width = 420
	}
	return width
}

func (cc *ColumnChart) resolveCanvasHeight(canvasWidth int) int {
	if cc.Height > 0 {
		return cc.Height
	}

	if len(cc.Cards) == 0 {
		return 420
	}

	padding := cc.Padding
	if padding <= 0 {
		padding = 20
	}
	cardGap := cc.CardGap
	if cardGap < 0 {
		cardGap = 0
	}

	availableWidth := float64(canvasWidth - (2 * padding) - ((len(cc.Cards) - 1) * cardGap))
	if availableWidth <= 0 {
		return 420
	}
	cardW := availableWidth / float64(len(cc.Cards))

	maxCardHeight := 0.0
	for _, card := range cc.Cards {
		cardHeight := cc.resolveCardHeight(card, cardW)
		if cardHeight > maxCardHeight {
			maxCardHeight = cardHeight
		}
	}
	if maxCardHeight <= 0 {
		maxCardHeight = 340
	}

	height := int(math.Ceil(float64(2*padding+40) + maxCardHeight))
	if height < 420 {
		height = 420
	}
	return height
}

func (cc *ColumnChart) resolveCardHeight(card ColumnCard, cardW float64) float64 {
	headerOnlyHeight := 88.0
	if len(card.Bars) == 0 {
		return headerOnlyHeight
	}

	innerPad := 16.0
	plotTop := 70.0
	plotH := 180.0
	labelBottom := plotTop + plotH + 24.0

	axisLabelW := 44.0
	barsAreaW := cardW - (innerPad * 2) - axisLabelW
	barsCount := float64(len(card.Bars))
	gap := 24.0
	barW := (barsAreaW - gap*(barsCount-1)) / barsCount
	maxBarW := 44.0
	minBarW := 20.0
	if barW > maxBarW {
		barW = maxBarW
	}
	if barW < minBarW {
		gap = 8
		barW = (barsAreaW - gap*(barsCount-1)) / barsCount
		if barW < minBarW {
			barW = minBarW
		}
	}

	if !cc.shouldShowLegend(card.Bars, barW) {
		return labelBottom + 24
	}

	legendY := plotTop + plotH + 70.0
	legendMaxWidth := cardW - (innerPad * 2)
	legendRows := cc.resolveLegendRows(card.Bars, legendMaxWidth)
	if legendRows < 1 {
		legendRows = 1
	}

	lineHeight := 24.0
	legendBottom := legendY + float64(legendRows-1)*lineHeight
	return legendBottom + 20
}

func (cc *ColumnChart) resolveLegendRows(bars []ColumnBar, legendMaxWidth float64) int {
	if len(bars) == 0 {
		return 0
	}
	if legendMaxWidth <= 0 {
		return len(bars)
	}

	swatchW := 10.0
	gapAfterSwatch := 6.0
	minColumnGap := 22.0
	rowGap := 14.0
	legendFontSize := 18.0
	type legendEntry struct {
		width float64
	}

	entries := make([]legendEntry, 0, len(bars))
	for bi, bar := range bars {
		label := strings.TrimSpace(bar.Label)
		if label == "" {
			label = fmt.Sprintf("Item %d", bi+1)
		}
		itemText := fmt.Sprintf("%d: %s", bi+1, label)

		itemWidth := estimateTextWidth(itemText, legendFontSize) * 1.1
		if cc.shouldShowLegendSwatch() {
			itemWidth += swatchW + gapAfterSwatch
		}
		entries = append(entries, legendEntry{width: itemWidth})
	}

	totalOneRow := 0.0
	for i, entry := range entries {
		totalOneRow += entry.width
		if i > 0 {
			totalOneRow += rowGap
		}
	}
	if totalOneRow <= legendMaxWidth {
		return 1
	}

	maxCols := len(entries)
	if maxCols > 4 {
		maxCols = 4
	}

	for cols := maxCols; cols >= 2; cols-- {
		colWidths := make([]float64, cols)
		for idx, entry := range entries {
			col := idx % cols
			if entry.width > colWidths[col] {
				colWidths[col] = entry.width
			}
		}

		sumWidths := 0.0
		for _, w := range colWidths {
			sumWidths += w
		}

		minTotal := sumWidths + minColumnGap*float64(cols-1)
		if minTotal > legendMaxWidth {
			continue
		}

		return int(math.Ceil(float64(len(entries)) / float64(cols)))
	}

	return len(entries)
}

func (cc *ColumnChart) resolvedLegendPolicy() ColumnLegendPolicy {
	if cc.LegendPolicy == "" {
		return LegendAuto
	}
	switch cc.LegendPolicy {
	case LegendAuto, LegendAlways, LegendNever:
		return cc.LegendPolicy
	default:
		return LegendAuto
	}
}

func (cc *ColumnChart) resolvedPalette() []string {
	if len(cc.Palette) == 0 {
		return defaultColumnPalette
	}
	return cc.Palette
}

func (cc *ColumnChart) resolvedTheme() ColumnTheme {
	if cc.Theme == "" {
		return ColumnThemeLight
	}
	switch cc.Theme {
	case ColumnThemeLight, ColumnThemeDark:
		return cc.Theme
	default:
		return ColumnThemeLight
	}
}

type columnStyleColors struct {
	title      string
	cardTitle  string
	gridStroke string
	axisText   string
	barLabel   string
	barValue   string
	legendText string
}

func (cc *ColumnChart) resolveStyleColors() columnStyleColors {
	if cc.resolvedTheme() == ColumnThemeDark {
		return columnStyleColors{
			title:      "#f3f4f6",
			cardTitle:  "#f9fafb",
			gridStroke: "#4b5563",
			axisText:   "#d1d5db",
			barLabel:   "#f3f4f6",
			barValue:   "#e5e7eb",
			legendText: "#d1d5db",
		}
	}
	return columnStyleColors{
		title:      "#111111",
		cardTitle:  "#111111",
		gridStroke: "#e5e7eb",
		axisText:   "#6b7280",
		barLabel:   "#111111",
		barValue:   "#333333",
		legendText: "#111111",
	}
}

func (cc *ColumnChart) formatAxisValue(value float64) string {
	if cc.isPercentMode() {
		return fmt.Sprintf("%.0f%%", value)
	}
	return ptNumber(value)
}

func (cc *ColumnChart) resolveBarColor(index int, bar ColumnBar) string {
	if strings.TrimSpace(bar.Color) != "" {
		return bar.Color
	}

	if cc.usesScoreColorScale() {
		return scoreColor(bar.Value)
	}

	palette := cc.resolvedPalette()
	if len(palette) == 0 {
		return "#C8C8C8"
	}
	return palette[index%len(palette)]
}

func (cc *ColumnChart) resolveMaxValue(bars []ColumnBar) float64 {
	if cc.isPercentMode() {
		return 100
	}

	maxValue := 0.0
	for _, bar := range bars {
		if bar.Value > maxValue {
			maxValue = bar.Value
		}
	}
	if maxValue <= 0 {
		return 1
	}
	return maxValue
}

func (cc *ColumnChart) resolveBarRatio(value, maxValue float64) float64 {
	if cc.isPercentMode() {
		return clamp(value, 0, 100) / 100.0
	}
	if maxValue <= 0 {
		return 0
	}
	return clamp(value/maxValue, 0, 1)
}

func (cc *ColumnChart) formatValue(value float64) string {
	if cc.isPercentMode() {
		return ptPercent(value)
	}
	return ptNumber(value)
}

func (cc *ColumnChart) isPercentMode() bool {
	switch cc.resolvedValueMode() {
	case ValueModePercent, ValueModePercentColor:
		return true
	default:
		return false
	}
}

func (cc *ColumnChart) usesScoreColorScale() bool {
	return cc.resolvedValueMode() == ValueModePercentColor
}

func (cc *ColumnChart) shouldShowLegendSwatch() bool {
	return !cc.usesScoreColorScale()
}

func scoreColor(percent float64) string {
	switch {
	case percent >= 90:
		return "#10b981" // emerald-500
	case percent >= 70:
		return "#84cc16" // lime-500
	case percent >= 50:
		return "#fbbf24" // amber-400
	default:
		return "#f43f5e" // rose-500
	}
}

func (cc *ColumnChart) RenderPNG(opts PNGOptions) ([]byte, error) {
	return renderPNGViaSVG(cc, opts)
}

func (cc *ColumnChart) WriteSVG(path string) error {
	svg, err := cc.RenderSVG()
	if err != nil {
		return err
	}
	return writeTextFile(path, svg)
}

func (cc *ColumnChart) WritePNG(path string, opts PNGOptions) error {
	return writePNGViaSVG(path, cc, opts)
}

// ToSVG is deprecated and kept for compatibility with previous versions.
func (cc *ColumnChart) ToSVG() string {
	svg, err := cc.RenderSVG()
	if err != nil {
		return ""
	}
	return svg
}

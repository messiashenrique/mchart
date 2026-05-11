package mchart

import (
	"encoding/xml"
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestSpiderSVGIsValidAndMatchesGolden(t *testing.T) {
	chart := NewSpiderChart(
		[]string{"Ginecologia e Obstetrícia", "Pediatria", "Clínica Médica", "Clínica Cirúrgica", "Medicina de Família e Comunidade"},
		[]SpiderDataset{{Values: []float64{85, 60, 90, 45, 70}, Label: "Série A"}, {Values: []float64{50, 85, 45, 80, 60}, Label: "Série B"}},
	)

	svg, err := chart.RenderSVG()
	if err != nil {
		t.Fatalf("RenderSVG failed: %v", err)
	}
	assertValidSVG(t, svg)
	assertGoldenNormalized(t, "testdata/spider.svg.golden", svg)

	compat := chart.ToSVG()
	if normalizeWhitespace(svg) != normalizeWhitespace(compat) {
		t.Fatalf("ToSVG compatibility mismatch")
	}
}

func TestSpiderDarkThemeChangesKeyColors(t *testing.T) {
	chart := buildSpiderChartFixture()
	chart.Theme = SpiderThemeDark

	svg, err := chart.RenderSVG()
	if err != nil {
		t.Fatalf("RenderSVG failed: %v", err)
	}

	for _, expected := range []string{"#9ca3af", "#e5e7eb"} {
		if !strings.Contains(svg, expected) {
			t.Fatalf("expected dark theme color %s in spider svg", expected)
		}
	}
}

func TestSpiderRendersTitleWhenProvided(t *testing.T) {
	chart := buildSpiderChartFixture()
	chart.Title = "Título do Spider"

	svg, err := chart.RenderSVG()
	if err != nil {
		t.Fatalf("RenderSVG failed: %v", err)
	}

	if !strings.Contains(svg, "Título do Spider") {
		t.Fatalf("expected spider title in svg")
	}
	if !strings.Contains(svg, `class="title"`) {
		t.Fatalf("expected title class in spider svg")
	}
}

func TestColumnSVGIsValidAndMatchesGolden(t *testing.T) {
	chart := buildColumnChartFixture()
	svg, err := chart.RenderSVG()
	if err != nil {
		t.Fatalf("RenderSVG failed: %v", err)
	}
	assertValidSVG(t, svg)
	assertGoldenNormalized(t, "testdata/column.svg.golden", svg)
}

func TestColumnDarkThemeChangesKeyColors(t *testing.T) {
	chart := buildColumnChartFixture()
	chart.Theme = ColumnThemeDark

	svg, err := chart.RenderSVG()
	if err != nil {
		t.Fatalf("RenderSVG failed: %v", err)
	}

	for _, expected := range []string{"#111827", "#374151", "#f3f4f6"} {
		if !strings.Contains(svg, expected) {
			t.Fatalf("expected dark theme color %s in column svg", expected)
		}
	}
}

func TestColumnPercentModeUsesThresholdColorsAndNoLegendSwatch(t *testing.T) {
	chart := NewColumnChart("Teste", []ColumnCard{{
		Title: "A",
		Bars: []ColumnBar{
			{Label: "Muito grande para forçar legenda um", Value: 95},
			{Label: "Muito grande para forçar legenda dois", Value: 75},
			{Label: "Muito grande para forçar legenda tres", Value: 55},
			{Label: "Muito grande para forçar legenda quatro", Value: 35},
		},
	}})
	chart.ValueMode = ValueModePercent
	chart.Width = 900

	svg, err := chart.RenderSVG()
	if err != nil {
		t.Fatalf("RenderSVG failed: %v", err)
	}

	for _, expected := range []string{"#10b981", "#84cc16", "#fbbf24", "#f43f5e"} {
		if !strings.Contains(svg, expected) {
			t.Fatalf("expected threshold color %s in svg", expected)
		}
	}
	if strings.Contains(svg, `<rect x="`) {
		// prevent false positives: swatches are small rects with rx="2".
		if strings.Contains(svg, `rx="2" fill="#`) {
			t.Fatalf("percent mode legend should not include color swatches")
		}
	}
	if !strings.Contains(svg, "%") {
		t.Fatalf("expected percent symbol in value labels")
	}
}

func TestColumnNumberModeUsesPaletteAndLegendSwatch(t *testing.T) {
	bars := make([]ColumnBar, 12)
	for i := range bars {
		bars[i] = ColumnBar{Label: "Label muito grande " + ptNumber(float64(i)), Value: float64(i + 1)}
	}

	chart := NewColumnChart("Teste", []ColumnCard{{Title: "A", Bars: bars}})
	chart.ValueMode = ValueModeNumber
	chart.Width = 900

	svg, err := chart.RenderSVG()
	if err != nil {
		t.Fatalf("RenderSVG failed: %v", err)
	}

	if !strings.Contains(svg, defaultColumnPalette[0]) || !strings.Contains(svg, defaultColumnPalette[9]) {
		t.Fatalf("expected default palette colors in number mode")
	}
	if strings.Contains(svg, "%") {
		t.Fatalf("number mode should not include percent sign")
	}
	if !strings.Contains(svg, `rx="2" fill="#`) {
		t.Fatalf("number mode legend should include color swatches")
	}
}

func TestColumnManualColorOverridesAutoRule(t *testing.T) {
	chart := NewColumnChart("Teste", []ColumnCard{{
		Title: "A",
		Bars:  []ColumnBar{{Label: "L1", Value: 10, Color: "#123456"}},
	}})

	svgPercent, err := chart.RenderSVG()
	if err != nil {
		t.Fatalf("RenderSVG percent failed: %v", err)
	}
	if !strings.Contains(svgPercent, "#123456") {
		t.Fatalf("manual color should override in percent mode")
	}

	chart.ValueMode = ValueModeNumber
	svgNumber, err := chart.RenderSVG()
	if err != nil {
		t.Fatalf("RenderSVG number failed: %v", err)
	}
	if !strings.Contains(svgNumber, "#123456") {
		t.Fatalf("manual color should override in number mode")
	}
}

func TestColumnAutoHeightAdaptsToLegendRows(t *testing.T) {
	cards := []ColumnCard{{
		Title: "A",
		Bars: []ColumnBar{
			{Label: "Legenda longa item um", Value: 95},
			{Label: "Legenda longa item dois", Value: 85},
			{Label: "Legenda longa item tres", Value: 75},
			{Label: "Legenda longa item quatro", Value: 65},
			{Label: "Legenda longa item cinco", Value: 55},
			{Label: "Legenda longa item seis", Value: 45},
		},
	}}

	wide := NewColumnChart("Teste", cards)
	wide.ValueMode = ValueModeNumber
	wide.Width = 1800

	wideSVG, err := wide.RenderSVG()
	if err != nil {
		t.Fatalf("RenderSVG wide failed: %v", err)
	}
	wideH, ok := extractViewBoxHeight(wideSVG)
	if !ok {
		t.Fatalf("could not parse wide svg viewBox height")
	}
	if wideH >= 500 {
		t.Fatalf("expected compact auto height with one legend row, got %d", wideH)
	}

	narrow := NewColumnChart("Teste", cards)
	narrow.ValueMode = ValueModeNumber
	narrow.Width = 760

	narrowSVG, err := narrow.RenderSVG()
	if err != nil {
		t.Fatalf("RenderSVG narrow failed: %v", err)
	}
	narrowH, ok := extractViewBoxHeight(narrowSVG)
	if !ok {
		t.Fatalf("could not parse narrow svg viewBox height")
	}

	if narrowH <= wideH {
		t.Fatalf("expected taller svg when legend wraps to multiple rows, wide=%d narrow=%d", wideH, narrowH)
	}
}

func TestBarSVGIsValid(t *testing.T) {
	chart := buildBarChartFixture()
	svg, err := chart.RenderSVG()
	if err != nil {
		t.Fatalf("RenderSVG failed: %v", err)
	}
	assertValidSVG(t, svg)
	assertGoldenNormalized(t, "testdata/bar.svg.golden", svg)
}

func TestBarUsesDefaultPaletteWhenColorMissing(t *testing.T) {
	chart := NewBarChart("Teste", []BarItem{
		{Label: "A", Value: 30},
		{Label: "B", Value: 60},
	})
	chart.Tipo = BarTypeNumber

	svg, err := chart.RenderSVG()
	if err != nil {
		t.Fatalf("RenderSVG failed: %v", err)
	}

	if !strings.Contains(svg, defaultColumnPalette[0]) || !strings.Contains(svg, defaultColumnPalette[1]) {
		t.Fatalf("expected default palette colors in svg")
	}
}

func TestBarManualColorAndPercentMode(t *testing.T) {
	chart := NewBarChart("Teste", []BarItem{
		{Label: "A", Value: 44.5, Color: "#112233"},
		{Label: "B", Value: 0, Color: "#556677"},
	})
	chart.Tipo = BarTypePercent

	svg, err := chart.RenderSVG()
	if err != nil {
		t.Fatalf("RenderSVG failed: %v", err)
	}

	if !strings.Contains(svg, "#112233") {
		t.Fatalf("expected manual color in svg")
	}
	if !strings.Contains(svg, "%") {
		t.Fatalf("expected percent symbol in value labels")
	}
	if strings.Count(svg, `class="zero-mark"`) != 1 {
		t.Fatalf("expected one zero marker for one zero bar")
	}
	if !strings.Contains(svg, `class="zero-mark"`) || !strings.Contains(svg, `fill="#556677"`) {
		t.Fatalf("expected zero marker using the zero bar color")
	}
}

func TestBarDarkThemeChangesTypographyAndTrackColors(t *testing.T) {
	chart := NewBarChart("Teste", []BarItem{
		{Label: "A", Value: 20},
	})
	chart.Theme = BarThemeDark

	svg, err := chart.RenderSVG()
	if err != nil {
		t.Fatalf("RenderSVG failed: %v", err)
	}

	for _, color := range []string{"#f3f4f6", "#d1d5db", "#f9fafb", "#374151"} {
		if !strings.Contains(svg, color) {
			t.Fatalf("expected dark theme color %s in svg", color)
		}
	}
}

func TestDonutSVGIsValid(t *testing.T) {
	chart := buildDonutChartFixture()
	svg, err := chart.RenderSVG()
	if err != nil {
		t.Fatalf("RenderSVG failed: %v", err)
	}
	assertValidSVG(t, svg)
	assertGoldenNormalized(t, "testdata/donut.svg.golden", svg)
}

func TestDonutUsesDefaultPaletteWhenColorMissing(t *testing.T) {
	chart := NewDonutChart("Teste", []DonutSlice{
		{Label: "A", Value: 10},
		{Label: "B", Value: 20},
		{Label: "C", Value: 30},
	})

	svg, err := chart.RenderSVG()
	if err != nil {
		t.Fatalf("RenderSVG failed: %v", err)
	}

	if !strings.Contains(svg, defaultColumnPalette[0]) || !strings.Contains(svg, defaultColumnPalette[1]) {
		t.Fatalf("expected default palette colors in donut svg")
	}
	if !strings.Contains(svg, `class="slice-label-line"`) {
		t.Fatalf("expected outside percentage connector lines in donut svg")
	}
}

func TestDonutLongLabelsMoveLegendToBottom(t *testing.T) {
	chart := NewDonutChart("Teste", []DonutSlice{
		{Label: "Série com nome extremamente longo para quebrar layout lateral 1", Value: 10},
		{Label: "Série com nome extremamente longo para quebrar layout lateral 2", Value: 20},
		{Label: "Série com nome extremamente longo para quebrar layout lateral 3", Value: 30},
	})

	svg, err := chart.RenderSVG()
	if err != nil {
		t.Fatalf("RenderSVG failed: %v", err)
	}

	if !strings.Contains(svg, `class="legend legend-bottom"`) {
		t.Fatalf("expected bottom legend for long labels")
	}
	height, ok := extractViewBoxHeight(svg)
	if !ok {
		t.Fatalf("could not parse svg viewBox height")
	}
	if height == 520 {
		t.Fatalf("expected dynamic height for long bottom legend, got fixed default %d", height)
	}
}

func TestDonutDarkThemeChangesTypographyColors(t *testing.T) {
	chart := NewDonutChart("Teste", []DonutSlice{
		{Label: "A", Value: 20},
		{Label: "B", Value: 80},
	})
	chart.Theme = DonutThemeDark

	svg, err := chart.RenderSVG()
	if err != nil {
		t.Fatalf("RenderSVG failed: %v", err)
	}

	for _, color := range []string{"#f3f4f6", "#f9fafb", "#e5e7eb", "#111827"} {
		if !strings.Contains(svg, color) {
			t.Fatalf("expected dark theme color %s in donut svg", color)
		}
	}
}

func TestSplineSVGIsValid(t *testing.T) {
	chart := buildSplineChartFixture()
	svg, err := chart.RenderSVG()
	if err != nil {
		t.Fatalf("RenderSVG failed: %v", err)
	}
	assertValidSVG(t, svg)
	assertGoldenNormalized(t, "testdata/spline.svg.golden", svg)
}

func TestSplineUsesDefaultPaletteWhenColorMissing(t *testing.T) {
	chart := NewSplineChart("Teste", []string{"Jan", "Fev", "Mar"}, []SplineSeries{
		{Label: "A", Values: []float64{10, 20, 30}},
		{Label: "B", Values: []float64{20, 30, 10}},
	})

	svg, err := chart.RenderSVG()
	if err != nil {
		t.Fatalf("RenderSVG failed: %v", err)
	}

	if !strings.Contains(svg, defaultColumnPalette[0]) || !strings.Contains(svg, defaultColumnPalette[1]) {
		t.Fatalf("expected default palette colors in spline svg")
	}
}

func TestSplineDarkThemeChangesKeyColors(t *testing.T) {
	chart := buildSplineChartFixture()
	chart.Theme = SplineThemeDark

	svg, err := chart.RenderSVG()
	if err != nil {
		t.Fatalf("RenderSVG failed: %v", err)
	}

	for _, expected := range []string{"#f3f4f6", "#4b5563", "#9ca3af", "#d1d5db", "#e5e7eb"} {
		if !strings.Contains(svg, expected) {
			t.Fatalf("expected dark theme color %s in spline svg", expected)
		}
	}
}

func TestWriteSVGAndWritePNG(t *testing.T) {
	chart := NewSpiderChart(
		[]string{"A", "B", "C"},
		[]SpiderDataset{{Values: []float64{20, 60, 90}, Label: "Série A"}},
	)

	tmpDir := t.TempDir()
	svgPath := filepath.Join(tmpDir, "chart.svg")
	pngPath := filepath.Join(tmpDir, "chart.png")

	if err := chart.WriteSVG(svgPath); err != nil {
		t.Fatalf("WriteSVG failed: %v", err)
	}
	if _, err := os.Stat(svgPath); err != nil {
		t.Fatalf("WriteSVG did not create file: %v", err)
	}

	if err := chart.WritePNG(pngPath, PNGOptions{Backend: PNGBackendCanvas, Width: 320, Height: 240}); err != nil {
		t.Fatalf("WritePNG failed: %v", err)
	}

	file, err := os.Open(pngPath)
	if err != nil {
		t.Fatalf("open png: %v", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		t.Fatalf("decode png: %v", err)
	}

	if img.Bounds().Dx() != 320 || img.Bounds().Dy() != 240 {
		t.Fatalf("unexpected image size: got %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
	}
}

func assertValidSVG(t *testing.T, svg string) {
	t.Helper()
	if strings.TrimSpace(svg) == "" {
		t.Fatal("svg is empty")
	}

	type svgRoot struct {
		XMLName xml.Name
	}
	var root svgRoot
	if err := xml.Unmarshal([]byte(svg), &root); err != nil {
		t.Fatalf("svg is not valid xml: %v", err)
	}
	if strings.ToLower(root.XMLName.Local) != "svg" {
		t.Fatalf("expected root <svg>, got <%s>", root.XMLName.Local)
	}
}

func assertGoldenNormalized(t *testing.T, goldenPath string, got string) {
	t.Helper()
	expectedBytes, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden file: %v", err)
	}

	expected := normalizeWhitespace(string(expectedBytes))
	actual := normalizeWhitespace(got)

	if expected != actual {
		t.Fatalf("golden mismatch for %s", goldenPath)
	}
}

func normalizeWhitespace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func extractViewBoxHeight(svg string) (int, bool) {
	re := regexp.MustCompile(`viewBox="0 0 \d+ (\d+)"`)
	m := re.FindStringSubmatch(svg)
	if len(m) != 2 {
		return 0, false
	}
	h, err := strconv.Atoi(m[1])
	if err != nil {
		return 0, false
	}
	return h, true
}

# mchart

🇧🇷 **[Veja em Português](README-pt-BR.md)**

Go library for static chart generation (SVG/PNG), focused on reports and PDFs.

- Repository: `https://github.com/messiashenrique/mchart`
- Module: `github.com/messiashenrique/mchart`


### Installation

```bash
go get github.com/messiashenrique/mchart
```

### Quick start

```go
package main

import (
	"os"

	"github.com/messiashenrique/mchart"
)

func main() {
	c := mchart.NewBarChart("Summary", []mchart.BarItem{
		{Label: "Part A", Value: 70},
		{Label: "Part B", Value: 45},
	})
	c.Tipo = mchart.BarTypePercent

	svg, _ := c.RenderSVG()
	_ = os.WriteFile("bar.svg", []byte(svg), 0o644)
}
```

### Chart types

#### 1) ColumnChart

```go
cards := []mchart.ColumnCard{
  {
    Title: "Emissão por setor",
    Bars: []mchart.ColumnBar{
      {Label: "Escritório", Value: 100.0},
      {Label: "Auditório", Value: 75.0},
      {Label: "Almoxarifado", Value: 56.25},
      {Label: "Matriz", Value: 68.75},
      {Label: "Fiscal", Value: 75.0},
      {Label: "Financeiro", Value: 0.0},
    },
  },
}

chart := mchart.NewColumnChart("Portarias emitidas.", cards)
chart.ValueMode = mchart.ValueModeNumber
```

![image](https://raw.githubusercontent.com/messiashenrique/mchart/main/images/columns.png)

#### 2) SpiderChart

```go
labels := []string{"Clínica Médica", "Cirurgia", "Pediatria", "G.O.", "M.F.C."}
datasets := []mchart.SpiderDataset{
  {Label: "Turma A", Values: []float64{70, 85, 60, 78, 90}},
  {Label: "Turma B", Values: []float64{85, 72, 80, 69, 75}},
}
chart := mchart.NewSpiderChart(labels, datasets)
chart.Title = "Desempenho por área"
```

![image](https://raw.githubusercontent.com/messiashenrique/mchart/main/images/spider.png)

#### 3) BarChart

```go
items := []mchart.BarItem{
  {Label: "Parte específica (10/20)", Value: 50.0, Color: "#10b981"},
  {Label: "Parte não específica (19/80)", Value: 23.75, Color: "#f97316"},
  {Label: "Percentual global (29/100)", Value: 29.0, Color: "#d6bf8f"},
}
chart := mchart.NewBarChart("Resumo de acertos", items)
chart.Tipo = mchart.BarTypePercent
```

![image](https://raw.githubusercontent.com/messiashenrique/mchart/main/images/bars.png)

#### 4) DonutChart

```go
slices := []mchart.DonutSlice{
  {Label: "series-1", Value: 26},
  {Label: "series-2", Value: 32},
  {Label: "series-3", Value: 24},
}
chart := mchart.NewDonutChart("Resumo por série", slices)
```

![image](https://raw.githubusercontent.com/messiashenrique/mchart/main/images/donut.png)

#### 5) SplineChart

```go
labels := []string{"Jan", "Fev", "Mar", "Abr", "Mai", "Jun", "Jul"}
series := []mchart.SplineSeries{
  {Label: "series1", Values: []float64{31, 40, 28, 51, 42, 108, 100}},
  {Label: "series2", Values: []float64{11, 32, 45, 32, 34, 52, 41}},
}
chart := mchart.NewSplineChart("Resumo por mês", labels, series)
chart.Theme = mchart.SplineThemeLight
```

![image](https://raw.githubusercontent.com/messiashenrique/mchart/main/images/spline.png)

### Themes and default colors

- Light theme is the default across charts.
- Charts with dark mode support expose `Theme` (`...ThemeDark`).
- If no explicit color is provided, the default palette from `colors.go` is used.

### SVG/PNG output

All charts support:

- `RenderSVG() (string, error)`
- `RenderPNG(opts PNGOptions) ([]byte, error)`
- `WriteSVG(path string) error`
- `WritePNG(path string, opts PNGOptions) error`

### Runnable examples

Generate the 5 demo SVG files:

```bash
go run ./examples/all-charts
```

Generated output: `examples/all-charts/out/`.

### Version

`v0.1.1`

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	mchart "github.com/messiashenrique/mchart"
)

func main() {
	type chartOutput struct {
		name string
		svg  string
	}

	outputs := make([]chartOutput, 0, 5)

	columnSVG, err := buildColumnExampleChart().RenderSVG()
	if err != nil {
		log.Fatalf("falha ao gerar svg de colunas: %v", err)
	}
	outputs = append(outputs, chartOutput{name: "column_chart.svg", svg: columnSVG})

	spiderSVG, err := buildSpiderExampleChart().RenderSVG()
	if err != nil {
		log.Fatalf("falha ao gerar svg de spider: %v", err)
	}
	outputs = append(outputs, chartOutput{name: "spider_chart.svg", svg: spiderSVG})

	barSVG, err := buildBarExampleChart().RenderSVG()
	if err != nil {
		log.Fatalf("falha ao gerar svg de barras horizontais: %v", err)
	}
	outputs = append(outputs, chartOutput{name: "bar_chart.svg", svg: barSVG})

	donutSVG, err := buildDonutExampleChart().RenderSVG()
	if err != nil {
		log.Fatalf("falha ao gerar svg de rosca: %v", err)
	}
	outputs = append(outputs, chartOutput{name: "donut_chart.svg", svg: donutSVG})

	splineSVG, err := buildSplineExampleChart().RenderSVG()
	if err != nil {
		log.Fatalf("falha ao gerar svg de spline: %v", err)
	}
	outputs = append(outputs, chartOutput{name: "spline_chart.svg", svg: splineSVG})

	outputDir := filepath.Join("examples", "all-charts", "out")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		log.Fatalf("erro ao criar diretório de saída %s: %v", outputDir, err)
	}

	for _, output := range outputs {
		if output.svg == "" {
			log.Fatalf("falha ao gerar SVG: %s", output.name)
		}
		outputFile := filepath.Join(outputDir, output.name)
		if err := os.WriteFile(outputFile, []byte(output.svg), 0o644); err != nil {
			log.Fatalf("erro ao salvar %s: %v", outputFile, err)
		}
		fmt.Printf("Sucesso! Arquivo salvo em: %s\n", outputFile)
	}
}

func buildColumnExampleChart() *mchart.ColumnChart {
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
	return chart
}

func buildSpiderExampleChart() *mchart.SpiderChart {
	labels := []string{"Clínica Médica", "Cirurgia", "Pediatria", "G.O.", "M.F.C."}
	datasets := []mchart.SpiderDataset{
		{Label: "Turma A", Values: []float64{70, 85, 60, 78, 90}},
		{Label: "Turma B", Values: []float64{85, 72, 80, 69, 75}},
	}
	chart := mchart.NewSpiderChart(labels, datasets)
	chart.Title = "Desempenho por área"
	chart.Theme = mchart.SpiderThemeLight
	return chart
}

func buildBarExampleChart() *mchart.BarChart {
	items := []mchart.BarItem{
		{Label: "Parte específica (10/20)", Value: 50.0, Color: "#10b981"},
		{Label: "Parte não específica (19/80)", Value: 23.75, Color: "#f97316"},
		{Label: "Percentual global (29/100)", Value: 29.0, Color: "#d6bf8f"},
	}

	chart := mchart.NewBarChart("Resumo de acertos", items)
	chart.Tipo = mchart.BarTypePercent
	chart.MinWidth = 760
	chart.Padding = 24
	return chart
}

func buildDonutExampleChart() *mchart.DonutChart {
	slices := []mchart.DonutSlice{
		{Label: "series-1", Value: 26},
		{Label: "series-2", Value: 32},
		{Label: "series-3", Value: 24},
	}
	chart := mchart.NewDonutChart("Resumo por série", slices)
	chart.Theme = mchart.DonutThemeLight
	return chart
}

func buildSplineExampleChart() *mchart.SplineChart {
	labels := []string{"Jan", "Fev", "Mar", "Abr", "Mai", "Jun", "Jul"}
	series := []mchart.SplineSeries{
		{Label: "series1", Values: []float64{31, 40, 28, 51, 42, 108, 100}},
		{Label: "series2", Values: []float64{11, 32, 45, 32, 34, 52, 41}},
	}
	chart := mchart.NewSplineChart("Resumo por mês", labels, series)
	return chart
}

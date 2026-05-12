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
		name  string
		chart interface {
			RenderSVG() (string, error)
			RenderPNG(opts mchart.PNGOptions) ([]byte, error)
		}
	}

	outputs := make([]chartOutput, 0, 6)

	outputs = append(outputs, chartOutput{name: "column_chart", chart: buildColumnExampleChart()})
	outputs = append(outputs, chartOutput{name: "spider_chart", chart: buildSpiderExampleChart()})
	outputs = append(outputs, chartOutput{name: "bar_chart", chart: buildBarExampleChart()})
	outputs = append(outputs, chartOutput{name: "donut_chart", chart: buildDonutExampleChart()})
	outputs = append(outputs, chartOutput{name: "spline_chart", chart: buildSplineExampleChart()})
	outputs = append(outputs, chartOutput{name: "funnel_chart", chart: buildFunnelExampleChart()})

	outputDir := filepath.Join("examples", "all-charts", "out")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		log.Fatalf("erro ao criar diretório de saída %s: %v", outputDir, err)
	}

	for _, output := range outputs {
		svg, err := output.chart.RenderSVG()
		if err != nil {
			log.Fatalf("falha ao gerar svg de %s: %v", output.name, err)
		}

		svgFile := filepath.Join(outputDir, output.name+".svg")
		if err := os.WriteFile(svgFile, []byte(svg), 0o644); err != nil {
			log.Fatalf("erro ao salvar %s: %v", svgFile, err)
		}
		fmt.Printf("Sucesso! Arquivo salvo em: %s\n", svgFile)

		pngData, err := output.chart.RenderPNG(mchart.PNGOptions{})
		if err != nil {
			log.Fatalf("falha ao gerar png de %s: %v", output.name, err)
		}
		pngFile := filepath.Join(outputDir, output.name+".png")
		if err := os.WriteFile(pngFile, pngData, 0o644); err != nil {
			log.Fatalf("erro ao salvar %s: %v", pngFile, err)
		}
		fmt.Printf("Sucesso! Arquivo salvo em: %s\n", pngFile)
	}
}

func buildColumnExampleChart() *mchart.ColumnChart {
	cards := []mchart.ColumnCard{
		{
			Title: "Emissão por setor",
			Bars: []mchart.ColumnBar{
				// {Label: "Q1", Value: 100.0},
				// {Label: "Q2", Value: 75.0},
				// {Label: "Q3", Value: 56.25},
				// {Label: "Q4", Value: 68.75},
				// {Label: "Q5", Value: 75.0},
				// {Label: "Q6TYMQi", Value: 0.0},
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

func buildFunnelExampleChart() *mchart.FunnelChart {
	sections := []mchart.FunnelSection{
		{Label: "> 4 pontos", Value: 700},
		{Label: "> de 5 pontos", Value: 600},
		{Label: "> de 6 pontos", Value: 500},
		{Label: "> de 7 pontos", Value: 400},
		{Label: "> de 8 pontos", Value: 300},
		{Label: "> de 9 pontos", Value: 70},
	}
	chart := mchart.NewFunnelChart("Distribuição de estudantes por notas", sections)
	chart.ValueMode = mchart.FunnelValueInteger
	chart.Theme = mchart.FunnelThemeLight
	return chart
}

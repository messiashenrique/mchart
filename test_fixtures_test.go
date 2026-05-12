package mchart

import "math"

func buildColumnChartFixture() *ColumnChart {
	cards := []ColumnCard{
		{
			Title: "Gastos por setor",
			Bars: []ColumnBar{
				{Label: "Escritório", Value: 100.0},
				{Label: "Auditório", Value: 75.0},
				{Label: "Almoxarifado", Value: 56.25},
				{Label: "Matriz", Value: 68.75},
				{Label: "Fiscal", Value: 75.0},
				{Label: "Financeiro", Value: 0.0},
			},
		},
		{
			Title: "Itens diversos",
			Bars: []ColumnBar{
				{Value: 100.0},
				{Value: 76.92},
				{Value: 53.85},
				{Value: 53.85},
				{Value: 61.54},
				{Value: 46.15},
			},
		},
	}

	chart := NewColumnChart("Conjunto de gráficos de colunas.", cards)
	chart.MinWidth = 520
	chart.Height = 500
	chart.Padding = 16
	chart.CardGap = 16
	chart.ValueMode = ValueModeNumber

	for ci := range chart.Cards {
		for bi := range chart.Cards[ci].Bars {
			chart.Cards[ci].Bars[bi].Value = math.Round(chart.Cards[ci].Bars[bi].Value*100) / 100
		}
	}

	return chart
}

func buildSpiderChartFixture() *SpiderChart {
	labels := []string{
		"Ginecologia e Obstetrícia",
		"Pediatria",
		"Clínica Médica",
		"Clínica Cirúrgica",
		"Medicina de Família e Comunidade",
	}
	datasets := []SpiderDataset{
		{Values: []float64{85, 60, 90, 45, 70}, Label: "Série A"},
		{Values: []float64{50, 85, 45, 80, 60}, Label: "Série B"},
	}

	return NewSpiderChart(labels, datasets)
}

func buildBarChartFixture() *BarChart {
	items := []BarItem{
		{Label: "Parte específica (10/20)", Value: 50},
		{Label: "Parte não específica (19/80)", Value: 23.75},
		{Label: "Percentual global (29/100)", Value: 29},
	}

	chart := NewBarChart("Resumo de acertos", items)
	chart.Tipo = BarTypePercent
	chart.MinWidth = 760
	chart.Padding = 24
	return chart
}

func buildDonutChartFixture() *DonutChart {
	slices := []DonutSlice{
		{Label: "series-1", Value: 26},
		{Label: "series-2", Value: 32},
		{Label: "series-3", Value: 24},
		{Label: "series-4", Value: 10},
		{Label: "series-5", Value: 9},
	}

	chart := NewDonutChart("Resumo por série", slices)
	chart.Width = 820
	chart.Height = 520
	return chart
}

func buildSplineChartFixture() *SplineChart {
	labels := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul"}
	series := []SplineSeries{
		{Label: "series1", Values: []float64{31, 40, 28, 51, 42, 108, 100}},
		{Label: "series2", Values: []float64{11, 32, 45, 32, 34, 52, 41}},
	}

	chart := NewSplineChart("Resumo por mês", labels, series)
	chart.Width = 820
	chart.Height = 520
	return chart
}

func buildFunnelChartFixture() *FunnelChart {
	sections := []FunnelSection{
		{Label: "North America", Value: 700},
		{Label: "South America", Value: 600},
		{Label: "Africa", Value: 500},
		{Label: "Asia", Value: 400},
		{Label: "Oceania", Value: 300},
		{Label: "Europe", Value: 200},
	}

	chart := NewFunnelChart("Pipeline por região", sections)
	chart.Width = 820
	chart.Height = 540
	return chart
}

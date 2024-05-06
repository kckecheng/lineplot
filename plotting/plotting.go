package plotting

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func Plot(output, title, xTitle, yTitle string, xAxis []any, seriesNames []string, seriesItems [][]any, width, height int32, smooth bool) error {
	// check parameters: some parameters are optional, some are mandatory
	if output == "" {
		return errors.New("output plotting file must be specified")
	}
	if title == "" {
		title = "unamed line chart"
	}
	if xTitle == "" {
		xTitle = "x"
	}
	if yTitle == "" {
		yTitle = "y"
	}

	if width < 0 {
		return errors.New("chart width must be larger than 0, default 2400px")
	}
	if width == 0 {
		width = 2400
	}
	if height < 0 {
		return errors.New("chart height must be larger than 0, default 500px")
	}
	if height == 0 {
		height = 500
	}

	if len(seriesItems) == 0 {
		return errors.New("at least one series should be defined")
	}

	if len(seriesNames) == 0 {
		for i := range len(seriesItems) {
			seriesNames = append(seriesNames, fmt.Sprintf("series%d", i+1))
		}
	}

	count := len(seriesItems[0])
	for _, v := range seriesItems {
		if count != len(v) {
			return errors.New("the num. of data items each series contains should be the same")
		}
	}

	if len(xAxis) == 0 {
		for i := range count {
			xAxis = append(xAxis, i+1)
		}
	} else {
		if len(xAxis) != count {
			return errors.New("the num. of x axis data items should be the same as the num. of data items each series contains")
		}
	}

	if len(seriesNames) != len(seriesItems) {
		return errors.New("the num. of series names must be the same as the num. of series")
	}

	line := charts.NewLine()
	// global options
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Width: fmt.Sprintf("%dpx", width), Height: fmt.Sprintf("%dpx", height)}),
		charts.WithTitleOpts(opts.Title{Title: title}),
		charts.WithXAxisOpts(opts.XAxis{Name: xTitle}),
		charts.WithYAxisOpts(opts.YAxis{Name: yTitle, Show: true}),
		charts.WithTooltipOpts(opts.Tooltip{Show: true, Trigger: "axis"}),
	)
	line.SetXAxis(xAxis)

	// populate each series
	for i, name := range seriesNames {
		data := make([]opts.LineData, 0)
		for _, v := range seriesItems[i] {
			data = append(data, opts.LineData{Value: v})
		}
		line = line.AddSeries(name, data)
	}
	// enable series options
	// min/avg./max are only enabled for chart with single line
	if len(seriesNames) == 1 {
		line.SetSeriesOptions(
			charts.WithMarkLineNameTypeItemOpts(opts.MarkLineNameTypeItem{Name: "Minimum", Type: "min"}),
			charts.WithMarkLineNameTypeItemOpts(opts.MarkLineNameTypeItem{Name: "Average", Type: "average"}),
			charts.WithMarkLineNameTypeItemOpts(opts.MarkLineNameTypeItem{Name: "Maximum", Type: "max"}),
		)
	}
	line.SetSeriesOptions(
		charts.WithMarkPointStyleOpts(opts.MarkPointStyle{Label: &opts.Label{Show: true}}),
		charts.WithLineChartOpts(opts.LineChart{ShowSymbol: true}),
		charts.WithLabelOpts(opts.Label{Show: true}),
	)
	if smooth {
		line.SetSeriesOptions(
			charts.WithLineChartOpts(opts.LineChart{Smooth: true}),
		)
	}

	// create html file for holding the line chart
	page := components.NewPage()
	page.PageTitle = title
	page.AddCharts(line)
	f, err := os.Create(output)
	if err != nil {
		return err
	}
	return page.Render(io.MultiWriter(f))
}

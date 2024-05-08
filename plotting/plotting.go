package plotting

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// LoadData: load csv data, parse them, fill in correspondingly vars and return them
// return:
// - []any   : x-axis coordinate point
// - []string: series names
// - [][]any : series items
// - error   : error
func LoadData(csvf string, c1x, r1h bool) ([]any, []string, [][]any, error) {
	var (
		xAxis       []any
		seriesNames []string
		seriesItems [][]any
		err         error
	)

	f, err := os.Open(csvf)
	if err != nil {
		return xAxis, seriesNames, seriesItems, err
	}
	defer f.Close()

	csvr := csv.NewReader(f)
	records, err := csvr.ReadAll()
	if err != nil {
		return xAxis, seriesNames, seriesItems, err
	}

	// whether to use the 1st column for x axis
	cindex := 0
	if c1x {
		cindex = 1
	}
	// whether to use the 1st row for heading
	rindex := 0
	if r1h {
		rindex = 1
	}

	// fill in seriesNames if the 1st row is used for heading
	if r1h {
		for _, name := range records[0][cindex:] {
			seriesNames = append(seriesNames, name)
		}
	}

	// fill in xAxis if the 1st column is used for x-axis coordinate point
	if c1x {
		for _, v := range records[rindex:] {
			xAxis = append(xAxis, v[0])
		}
	}

	// init seriesItems
	itemNum := len(records[rindex:])
	for range len(records[0][cindex:]) {
		items := make([]any, itemNum)
		seriesItems = append(seriesItems, items)
	}

	// fill in seriesItems
	for i, row := range records[rindex:] {
		for j := range row[cindex:] {
			seriesItems[j][i] = records[i+rindex][j+cindex]
		}
	}
	return xAxis, seriesNames, seriesItems, err
}

func LinePlot(title, xTitle, yTitle string, xAxis []any, seriesNames []string, seriesItems [][]any, width, height int32, smooth bool) (*charts.Line, error) {
	// check parameters: some parameters are optional, some are mandatory
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
		return nil, errors.New("chart width must be larger than 0, default 2400px")
	}
	if width == 0 {
		width = 2400
	}
	if height < 0 {
		return nil, errors.New("chart height must be larger than 0, default 500px")
	}
	if height == 0 {
		height = 500
	}

	if len(seriesItems) == 0 {
		return nil, errors.New("at least one series should be defined")
	}

	if len(seriesNames) == 0 {
		for i := range len(seriesItems) {
			seriesNames = append(seriesNames, fmt.Sprintf("series%d", i+1))
		}
	}

	count := len(seriesItems[0])
	for _, v := range seriesItems {
		if count != len(v) {
			return nil, errors.New("the num. of data items each series contains should be the same")
		}
	}

	if len(xAxis) == 0 {
		for i := range count {
			xAxis = append(xAxis, i+1)
		}
	} else {
		if len(xAxis) != count {
			return nil, errors.New("the num. of x axis data items should be the same as the num. of data items each series contains")
		}
	}

	if len(seriesNames) != len(seriesItems) {
		return nil, errors.New("the num. of series names must be the same as the num. of series")
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
	return line, nil
}

func GenCharts(output string, lines []*charts.Line) error {
	if output == "" {
		return errors.New("output plotting file must be specified")
	}

	page := components.NewPage()
	page.PageTitle = "Line Charts"
	for _, line := range lines {
		page.AddCharts(line)
	}

	// create html file for holding the chart
	f, err := os.Create(output)
	if err != nil {
		return err
	}
	return page.Render(io.MultiWriter(f))
}

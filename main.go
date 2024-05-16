package main

import (
	"fmt"
	"os"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/kckecheng/lineplot/plotting"
	"github.com/kckecheng/lineplot/webserver"
	flag "github.com/spf13/pflag"
)

const (
	ErrParams = 1 << iota
	ErrLoadData
	ErrPlotting
	ErrExample
	ErrGenF
	ErrWebStart
)

var (
	name           string // output file name
	c1x            bool   // whether to use the 1st column for x-axis coordinate point
	r1h            bool   // whether to use the 1st row for heading(where to get series names)
	smooth         bool   // whether to draw smooth line, if true, the data point lable will not be shown
	example        bool   // show usage example
	width, height  int32  = 2400, 500
	xTitle, yTitle string
	title          []string
	dataf          []string
	runweb         bool // web to run as a web service
)

func showExample() {
	e := `

Example 1: use the 1st column for x axis and the 1st row for series names

- data.csv:

	X, Series 1, Series 2
	1, 100, 200
	2, 210, 210
	3, 89, 300

- command: lineplot -o example1.html -t "example 1" -d data.csv --c1x --r1h

Example 2: use default x axis and the 1st row for series names

- data.csv:

	Series 1, Series 2
	100, 200
	210, 210
	89, 300

- command: lineplot -o example2.html -t "example 2" -d data.csv --r1h

Example 3: use default x axis and defaut series names

- data.csv:

	100
	210
	89

- command: lineplot -o example3.html -t "example 3" -d data.csv

Example 4: plot with multiple csv files, and all use the 1st column for x axis and the 1st row for series names

- data1.csv:

	X, Series 1, Series 2
	1, 100, 200
	2, 210, 210
	3, 89, 300

- data2.csv

	X, Series 1, Series 2
	1, 105, 195
	2, 215, 220
	3, 130, 260

- command:
	- w/ automatic title: lineplot -o example4.html -d data1.csv -d data2.csv --c1x --r1h
	- w/ specific title : lineplot -o example4.html -t line1 -d data1.csv -t line2 -d data2.csv --c1x --r1h

Example 5: plot by leveraging the web interface

- Command: lineplot --runweb
	`
	fmt.Fprintf(os.Stderr, "%s\n", e)
}

func cliPlot() {
	if len(title) != 0 && len(title) != len(dataf) {
		fmt.Fprintln(os.Stderr, "the num. of titles should be empty or the same as the num. of csv files")
		os.Exit(ErrParams)
	}

	if len(title) == 0 && len(dataf) > 0 {
		for _, v := range dataf {
			title = append(title, fmt.Sprintf("line chart for %s", v))
		}
	}

	var lines []*charts.Line
	for i, v := range dataf {
		xAxis, seriesNames, seriesItems, err := plotting.LoadData(v, c1x, r1h)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fail to load data for %s\n", v)
			os.Exit(ErrLoadData)
		}
		line, err := plotting.LinePlot(title[i], xTitle, yTitle, xAxis, seriesNames, seriesItems, width, height, smooth)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fail to plot for %s\n", v)
			os.Exit(ErrPlotting)
		}
		lines = append(lines, line)
	}

	fp, err := os.Create(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fail to create file %s for holding charts\n", name)
		os.Exit(ErrGenF)
	}
	defer fp.Close()

	err = plotting.GenCharts(fp, lines)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fail to add charts to file %s\n", name)
		os.Exit(ErrGenF)
	}

	fmt.Fprintf(os.Stdout, "plotting file: %s\n", name)
	os.Exit(0)
}

func main() {
	flag.StringVarP(&name, "output", "o", "lines.html", "output file used for holding the chart")
	flag.StringSliceVarP(&title, "title", "t", []string{}, "chart title")
	flag.StringVarP(&xTitle, "xtitle", "x", "X", "x-axis title")
	flag.StringVarP(&yTitle, "ytitle", "y", "Y", "y-axis title")
	flag.BoolVar(&c1x, "c1x", false, "use the 1st column from the csv data as x-axis coordinate point")
	flag.BoolVar(&r1h, "r1h", false, "use the 1st row from the csv data as Y series names")
	flag.BoolVarP(&smooth, "smooth", "s", false, "draw smooth line(no data point mark)")
	flag.Int32Var(&width, "width", 2400, "chart width")
	flag.Int32Var(&height, "height", 500, "chart height")
	flag.StringSliceVarP(&dataf, "data", "d", []string{}, "data for plotting")
	flag.BoolVar(&example, "example", false, "show example on how to use the tool")
	flag.BoolVar(&runweb, "runweb", false, "whether to run as a web application")
	flag.Parse()

	if example {
		showExample()
		os.Exit(ErrExample)
	}

	if runweb {
		err := webserver.Start()
		if err != nil {
			fmt.Fprintln(os.Stderr, "fail to run the appliaction as a web application")
			os.Exit(ErrWebStart)
		}
	} else {
		cliPlot()
	}
}

package main

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/kckecheng/lineplot/plotting"
	flag "github.com/spf13/pflag"
)

const (
	ErrLoadData = 1 << iota
	ErrPlotting
	ErrExample
)

var (
	name, title, xTitle, yTitle string
	xAxis                       []any
	// c1x - whether to use the 1st column of data as x-axis coordinate point
	// r1h - whether to use the 1st row as series names
	c1x, r1h, smooth, example bool
	seriesNames               []string
	seriesItems               [][]any
	width, height             int32 = 2400, 500
	dataf                     string
)

func showExample() {
	e := `

Example 1:

- data.csv:

	X, Series 1, Series 2
	1, 100, 200
	2, 210, 210
	3, 89, 300

- overview:
	- the 1st column is used as x-axis coordinate point
	- the 1st row is used for heading

- command: lineplot -o example1.html -t "example 1" -d data.csv --c1x --r1h

Example 2:

- data.csv:

	Series 1, Series 2
	100, 200
	210, 210
	89, 300

- overview:
	- there is no specific data for x-axis coordinate point
	- the 1st row is used for heading

- command: lineplot -o example2.html -t "example 2" -d data.csv --r1h

Example 3:

- data.csv:

	100
	210
	89

- overview:
	- there is no specific data for x-axis coordinate point
	- there is no heading

- command: lineplot -o example3.html -t "example 3" -d data.csv
	`
	fmt.Fprintf(os.Stderr, "%s\n", e)
}

func loadData() error {
	// load csv data, parse them, and fill in correspondingly vars
	f, err := os.Open(dataf)
	if err != nil {
		return err
	}
	defer f.Close()

	csvr := csv.NewReader(f)
	records, err := csvr.ReadAll()
	if err != nil {
		return err
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
	return nil
}

func main() {
	var err error

	flag.StringVarP(&name, "output", "o", "lines.html", "output file used for holding the chart")
	flag.StringVarP(&title, "title", "t", "unamed line chart", "chart title")
	flag.StringVarP(&xTitle, "xtitle", "x", "X", "x-axis title")
	flag.StringVarP(&yTitle, "ytitle", "y", "Y", "y-axis title")
	flag.BoolVar(&c1x, "c1x", false, "use the 1st column from the csv data as x-axis coordinate point")
	flag.BoolVar(&r1h, "r1h", false, "use the 1st row from the csv data as Y series names")
	flag.BoolVarP(&smooth, "smooth", "s", false, "draw smooth line")
	flag.Int32Var(&width, "width", 2400, "chart width")
	flag.Int32Var(&height, "height", 500, "chart height")
	flag.StringVarP(&dataf, "data", "d", "data.csv", "data for plotting")
	flag.BoolVar(&example, "example", false, "show example on how to use the tool")
	flag.Parse()

	if example {
		showExample()
		os.Exit(ErrExample)
	}

	err = loadData()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot load data from %s\n", dataf)
		os.Exit(ErrLoadData)
	}

	// fmt.Println(name)
	// fmt.Println(title)
	// fmt.Println(xTitle)
	// fmt.Println(yTitle)
	// fmt.Println(xAxis)
	// fmt.Println(seriesNames)
	// fmt.Println(width)
	// fmt.Println(height)
	// for _, series := range seriesItems {
	// 	fmt.Println(series)
	// }
	err = plotting.Plot(name, title, xTitle, yTitle, xAxis, seriesNames, seriesItems, width, height, smooth)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fail to plot due to error: %s\n", err.Error())
		os.Exit(ErrPlotting)
	}

	fmt.Fprintf(os.Stdout, "plotting file: %s\n", name)
}

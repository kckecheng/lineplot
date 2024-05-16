package webserver

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/kckecheng/lineplot/plotting"
)

const (
	STATIC  = "webserver/static"
	OUTPUT  = "./output"
	TMP     = "./tmp"
	MAXSIZE = 20 * 1024 * 1024
)

type handlerMapper struct {
	src     *multipart.FileHeader
	srcName string
	dst     *os.File
	dstName string
}

func init() {
	tryCreateDir(OUTPUT)
	tryCreateDir(TMP)
}

func tryCreateDir(d string) {
	var err error
	_, err = os.Stat(d)
	if err != nil && os.IsNotExist(err) {
		err = os.Mkdir(d, 0755)
		if err != nil {
			panic(fmt.Sprintf("fail to create dir %s", d))
		}
	}
}

func staticFile(f string) string {
	return path.Join(STATIC, f)
}

func getFormDefault(c *gin.Context, k, vdefault string) string {
	v := strings.TrimSpace(c.PostForm(k))
	if v != "" {
		return v
	}
	if vdefault == "" {
		panic("default value must be defined")
	}
	return vdefault
}

func getBool(s string) bool {
	s = strings.TrimSpace(s)
	if s == "true" || s == "True" || s == "TRUE" {
		return true
	}
	return false
}

func plotCharts(dataf, origName []string, xTitle, yTitle string, c1x, r1h, smooth bool) (string, error) {
	var lines []*charts.Line
	for i, v := range dataf {
		xAxis, seriesNames, seriesItems, err := plotting.LoadData(v, c1x, r1h)
		if err != nil {
			return "", errors.New(fmt.Sprintf("fail to load data for %s\n", v))
		}

		title := fmt.Sprintf("chart for %s", origName[i])
		// plot with fixed size
		line, err := plotting.LinePlot(title, xTitle, yTitle, xAxis, seriesNames, seriesItems, 2400, 500, smooth)
		if err != nil {
			return "", errors.New(fmt.Sprintf("fail to plot for %s\n", origName[i]))
		}
		lines = append(lines, line)
	}

	dstFile, err := os.CreateTemp(OUTPUT, "chart*.html")
	if err != nil {
		return "", errors.New("fail to create the destination file for hodling charts")
	}
	defer dstFile.Close()

	err = plotting.GenCharts(dstFile, lines)
	if err != nil {
		return "", errors.New("fail to plot charts")
	}

	return dstFile.Name(), nil
}

func Start() error {
	r := gin.Default()
	r.Static("/static", STATIC)
	r.Static("/output", OUTPUT)
	r.Static("/tmp", TMP)

	r.GET("/", func(c *gin.Context) {
		c.File(staticFile("index.html"))
	})

	r.POST("/upload", func(c *gin.Context) {
		xTitle := getFormDefault(c, "xaxis", "x")
		yTitle := getFormDefault(c, "yaxis", "y")
		c1x := getBool(getFormDefault(c, "c1x", "true"))
		r1h := getBool(getFormDefault(c, "r1h", "true"))
		smooth := getBool(getFormDefault(c, "smooth", "false"))

		mform, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		files := mform.File["files"]

		var total int64
		for _, srcFile := range files {
			total += srcFile.Size
		}
		if total > MAXSIZE {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file size exceeds the maximum limit of 20MB"})
			return
		}

		var hm []handlerMapper
		for _, srcFile := range files {
			dstFile, err := os.CreateTemp(TMP, "uploaded-*.txt")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("fail to create the destination file for storing %s", srcFile.Filename)})
				return
			}
			defer dstFile.Close()
			hm = append(hm, handlerMapper{
				src:     srcFile,
				srcName: filepath.Base(srcFile.Filename),
				dst:     dstFile,
				dstName: dstFile.Name(),
			})
		}

		for _, v := range hm {
			if err := c.SaveUploadedFile(v.src, v.dstName); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("fail to upload file %s due to %s", v.srcName, err.Error())})
				return
			}
			v.dst.Close()
		}

		var dataf, origName []string
		for _, v := range hm {
			origName = append(origName, v.srcName)
			dataf = append(dataf, v.dstName)
		}

		output, err := plotCharts(dataf, origName, xTitle, yTitle, c1x, r1h, smooth)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("fail to plot chart due to %s", err.Error())})
			return
		}
		_, fname := path.Split(output)
		c.Redirect(http.StatusFound, fmt.Sprintf("/output/%s", fname))
	})

	return r.Run(":8080")
}

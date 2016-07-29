package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
	"github.com/wcharczuk/go-web"
)

func chartHandler(rc *web.RequestContext) web.ControllerResult {

	rc.Response.Header().Set("Content-Type", "image/png")
	k := rc.Request.FormValue("K")
	startDate := rc.Request.FormValue("start")
	stopDate := rc.Request.FormValue("stop")

	startDt, _ := time.Parse("2006-01-02 15:04", startDate)
	stopDt, _ := time.Parse("2006-01-02 15:04", stopDate)

	log.Println("Requesting :", k, "start", startDt.Unix(), "stop", stopDt.Unix())

	startTS := strconv.FormatInt(startDt.Unix(), 10)
	stopTS := strconv.FormatInt(stopDt.Unix(), 10)

	resp, err := http.Get("http://localhost:7777/get?K=" + k + "&StartTS=" + startTS + "&StopTS=" + stopTS)
	if err != nil {

		log.Println("Error in http request")
		time.Sleep(time.Hour)
		return nil
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	var rmap map[string]interface{}

	err = json.Unmarshal(body, &rmap)

	if err != nil {

		log.Println(err)
		log.Println(string(body))
		return nil
	}

	var s1x []time.Time
	var s1y []float64

	keys := make([]string, len(rmap))
	i := 0
	for k := range rmap {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	for _, k := range keys {

		log.Println(k, " -> ", rmap[k].(float64))

		i, err := strconv.ParseInt(k, 10, 64)
		if err != nil {
			panic(err)
		}
		tm := time.Unix(i, 0)

		s1x = append(s1x, tm)

		s1y = append(s1y, rmap[k].(float64))
	}

	s1 := chart.TimeSeries{
		Name:    k,
		XValues: s1x,
		YValues: s1y,
		Style: chart.Style{
			Show:        true,
			FillColor:   drawing.Color{R: 150, G: 0, B: 40, A: 255},
			StrokeColor: drawing.Color{R: 255, G: 102, B: 75, A: 255},
			FontColor:   drawing.Color{R: 255, G: 255, B: 255, A: 255},
		},
	}

	c := chart.Chart{
		Title: k,
		TitleStyle: chart.Style{
			Show: false,
		},
		Canvas: chart.Style{

			FillColor:   drawing.Color{R: 0, G: 0, B: 0, A: 255},
			StrokeColor: drawing.Color{R: 0, G: 0, B: 0, A: 255},
			FontColor:   drawing.Color{R: 255, G: 255, B: 255, A: 255},
			StrokeWidth: chart.DefaultStrokeWidth,
		},

		Width:  1024,
		Height: 400,
		XAxis: chart.XAxis{
			Style: chart.Style{
				Show:      true,
				FontColor: drawing.Color{R: 255, G: 102, B: 75, A: 255},
			},
			ValueFormatter: chart.TimeHourValueFormatter,
		},
		YAxis: chart.YAxis{
			Style: chart.Style{
				Show:      true,
				FontColor: drawing.Color{R: 255, G: 102, B: 75, A: 255},
			},
			Zero: chart.GridLine{
				Style: chart.Style{
					Show:        true,
					StrokeWidth: 1.0,
				},
			},
			GridMajorStyle: chart.Style{
				Show: false,
			},
			GridMinorStyle: chart.Style{
				Show: true,
			},
		},
		Series: []chart.Series{
			s1,
		},
	}

	err = c.Render(chart.PNG, rc.Response)
	if err != nil {
		return rc.API().InternalError(err)
	}
	rc.Response.WriteHeader(http.StatusOK)
	return nil
}

func main() {
	app := web.New()
	app.SetName("Axihome chart view")
	app.SetLogger(web.NewStandardOutputLogger())
	app.GET("/", chartHandler)
	app.GET("/favico.ico", func(rc *web.RequestContext) web.ControllerResult {
		return rc.Raw([]byte{})
	})
	log.Fatal(app.Start())
}

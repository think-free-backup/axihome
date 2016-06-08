package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/think-free/jsonrpc/client"
	"github.com/think-free/jsonrpc/common/messages"
	"github.com/think-free/jsonrpc/common/tools"
	"github.com/think-free/log"
)

const (
	SERVERIP     = "localhost"
	SERVERPORT   = "3332"
	SERVERNAME   = "axihome"
	POOLINTERVAL = 5
)

type Weather struct {
	Base   string `json:"base"`
	Clouds struct {
		All int `json:"all"`
	} `json:"clouds"`
	Cod   int `json:"cod"`
	Coord struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"coord"`
	Dt   int `json:"dt"`
	ID   int `json:"id"`
	Main struct {
		Humidity int `json:"humidity"`
		Pressure int `json:"pressure"`
		Temp     int `json:"temp"`
		TempMax  int `json:"temp_max"`
		TempMin  int `json:"temp_min"`
	} `json:"main"`
	Name string `json:"name"`
	Sys  struct {
		Country string  `json:"country"`
		ID      int     `json:"id"`
		Message float64 `json:"message"`
		Sunrise int     `json:"sunrise"`
		Sunset  int     `json:"sunset"`
		Type    int     `json:"type"`
	} `json:"sys"`
	Weather []struct {
		Description string `json:"description"`
		Icon        string `json:"icon"`
		ID          int    `json:"id"`
		Main        string `json:"main"`
	} `json:"weather"`
	Wind struct {
		Deg   int     `json:"deg"`
		Speed float64 `json:"speed"`
	} `json:"wind"`
}

func main() {

	flag.Parse()
	instance := flag.Arg(0)

	log.SetProcess(instance)

	sendChannel := make(chan []byte)
	stateChannel := make(chan *jsonrpcmessage.StateBody)
	rpcMessageChannel := make(chan *jsonrpcmessage.RpcMessage)

	client := jsonrpclient.New(instance, SERVERIP, SERVERPORT, sendChannel, stateChannel, rpcMessageChannel, nil, nil)
	go client.Run()

	var domain string

	for {

		select {

		case state := <-stateChannel:

			domain = state.Domain

			log.Println("New domain : " + state.Domain)

			params := make(map[string]string)
			params["bucket"] = "Instances"
			params["variable"] = flag.Arg(0)

			jsontools.GenerateRpcMessage(&sendChannel, "bucket", "getVar", params, domain, SERVERNAME)

		case message := <-rpcMessageChannel:

			if (message.Body.Module + "." + message.Body.Fct) == "bucket.setVar" {

				p := message.Body.Params.(map[string]interface{})

				if p["bucket"] == "Instances" {

					city := p["value"].(map[string]interface{})["params"].(map[string]interface{})["city"].(string)
					apikey := p["value"].(map[string]interface{})["params"].(map[string]interface{})["apikey"].(string)
					lang := p["value"].(map[string]interface{})["params"].(map[string]interface{})["lang"].(string)
					units := p["value"].(map[string]interface{})["params"].(map[string]interface{})["units"].(string)

					url := "http://api.openweathermap.org/data/2.5/weather?q=" + city + "&appid=" + apikey + "&lang=" + lang + "&units=" + units

					for {

						resp, err := http.Get(url)
						if err != nil {

							log.Println("Error in http request")
							time.Sleep(time.Hour)
							continue
						}

						body, err := ioutil.ReadAll(resp.Body)
						resp.Body.Close()
						if err != nil {

							log.Println("Error reading body")
							time.Sleep(time.Hour)
							continue
						}

						var weather Weather

						err = json.Unmarshal(body, &weather)
						if err != nil {

							log.Println("Error unmarshalling json")
							time.Sleep(time.Hour)
							continue
						}

						p := make(map[string]interface{})
						p["openweather."+instance+".main"] = weather.Weather[0].Main
						p["openweather."+instance+".id"] = float64(weather.Weather[0].ID)
						p["openweather."+instance+".desc"] = weather.Weather[0].Description
						p["openweather."+instance+".icon"] = weather.Weather[0].Icon
						p["openweather."+instance+".iconurl"] = "http://openweathermap.org/img/w/" + weather.Weather[0].Icon + ".png"
						p["openweather."+instance+".temp"] = float64(weather.Main.Temp)
						p["openweather."+instance+".pressure"] = float64(weather.Main.Pressure)
						p["openweather."+instance+".humidity"] = float64(weather.Main.Humidity)
						p["openweather."+instance+".wind.speed"] = weather.Wind.Speed
						p["openweather."+instance+".wind.deg"] = float64(weather.Wind.Deg)
						p["openweather."+instance+".sunrise"] = float64(weather.Sys.Sunrise)
						p["openweather."+instance+".sunset"] = float64(weather.Sys.Sunset)

						jsontools.GenerateRpcMessage(&sendChannel, "variables", "set", p, instance+".backend.axihome", "axihome")

						time.Sleep(POOLINTERVAL * time.Minute)
					}
				}

			} else {

				js, _ := json.Marshal(message.Body.Params)
				log.Println("Unknow request received : " + message.Body.Module + "." + message.Body.Fct + " " + string(js))
			}
		}
	}
}

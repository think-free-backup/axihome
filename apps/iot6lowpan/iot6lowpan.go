package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	"github.com/think-free/axihome/apps/iot6lowpan/sensors/ds18b20"
	"github.com/think-free/axihome/apps/iot6lowpan/sensors/gateway"
	"github.com/think-free/axihome/apps/iot6lowpan/sensors/si1145"
	"github.com/think-free/jsonrpc/client"
	"github.com/think-free/jsonrpc/common/tools"
)

type Sensor struct {
	Run    bool
	Params interface{}
}

func main() {

	// Variables
	cmutex := &sync.Mutex{}
	cache := make(map[string]interface{})
	c := make(chan map[string]interface{})

	var sendChannel chan []byte
	name := ""

	// Read config

	content, e := ioutil.ReadFile("config.json")
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}

	/* Load config */

	var config map[string]Sensor
	json.Unmarshal(content, &config)

	// Sensors
	if config["si1145"].Run {
		go si1145.Run(&c, config["si1145"].Params)
	}

	if config["ds18b20"].Run {
		go ds18b20.Run(&c, config["ds18b20"].Params)
	}

	if config["gateway"].Run {
		go gateway.Run(&c, config["gateway"].Params)
	}

	if config["client"].Run {

		sendChannel = make(chan []byte)

		cf := config["client"].Params.(map[string]interface{})
		name = cf["name"].(string)
		client := jsonrpclient.New(name, cf["server"].(string), "3332", sendChannel, nil, nil, nil, nil)
		go client.Run()
	}

	// Web server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		js, _ := json.Marshal(cache)
		w.Write(js)
	})

	go http.ListenAndServe(":6103", nil)

	// Get values
	for {

		select {
		case val := <-c:

			for k, v := range val {

				cmutex.Lock()
				cache[k] = v
				cmutex.Unlock()
			}

			if sendChannel != nil {

				cmutex.Lock()
				jsontools.GenerateRpcMessage(&sendChannel, "variables", "set", val, name+".core.axihome", "axihome")
				cmutex.Unlock()
			}
		}
	}
}

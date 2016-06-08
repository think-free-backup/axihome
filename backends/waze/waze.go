package main

import (
	"flag"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"strings"
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
			params["variable"] = instance

			jsontools.GenerateRpcMessage(&sendChannel, "bucket", "getVar", params, domain, SERVERNAME)

		case message := <-rpcMessageChannel:

			if (message.Body.Module + "." + message.Body.Fct) == "bucket.setVar" {

				p := message.Body.Params.(map[string]interface{})

				if p["bucket"] == "Instances" {

					fromX := p["value"].(map[string]interface{})["params"].(map[string]interface{})["fromX"].(string)
					fromY := p["value"].(map[string]interface{})["params"].(map[string]interface{})["fromY"].(string)
					toX := p["value"].(map[string]interface{})["params"].(map[string]interface{})["toX"].(string)
					toY := p["value"].(map[string]interface{})["params"].(map[string]interface{})["toY"].(string)

					url := "https://www.waze.com/row-RoutingManager/routingRequest?from=x%3A" + fromX + "+y%3A" + fromY + "&to=x%3A" + toX + "+y%3A" + toY + "&at=0&returnGeometries=fasle&returnInstructions=false&timeout=60000&nPaths=1&clientVersion=4.0.0&options=AVOID_TRAILS%3At%2CALLOW_UTURNS&returnXML=true"

					go func() {

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

							slc := strings.Split(string(body), "time=\"")
							if len(slc) < 2 {

								log.Println("Error 1 parsing answer from : " + url)
								log.Println(string(body))
								time.Sleep(time.Hour)
								continue
							}
							if len(slc) == 0 {

								log.Println("Error 2 parsing answer from : " + url)
								log.Println(string(body))
								time.Sleep(time.Hour)
								continue
							}
							slc2 := strings.Split(slc[1], "\"")
							strVal := slc2[0]
							flt, err := strconv.ParseFloat(strVal, 64)
							if err != nil {

								log.Println(err.Error())
								time.Sleep(time.Hour)
								continue
							}

							tim := math.Floor((flt/60)*100) / 100

							par := make(map[string]interface{})
							par["waze.time."+instance] = tim

							jsontools.GenerateRpcMessage(&sendChannel, "variables", "set", par, instance+".backend.axihome", "axihome")

							time.Sleep(POOLINTERVAL * time.Minute)
						}
					}()
				}

			}
		}
	}
}

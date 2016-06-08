package main

import (
	"flag"
	"math"

	"github.com/think-free/jsonrpc/client"
	"github.com/think-free/jsonrpc/common/messages"
	"github.com/think-free/jsonrpc/common/tools"
	"github.com/think-free/log"
)

const (
	SERVERIP   = "localhost"
	SERVERPORT = "3330"
	SERVERNAME = "axihome"
)

/* Main app */

func main() {

	flag.Parse()
	instance := flag.Arg(0)

	log.SetProcess(instance)

	// Axihome communication

	sendChannel := make(chan []byte)
	stateChannel := make(chan *jsonrpcmessage.StateBody)
	rpcMessageChannel := make(chan *jsonrpcmessage.RpcMessage)

	client := jsonrpclient.New(instance, SERVERIP, SERVERPORT, sendChannel, stateChannel, rpcMessageChannel, nil, nil)

	// Config

	var domain string
	watchlst := make(map[string]interface{})
	values := make(map[string]float64)
	serverLastTime := 0.0

	// Starting axihome communication

	go client.Run()

	for {

		select {

		case state := <-stateChannel:

			domain = state.Domain

			log.Println("New domain : " + domain)

			if state.Tld {

				jsontools.GenerateRpcMessage(&sendChannel, "bucket", "getAll", "VariablesTimer", instance+".core.axihome", "axihome")
			}

		case message := <-rpcMessageChannel:

			modfct := message.Body.Module + "." + message.Body.Fct

			if modfct == "bucket.setAll" {

				params := message.Body.Params.(map[string]interface{})

				if params["bucket"] == "VariablesTimer" {

					log.Println("Received variable timer list")

					wlst := params["content"].(map[string]interface{})["watchlist"].([]interface{})

					for _, v := range wlst {

						watchlst[v.(string)] = nil

						log.Println("Watching : " + v.(string))
					}

					jsontools.GenerateRpcMessage(&sendChannel, "variables", "getAll", "", instance+".core.axihome", "axihome")
				}
			} else if modfct == "variables.set" {

				params := message.Body.Params.(map[string]interface{})

				for k, v := range params {

					if k == "server.time" {

						diff := v.(float64) - serverLastTime

						if diff != v.(float64) {

							for timerk, timerv := range values {

								if timerv > 0.0 {

									p := make(map[string]interface{})
									p[timerk] = math.Max(timerv-diff, 0.0)
									jsontools.GenerateRpcMessage(&sendChannel, "variables", "write", p, instance+".core.axihome", "axihome")
								}
							}
						}

						serverLastTime = v.(float64)

					} else if _, ok := watchlst[k]; ok {

						values[k] = v.(float64)
					}
				}
			}
		}
	}
}

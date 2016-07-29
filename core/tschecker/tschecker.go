package main

import (
	"flag"
	"time"

	"github.com/think-free/jsonrpc/client"
	"github.com/think-free/jsonrpc/common/messages"
	"github.com/think-free/jsonrpc/common/tools"
	"github.com/think-free/log"
)

const (
	SERVERIP       = "localhost"
	SERVERPORT     = "3330"
	SERVERNAME     = "axihome"
	MYCONFIGBUCKET = "TsChecker"
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

	tickChan := time.NewTicker(time.Second * 60).C

	watchlst := make(map[string]interface{})
	values := make(map[string]float64)

	for {

		select {

		case <-tickChan:

			now := float64(time.Now().Unix())

			p := make(map[string]interface{})

			for k, v := range values {

				if v >= now+30 || v <= now-60 {

					p[k+".alarm"] = true

				} else {

					p[k+".alarm"] = false
				}
			}

			jsontools.GenerateRpcMessage(&sendChannel, "variables", "write", p, instance+".core.axihome", "axihome")

		case state := <-stateChannel:

			domain = state.Domain

			log.Println("New domain : " + state.Domain)

			if state.Tld {

				// Core

				jsontools.GenerateRpcMessage(&sendChannel, "bucket", "getAll", MYCONFIGBUCKET, domain, SERVERNAME)
			}

		case message := <-rpcMessageChannel:

			modfct := message.Body.Module + "." + message.Body.Fct

			if modfct == "bucket.setAll" { // Core

				params := message.Body.Params.(map[string]interface{})

				if params["bucket"] == MYCONFIGBUCKET {

					wlst := params["content"].(map[string]interface{})["variables"].([]interface{})

					for _, v := range wlst {

						watchlst[v.(string)] = 0

						log.Println("Watching : " + v.(string))
					}
				}

			} else if modfct == "variables.set" {

				params := message.Body.Params.(map[string]interface{})

				for k, v := range params {

					if _, ok := watchlst[k]; ok {

						values[k] = v.(float64)
					}
				}
			}
		}
	}
}

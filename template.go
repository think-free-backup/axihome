package main

import (
	"encoding/json"
	"flag"
	"os/exec"
	"strings"
	"time"

	"github.com/think-free/jsonrpc/client"
	"github.com/think-free/jsonrpc/common/messages"
	"github.com/think-free/jsonrpc/common/tools"
	"github.com/think-free/log"
)

const (
	SERVERIP   = "localhost"
	SERVERPORT = "3332" // -> backend
	// SERVERPORT = "3330" // -> core
	SERVERNAME     = "axihome"
	MYCONFIGBUCKET = "" // -> core
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

			if state.Tld {

				// Backend

				params := make(map[string]string)
				params["bucket"] = "Instances"
				params["variable"] = flag.Arg(0)
				jsontools.GenerateRpcMessage(&sendChannel, "bucket", "getVar", params, domain, SERVERNAME)

				// Core

				jsontools.GenerateRpcMessage(&sendChannel, "bucket", "getAll", MYCONFIGBUCKET, domain, SERVERNAME)
			}

		case message := <-rpcMessageChannel:

			if (message.Body.Module + "." + message.Body.Fct) == "bucket.setVar" { // Backend

				p := message.Body.Params.(map[string]interface{})

				if p["bucket"] == "Instances" {

					instanceParam1 := p["value"].(map[string]interface{})["params"].(map[string]interface{})["instanceParam1"].(string)
				}

			} else if modfct == "bucket.setAll" { // Core

				params := message.Body.Params.(map[string]interface{})

				if params["bucket"] == MYCONFIGBUCKET {

					content := params["content"].(map[string]interface{})
				}

			} else {

				js, _ := json.Marshal(message.Body.Params)
				log.Println("Unknow request received : " + message.Body.Module + "." + message.Body.Fct + " " + string(js))
			}
		}
	}
}

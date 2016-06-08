package main

import (
	"encoding/json"
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/think-free/jsonrpc/client"
	"github.com/think-free/jsonrpc/common/messages"
	"github.com/think-free/log"
)

func main() {

	log.SetProcess("v2bridge")

	flag.Parse()

	var v3domain string
	var mutex *sync.Mutex = &sync.Mutex{}

	// Create the v2 client
	v2sendChannel := make(chan []byte)
	v2stateChannel := make(chan *jsonrpcmessage.StateBody)
	v2rpcMessageChannel := make(chan *jsonrpcmessage.RpcMessage)

	v2client := jsonrpclient.New("v2bridge", "localhost", "2015", v2sendChannel, v2stateChannel, v2rpcMessageChannel, nil, nil)
	go v2client.Run()

	// Create the v3 client
	v3sendChannel := make(chan []byte)
	v3stateChannel := make(chan *jsonrpcmessage.StateBody)
	v3rpcMessageChannel := make(chan *jsonrpcmessage.RpcMessage)

	v3client := jsonrpclient.New("v2bridge", "localhost", "3332", v3sendChannel, v3stateChannel, v3rpcMessageChannel, nil, nil)
	go v3client.Run()

	// Wait for channels
	for {

		select {

		// V2 Rpc message
		case message := <-v2rpcMessageChannel:

			message.Dst = "axihome"
			message.Src = v3domain

			js, _ := json.Marshal(message)
			v3sendChannel <- js

		// V3 State
		case state := <-v3stateChannel:

			log.Println("New v3 domain : " + state.Domain)
			mutex.Lock()
			v3domain = state.Domain
			mutex.Unlock()

			body := make(map[string]interface{})
			body["module"] = "bucket"
			body["fct"] = "getVar"
			params := make(map[string]string)
			params["bucket"] = "Instances"
			params["variable"] = flag.Arg(0)
			body["params"] = params

			mes := jsonrpcmessage.NewRoutedMessage("rpc", body, v3domain, "axihome")
			json, _ := json.Marshal(mes)
			v3sendChannel <- json

		// V3 Rpc message
		case message := <-v3rpcMessageChannel:

			if (message.Body.Module + "." + message.Body.Fct) == "variables.write" {

				message.Src = "v2bridge.axihomeserver"
				message.Dst = "axihomeserver"
				message.Body.Module = "variable"
				jsonmessage, _ := json.Marshal(message)
				v2sendChannel <- jsonmessage
			} else {

				json, _ := json.Marshal(message.Body.Params)
				log.Println("Unknow request received : " + message.Body.Module + "." + message.Body.Fct + " " + string(json))
			}
		}

	}

	// Handle ctrl+c and exit signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)

	for {
		select {
		case _ = <-c:
			log.Println("\nClosing application")
			os.Exit(1)
		}
	}
}

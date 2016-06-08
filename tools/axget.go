package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/think-free/jsonrpc/client"
	"github.com/think-free/jsonrpc/common/messages"
)

func main() {

	flag.Parse()

	server := flag.Arg(0)

	if server == "-h" {

		fmt.Println("axget server port variable")
		os.Exit(0)
	}

	port := flag.Arg(1)
	variable := flag.Arg(2)

	sendChannel := make(chan []byte)
	stateChannel := make(chan *jsonrpcmessage.StateBody)
	rpcMessageChannel := make(chan *jsonrpcmessage.RpcMessage)

	client := jsonrpclient.New("axget", server, port, sendChannel, stateChannel, rpcMessageChannel, nil, nil)
	go client.Run()

	for {
		select {

		case state := <-stateChannel:

			body := make(map[string]interface{})
			params := make(map[string]interface{})
			body["module"] = "variables"
			body["fct"] = "getAll"
			params["variable"] = variable
			body["params"] = params

			mes := jsonrpcmessage.NewRoutedMessage("rpc", body, state.Domain, "axihome")
			json, _ := json.Marshal(mes)
			sendChannel <- json

		case message := <-rpcMessageChannel:

			if message.Body.Fct == "set" {

				fmt.Println(message.Body.Params.(map[string]interface{})[variable])
				os.Exit(0)
			}
		}
	}
}

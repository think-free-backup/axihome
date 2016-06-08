package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/think-free/jsonrpc/client"
	"github.com/think-free/jsonrpc/common/messages"
)

func main() {

	flag.Parse()

	server := flag.Arg(0)

	if server == "-h" {

		fmt.Println("axhistorydump server port start end")
		os.Exit(0)
	}

	port := flag.Arg(1)
	start := flag.Arg(2)
	end := flag.Arg(3)

	sendChannel := make(chan []byte)
	stateChannel := make(chan *jsonrpcmessage.StateBody)
	rpcMessageChannel := make(chan *jsonrpcmessage.RpcMessage)

	client := jsonrpclient.New("axhistorydump", server, port, sendChannel, stateChannel, rpcMessageChannel, nil, nil)
	go client.Run()

	for {
		select {

		case state := <-stateChannel:

			body := make(map[string]interface{})
			params := make(map[string]interface{})
			body["module"] = "history"
			body["fct"] = "get"
			params["start"], _ = strconv.Atoi(start)
			params["end"], _ = strconv.Atoi(end)
			body["params"] = params

			mes := jsonrpcmessage.NewRoutedMessage("rpc", body, state.Domain, "axihome")
			json, _ := json.Marshal(mes)
			sendChannel <- json

		case message := <-rpcMessageChannel:

			if message.Body.Module == "history" && message.Body.Fct == "set" {

				fmt.Println(message.Body.Params)
				os.Exit(0)
			}
		}
	}
}

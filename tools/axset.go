package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/think-free/jsonrpc/client"
	"github.com/think-free/jsonrpc/common/messages"
)

func main() {

	flag.Parse()

	server := flag.Arg(0)

	if server == "-h" {

		fmt.Println("axset server port type variable value")
		os.Exit(0)
	}

	port := flag.Arg(1)
	vtype := flag.Arg(2)
	variable := flag.Arg(3)
	value := flag.Arg(4)

	sendChannel := make(chan []byte)
	stateChannel := make(chan *jsonrpcmessage.StateBody)
	rpcMessageChannel := make(chan *jsonrpcmessage.RpcMessage)

	client := jsonrpclient.New("axset", server, port, sendChannel, stateChannel, rpcMessageChannel, nil, nil)
	go client.Run()

	for {
		select {

		case state := <-stateChannel:

			body := make(map[string]interface{})
			body["module"] = "variables"
			body["fct"] = "set"
			params := make(map[string]interface{})

			if vtype == "string" {

				params[variable] = value
			} else if vtype == "float" {

				params[variable], _ = strconv.ParseFloat(value, 64)
			} else if vtype == "bool" {

				params[variable], _ = strconv.ParseBool(value)
			}

			body["params"] = params

			mes := jsonrpcmessage.NewRoutedMessage("rpc", body, state.Domain, "axihome")
			json, _ := json.Marshal(mes)
			sendChannel <- json

			time.Sleep(3 * time.Second)
			os.Exit(0)
		}
	}
}

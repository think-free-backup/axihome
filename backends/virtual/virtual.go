package main

import (
	"flag"

	"github.com/think-free/jsonrpc/client"
	"github.com/think-free/jsonrpc/common/messages"
	"github.com/think-free/jsonrpc/common/tools"
	"github.com/think-free/log"
)

const (
	SERVERIP   = "localhost"
	SERVERPORT = "3332"
	SERVERNAME = "axihome"
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

		case message := <-rpcMessageChannel:

			if (message.Body.Module + "." + message.Body.Fct) == "variables.write" {

				jsontools.GenerateRpcMessage(&sendChannel, "variables", "set", message.Body.Params, domain, SERVERNAME)
			}
		}
	}
}

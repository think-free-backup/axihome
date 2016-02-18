package main

import (
	"flag"
	"time"

	"github.com/think-free/jsonrpc/client"
	"github.com/think-free/jsonrpc/common/messages"
	"github.com/think-free/jsonrpc/common/tools"
	"github.com/think-free/log"
)

type WearReq struct {
	Type string `json:"type"`
	Body string `json:"body"`
}

const (
	SERVERIP   = "localhost"
	SERVERPORT = "3332"
	SERVERNAME = "axihome"
	INTERVAL   = 10
)

func main() {

	flag.Parse()
	instance := flag.Arg(0)

	log.SetProcess(instance)

	sendChannel := make(chan []byte)
	stateChannel := make(chan *jsonrpcmessage.StateBody)

	client := jsonrpclient.New(instance, SERVERIP, SERVERPORT, sendChannel, stateChannel, nil, nil, nil)

	// Ticker

	connected := false
	ticker := time.NewTicker(time.Second * INTERVAL)
	go func() {
		for t := range ticker.C {

			if connected {

				p := make(map[string]interface{})
				p["server.time"] = float64(t.Unix())

				//log.Println(int(p["server.time"]))

				jsontools.GenerateRpcMessage(&sendChannel, "variables", "set", p, instance+".backend.axihome", "axihome")
			}
		}
	}()

	// Starting axihome communication

	go client.Run()

	for {

		select {

		case state := <-stateChannel:

			connected = state.Tld

			log.Println("New domain : " + state.Domain)
		}
	}
}

package main

import (
	"encoding/json"
	"flag"
	"os/exec"
	"sync"
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
	POOLINTERVAL = 10
	SENDINTERVAL = 15
)

type State struct {
	Devices map[string]interface{}
	Mutex   sync.Mutex
}

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
	states := State{Devices: make(map[string]interface{})}

	for {

		select {

		case state := <-stateChannel:

			domain = state.Domain

			log.Println("New domain : " + state.Domain)

			params := make(map[string]string)
			params["bucket"] = "Instances"
			params["variable"] = flag.Arg(0)

			jsontools.GenerateRpcMessage(&sendChannel, "bucket", "getVar", params, domain, SERVERNAME)

		case message := <-rpcMessageChannel:

			if (message.Body.Module + "." + message.Body.Fct) == "bucket.setVar" {

				p := message.Body.Params.(map[string]interface{})

				if p["bucket"] == "Instances" {

					log.Println("Got instances")

					devices := p["value"].(map[string]interface{})["params"].([]interface{})

					for _, v := range devices {

						name := v.(map[string]interface{})["name"].(string)
						ip := v.(map[string]interface{})["ip"].(string)

						go func() {

							for {

								command := exec.Command("bash", "-c", "ping -c 1 "+ip)
								command.Run()

								if command.ProcessState.Success() {

									states.Mutex.Lock()
									states.Devices[instance+"."+name] = true
									states.Mutex.Unlock()
								} else {

									states.Mutex.Lock()
									states.Devices[instance+"."+name] = false
									states.Mutex.Unlock()
								}

								time.Sleep(POOLINTERVAL * time.Second)
							}
						}()
					}

					go func() {

						for {
							states.Mutex.Lock()
							jsontools.GenerateRpcMessage(&sendChannel, "variables", "set", states.Devices, instance+".backend.axihome", "axihome")
							states.Mutex.Unlock()

							time.Sleep(SENDINTERVAL * time.Second)
						}
					}()
				}

			} else {

				js, _ := json.Marshal(message.Body.Params)
				log.Println("Unknow request received : " + message.Body.Module + "." + message.Body.Fct + " " + string(js))
			}
		}
	}
}

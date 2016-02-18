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
	SERVERIP     = "localhost"
	SERVERPORT   = "3332"
	SERVERNAME   = "axihome"
	POOLINTERVAL = 2
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

			params := make(map[string]string)
			params["bucket"] = "Instances"
			params["variable"] = flag.Arg(0)

			jsontools.GenerateRpcMessage(&sendChannel, "bucket", "getVar", params, domain, SERVERNAME)

		case message := <-rpcMessageChannel:

			if (message.Body.Module + "." + message.Body.Fct) == "bucket.setVar" {

				p := message.Body.Params.(map[string]interface{})

				if p["bucket"] == "Instances" {

					url := p["value"].(map[string]interface{})["params"].(map[string]interface{})["url"].(string)

					go func() {

						for {
							command := exec.Command("bash", "-c", "upsc "+url)
							out, _ := command.Output()

							lines := strings.Split(string(out), "\n")

							p := make(map[string]interface{})

							for _, line := range lines {

								if line != "" {
									val := strings.Split(line, ":")
									p[instance+"."+strings.TrimSpace(val[0])] = strings.TrimSpace(val[1])
								}
							}

							jsontools.GenerateRpcMessage(&sendChannel, "variables", "set", p, instance+".backend.axihome", "axihome")
							time.Sleep(POOLINTERVAL * time.Second)
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

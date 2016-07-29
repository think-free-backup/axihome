package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
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
	POOLINTERVAL = 10
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

					ip := p["value"].(map[string]interface{})["params"].(map[string]interface{})["ip"].(string)

					go func() {

						tr := &http.Transport{
							TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
						}

						client := &http.Client{Transport: tr}

						for {

							resp, err := client.Get("https://" + ip + "/status_gateways_json.php?key=pfsense&rates=1")
							if err != nil {

								log.Println("Error in http request")
								time.Sleep(time.Hour)
								continue
							}
							body, err := ioutil.ReadAll(resp.Body)
							resp.Body.Close()

							bodyjson := make(map[string]interface{})

							err = json.Unmarshal(body, &bodyjson)

							if err != nil {

								log.Println(err)
								time.Sleep(time.Hour)
							}

							p := make(map[string]interface{})

							for ifacename, iface := range bodyjson {

								ifacestate := iface.(map[string]interface{})

								for statename, state := range ifacestate {

									if state != nil {

										if statename == "inKbps" || statename == "outKbps" {

											p[instance+"."+strings.TrimPrefix(ifacename, "_")+"."+statename] = state.(float64)
										} else {
											p[instance+"."+strings.TrimPrefix(ifacename, "_")+"."+statename] = state.(string)
										}
									}
								}
							}

							jsontools.GenerateRpcMessage(&sendChannel, "variables", "set", p, domain, SERVERNAME)
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

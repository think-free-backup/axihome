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
	POOLINTERVAL = 5
)

type Light struct {
	Manufacturername string `json:"manufacturername"`
	Modelid          string `json:"modelid"`
	Name             string `json:"name"`
	State            struct {
		Alert     string    `json:"alert"`
		Bri       int       `json:"bri"`
		Colormode string    `json:"colormode"`
		Effect    string    `json:"effect"`
		Hue       int       `json:"hue"`
		On        bool      `json:"on"`
		Reachable bool      `json:"reachable"`
		Sat       int       `json:"sat"`
		Xy        []float64 `json:"xy"`
	} `json:"state"`
	Swversion string `json:"swversion"`
	Type      string `json:"type"`
	Uniqueid  string `json:"uniqueid"`
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

	hueNames := make(map[string]string)
	hueOn := make(map[string]bool)
	baseUrl := ""

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

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
					key := p["value"].(map[string]interface{})["params"].(map[string]interface{})["key"].(string)
					baseUrl = "http://" + ip + "/api/" + key + "/"

					go func() {

						for {

							resp, err := http.Get(baseUrl + "lights")
							if err != nil {

								log.Println("Error in http request")
								time.Sleep(time.Hour)
								continue
							}

							body, err := ioutil.ReadAll(resp.Body)
							resp.Body.Close()

							lights := make(map[string]Light)

							err = json.Unmarshal(body, &lights)

							if err != nil {

								log.Println(err)
								time.Sleep(time.Hour)
							}

							para := make(map[string]interface{})

							for idx, light := range lights {

								name := strings.Replace(strings.ToLower(light.Name), " ", "", -1)

								hueNames[name] = idx
								if v, ok := hueOn[name]; !ok || v != light.State.On {

									hueOn[name] = light.State.On
									para[instance+"."+name+".on"] = light.State.On
								}
							}

							if len(para) != 0 {

								jsontools.GenerateRpcMessage(&sendChannel, "variables", "set", para, domain, SERVERNAME)
							}

							time.Sleep(POOLINTERVAL * time.Second)
						}
					}()
				}

			} else if (message.Body.Module + "." + message.Body.Fct) == "variables.write" {

				vars := message.Body.Params.(map[string]interface{})

				for k, v := range vars {

					varr := strings.Split(k, ".")
					if varr[2] == "on" {

						mes := "{\"on\":false}"

						if v.(bool) {

							mes = "{\"on\":true}"
						}

						client := &http.Client{Transport: tr}
						request, _ := http.NewRequest("PUT", baseUrl+"/lights/"+hueNames[varr[1]]+"/state", strings.NewReader(mes))
						response, _ := client.Do(request)
						defer response.Body.Close()
					}
				}

			} else {

				js, _ := json.Marshal(message.Body.Params)
				log.Println("Unknow request received : " + message.Body.Module + "." + message.Body.Fct + " " + string(js))
			}
		}
	}
}

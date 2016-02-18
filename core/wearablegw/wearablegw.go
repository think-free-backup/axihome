package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"strings"

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
	SERVERPORT = "3330"
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

	var domain string
	var section map[string]interface{}

	// Handle rest request from wearable

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		r.ParseForm()

		var req WearReq

		request := r.PostForm.Get("request")
		json.Unmarshal([]byte(request), &req)
		log.Println(req.Type)
		log.Println(req.Body)

		if req.Type == "voice" {

			log.Println("Voice action reveived, checking ...")

			checkSection(domain, req.Body, section, &sendChannel)
		} else {

			log.Println("Unknow request reveived")
		}

		http.StatusText(200)
	})

	go http.ListenAndServe(":3600", nil)

	// Starting axihome communication

	go client.Run()

	for {

		select {

		case state := <-stateChannel:

			domain = state.Domain

			log.Println("New domain : " + state.Domain)

			if state.Tld {

				jsontools.GenerateRpcMessage(&sendChannel, "bucket", "getAll", "WearableVoiceAction", "wearablegw.core.axihome", "axihome")
			}

		case message := <-rpcMessageChannel:

			modfct := message.Body.Module + "." + message.Body.Fct

			if modfct == "bucket.setAll" {

				params := message.Body.Params.(map[string]interface{})

				if params["bucket"] == "WearableVoiceAction" {

					log.Println("Received voice action list")

					section = params["content"].(map[string]interface{})
				}
			}
		}
	}
}

func checkSection(domain, message string, section map[string]interface{}, sendChannel *chan []byte) {

	log.Println("Checking section")

	for k, v := range section {

		action := v.(map[string]interface{})

		if strings.Contains(message, k) {

			log.Println("Section detected : " + k)
			checkDevice(domain, message, action, sendChannel)

			return
		}
	}

	log.Println("Section not found, aborting")
}

func checkDevice(domain, message string, action map[string]interface{}, sendChannel *chan []byte) {

	log.Println("Checking device")

	for k, v := range action {

		device := v.(map[string]interface{})

		if strings.Contains(message, k) {

			log.Println("Device detected : " + k)
			checkAction(domain, message, device, sendChannel)

			return
		}
	}

	log.Println("Device not found, aborting")
}

func checkAction(domain, message string, device map[string]interface{}, sendChannel *chan []byte) {

	log.Println("Checking action")

	for k, v := range device {

		values := v.(map[string]interface{})

		if strings.Contains(message, k) {

			log.Println("Action detected : " + k)

			log.Println(values["k"])

			params := make(map[string]interface{})
			params[values["k"].(string)] = values["v"]

			jsontools.GenerateRpcMessage(sendChannel, "variables", "write", params, domain, SERVERNAME)

			return
		}
	}

	log.Println("Action not found, aborting")
}

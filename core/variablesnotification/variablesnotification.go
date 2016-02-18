package main

import (
	"encoding/json"
	"flag"

	"github.com/think-free/jsonrpc/client"
	"github.com/think-free/jsonrpc/common/messages"
	"github.com/think-free/jsonrpc/common/tools"
	"github.com/think-free/log"
)

type Notification struct {
	Cond  string      `json:"cond"`
	Val   interface{} `json:"val"`
	Notif string      `json:"notif"`
}

func main() {

	flag.Parse()
	instance := flag.Arg(0)

	log.SetProcess(instance)

	var notifications map[string][]Notification

	// Create the v2 client
	sendChannel := make(chan []byte)
	stateChannel := make(chan *jsonrpcmessage.StateBody)
	rpcMessageChannel := make(chan *jsonrpcmessage.RpcMessage)

	client := jsonrpclient.New(instance, "localhost", "3330", sendChannel, stateChannel, rpcMessageChannel, nil, nil)
	go client.Run()

	for {

		select {

		// V3 State
		case state := <-stateChannel:

			if state.Tld {

				jsontools.GenerateRpcMessage(&sendChannel, "bucket", "getAll", "VariablesNotification", instance+".core.axihome", "axihome")
			}

		// Rpc message
		case message := <-rpcMessageChannel:

			modfct := message.Body.Module + "." + message.Body.Fct

			if modfct == "variables.set" {

				vars := message.Body.Params.(map[string]interface{})

				for k, v := range vars {

					if notifications[k] != nil {

						check(instance, &sendChannel, v, notifications[k])
					}
				}

			} else if modfct == "bucket.setAll" {

				params := message.Body.Params.(map[string]interface{})

				if params["bucket"] == "VariablesNotification" {

					vnotif := params["content"].(map[string]interface{})

					notifications = make(map[string][]Notification)

					for k, v := range vnotif {

						js, _ := json.Marshal(v)
						var val []Notification
						json.Unmarshal(js, &val)

						notifications[k] = val
					}
				}
			}
		}
	}
}

func check(instance string, sendChannel *chan []byte, val interface{}, notifAr []Notification) {

	for _, v := range notifAr {

		switch v.Cond {

		case ">":

			if val.(float64) > v.Val.(float64) {

				send(instance, sendChannel, v.Notif)
			}

		case ">=":

			if val.(float64) >= v.Val.(float64) {

				send(instance, sendChannel, v.Notif)
			}

		case "<=":

			if val.(float64) <= v.Val.(float64) {

				send(instance, sendChannel, v.Notif)
			}

		case "<":

			if val.(float64) < v.Val.(float64) {

				send(instance, sendChannel, v.Notif)
			}

		case "=":

			if val == v.Val {

				send(instance, sendChannel, v.Notif)
			}

		case "!=":

			if val != v.Val {

				send(instance, sendChannel, v.Notif)
			}
		}
	}
}

func send(instance string, sendChannel *chan []byte, message string) {

	jsontools.GenerateRpcMessage(sendChannel, "notifier", "send", message, instance+".core.axihome", "notification.core.axihome")
}

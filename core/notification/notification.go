// Package main
package main

import (
	"encoding/json"

	"github.com/think-free/axihome/core/notification/say"
	"github.com/think-free/axihome/core/notification/telegram"
	"github.com/think-free/jsonrpc/client"
	"github.com/think-free/jsonrpc/common/messages"
	"github.com/think-free/jsonrpc/common/tools"
	"github.com/think-free/log"
)

type Notifier interface {
	Send(string, string)
}

type Message struct {
	ShortText  string `json:"shortText"`
	MediumText string `json:"mediumText"`
	LargeText  string `json:"largeText"`
	Sound      string `json:"sound"`
	Image      string `json:"image"`
	Url        string `json:"url"`
}

// Subscriptions are map[string] []Subscription

type Subscription struct {
	Dev    string `json:"dev"`
	Type   string `json:"type"`
	Active string `json:"active"`
}

// Devices are map[string]Device

type Device struct {
	Notifier string `json:"notifier"`
	Desc     string `json:"desc"`
	Url      string `json:"url"`
}

func main() {

	log.SetProcess("notification")

	notifiers := make(map[string]Notifier)
	var messages map[string]Message
	var subscriptions map[string][]Subscription
	var devices map[string]Device

	var active map[string]interface{}

	sendChannel := make(chan []byte)
	stateChannel := make(chan *jsonrpcmessage.StateBody)
	rpcMessageChannel := make(chan *jsonrpcmessage.RpcMessage)

	v2client := jsonrpclient.New("notification", "localhost", "3330", sendChannel, stateChannel, rpcMessageChannel, nil, nil)
	go v2client.Run()

	for {

		select {

		// Rpc message
		case message := <-rpcMessageChannel:

			modfct := message.Body.Module + "." + message.Body.Fct

			if modfct == "variables.set" {

				vars := message.Body.Params.(map[string]interface{})

				for k, v := range vars {

					if _, ok := active[k]; ok {

						active[k] = v
					}
				}

			} else if modfct == "notifier.send" {

				id := message.Body.Params.(string)

				message := messages[id]

				subscriptList := subscriptions[id]

				for _, v := range subscriptList {

					if v.Active != "" {

						if active[v.Active] == true {

							typeOfMess := v.Type
							dev := devices[v.Dev]

							var text string
							if typeOfMess == "shortText" {
								text = message.ShortText
							} else if typeOfMess == "mediumText" {
								text = message.MediumText
							} else if typeOfMess == "largeText" {
								text = message.LargeText
							}

							notifiers[dev.Notifier].Send(dev.Url, text)
						}
					}
				}

			} else if modfct == "bucket.setAll" {

				params := message.Body.Params.(map[string]interface{})

				if params["bucket"] == "NotificationGeneral" {

					instances := params["content"].(map[string]interface{})

					for k, v := range instances {

						if notifiers[k] == nil {

							if k == "telegram" {

								notifiers[k] = telegram.New(v.(string))

							} else if k == "say" {

								notifiers[k] = say.New()
							}
						}
					}

				} else if params["bucket"] == "NotificationMessages" {

					messages = make(map[string]Message)
					messagesiface := params["content"].(map[string]interface{})

					for k, v := range messagesiface {

						js, _ := json.Marshal(v)
						var val Message
						json.Unmarshal(js, &val)

						messages[k] = val
					}

				} else if params["bucket"] == "NotificationMessagesSubscriptions" {

					active = make(map[string]interface{})

					subscriptions = make(map[string][]Subscription)
					subscriptionsiface := params["content"].(map[string]interface{})

					for k, v := range subscriptionsiface {

						js, _ := json.Marshal(v)
						var val []Subscription
						json.Unmarshal(js, &val)

						for _, vVal := range val {

							if vVal.Active != "" {

								active[vVal.Active] = false
							}
						}

						subscriptions[k] = val
					}

					jsontools.GenerateRpcMessage(&sendChannel, "variables", "getAll", "", "notification.core.axihome", "axihome")

				} else if params["bucket"] == "NotificationDevices" {

					devices = make(map[string]Device)
					devicesiface := params["content"].(map[string]interface{})

					for k, v := range devicesiface {

						js, _ := json.Marshal(v)
						var val Device
						json.Unmarshal(js, &val)

						devices[k] = val
					}
				}
			}

		case state := <-stateChannel:

			if state.Tld {

				jsontools.GenerateRpcMessage(&sendChannel, "bucket", "getAll", "NotificationGeneral", "notification.core.axihome", "axihome")
				jsontools.GenerateRpcMessage(&sendChannel, "bucket", "getAll", "NotificationMessages", "notification.core.axihome", "axihome")
				jsontools.GenerateRpcMessage(&sendChannel, "bucket", "getAll", "NotificationMessagesSubscriptions", "notification.core.axihome", "axihome")
				jsontools.GenerateRpcMessage(&sendChannel, "bucket", "getAll", "NotificationDevices", "notification.core.axihome", "axihome")
			}
		}
	}
}

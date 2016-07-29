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
	SERVERPORT = "3330"
	SERVERNAME = "axihome"
)

type Binding struct {
	cond    string
	val     string
	actions []Action
}

type Action struct {
	active   string
	variable string
	value    interface{}
}

func main() {

	flag.Parse()
	instance := flag.Arg(0)

	log.SetProcess(instance)

	sendChannel := make(chan []byte)
	stateChannel := make(chan *jsonrpcmessage.StateBody)
	rpcMessageChannel := make(chan *jsonrpcmessage.RpcMessage)

	client := jsonrpclient.New(instance, SERVERIP, SERVERPORT, sendChannel, stateChannel, rpcMessageChannel, nil, nil)

	// Starting axihome communication

	go client.Run()

	initialized := false
	keyvaluemap := make(map[string]interface{})
	bindings := make(map[string][]Binding)

	for {

		select {

		case state := <-stateChannel:

			log.Println("New domain : " + state.Domain)

			if state.Tld {

				jsontools.GenerateRpcMessage(&sendChannel, "variables", "getAll", "", "variablebinding.core.axihome", "axihome")
				jsontools.GenerateRpcMessage(&sendChannel, "bucket", "getAll", "VariablesBinding", "variablebinding.core.axihome", "axihome")
			}

		case message := <-rpcMessageChannel:

			modfct := message.Body.Module + "." + message.Body.Fct

			if modfct == "bucket.setAll" {

				params := message.Body.Params.(map[string]interface{})

				if params["bucket"] == "VariablesBinding" {

					log.Println("Received binding list")

					variables := params["content"].(map[string]interface{})

					for variable, bind := range variables {

						log.Println("Processing binding info for : " + variable)

						bindarray := bind.([]interface{})

						bindings[variable] = make([]Binding, len(bindarray))

						for bindidx, bind := range bindarray {

							bindmap := bind.(map[string]interface{})

							actarray := bindmap["actions"].([]interface{})

							actions := make([]Action, len(actarray))

							for actidx, act := range actarray {

								actionmap := act.(map[string]interface{})

								action := Action{active: actionmap["active"].(string), variable: actionmap["var"].(string), value: actionmap["value"]}

								actions[actidx] = action
							}

							binding := Binding{cond: bindmap["cond"].(string), val: bindmap["val"].(string), actions: actions}

							bindings[variable][bindidx] = binding
						}
					}

					initialized = true
				}

			} else if modfct == "variables.set" {

				params := message.Body.Params.(map[string]interface{})

				for k, v := range params {

					keyvaluemap[k] = v

					if initialized {

						if arrbind, ok := bindings[k]; ok {

							// Looping throught binding actions

							for _, bind := range arrbind {

								val := keyvaluemap[bind.val]

								if val == nil {

									log.Println("Can't find value to compare for " + bind.val)
									continue
								}

								sendVal := false

								switch bind.cond {

								case ">":

									if v.(float64) > val.(float64) {
										sendVal = true
									}
								case ">=":

									if v.(float64) >= val.(float64) {
										sendVal = true
									}
								case "<=":

									if v.(float64) <= val.(float64) {
										sendVal = true
									}
								case "<":

									if v.(float64) < val.(float64) {
										sendVal = true
									}
								case "=":

									if v == val {
										sendVal = true
									}
								case "!=":

									if v != val {
										sendVal = true
									}
								}

								if sendVal {

									for _, action := range bind.actions {

										if active, ok := keyvaluemap[action.active]; ok && active.(bool) {

											p := make(map[string]interface{})
											p[action.variable] = action.value

											jsontools.GenerateRpcMessage(&sendChannel, "variables", "write", p, "variablebinding.core.axihome", "axihome")
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

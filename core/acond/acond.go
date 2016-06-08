package main

import (
	"encoding/json"
	"flag"
	"math"

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

type Acond struct {
	Vars []string `json:"vars"`
	Fct  string   `json:"fct"`
	Var  string   `json:"var"`
}

type Config struct {
	acond  map[string]Acond
	lookup map[string][]string
	vars   map[string]interface{}
}

/* Main app */

func main() {

	fct := make(map[string]func(string, string, Config) (interface{}, bool))
	fct["avg"] = avg
	fct["max"] = max
	fct["min"] = min
	fct["sum"] = sum
	fct["and"] = and
	fct["or"] = or
	fct["lastactive"] = lastactive
	fct["diff"] = diff

	flag.Parse()
	instance := flag.Arg(0)

	log.SetProcess(instance)

	// Axihome communication

	sendChannel := make(chan []byte)
	stateChannel := make(chan *jsonrpcmessage.StateBody)
	rpcMessageChannel := make(chan *jsonrpcmessage.RpcMessage)

	client := jsonrpclient.New(instance, SERVERIP, SERVERPORT, sendChannel, stateChannel, rpcMessageChannel, nil, nil)

	// Config

	var domain string
	config := Config{}
	config.acond = make(map[string]Acond)
	config.lookup = make(map[string][]string)
	config.vars = make(map[string]interface{})

	// Starting axihome communication

	go client.Run()

	for {

		select {

		case state := <-stateChannel:

			domain = state.Domain

			log.Println("New domain : " + domain)

			if state.Tld {

				jsontools.GenerateRpcMessage(&sendChannel, "bucket", "getAll", "Acond", instance+".core.axihome", "axihome")
				jsontools.GenerateRpcMessage(&sendChannel, "variables", "getAll", "", instance+".core.axihome", "axihome")
			}

		case message := <-rpcMessageChannel:

			modfct := message.Body.Module + "." + message.Body.Fct

			if modfct == "bucket.setAll" {

				params := message.Body.Params.(map[string]interface{})

				if params["bucket"] == "Acond" {

					log.Println("Received acond list")

					acondlst := params["content"].(map[string]interface{})

					for k, v := range acondlst {

						js, _ := json.Marshal(v)
						var ac Acond
						json.Unmarshal(js, &ac)

						config.acond[k] = ac

						for _, va := range ac.Vars {

							config.lookup[va] = append(config.lookup[va], k)
						}
					}
				}
			} else if modfct == "variables.set" {

				params := message.Body.Params.(map[string]interface{})

				for triggerVar, value := range params {

					if look, ok := config.lookup[triggerVar]; ok {

						config.vars[triggerVar] = value

						for _, currentAcond := range look {

							if v, ok := fct[config.acond[currentAcond].Fct](triggerVar, currentAcond, config); ok {

								p := make(map[string]interface{})
								p[config.acond[currentAcond].Var] = v
								jsontools.GenerateRpcMessage(&sendChannel, "variables", "set", p, instance+".core.axihome", "axihome")
							}
						}
					}
				}
			}
		}
	}
}

/* Helper */

func round(val float64) float64 {

	return float64(int(val*100)) / 100
}

/* Acond functions */

func avg(triggerVar, currentAcond string, config Config) (interface{}, bool) {

	vlist := config.acond[currentAcond].Vars

	sum := 0.0
	count := 0.0

	for _, v := range vlist {

		if val, ok := config.vars[v]; ok {

			count++
			sum = sum + val.(float64)
		}
	}

	result := round(sum / count)

	return result, true
}

func max(triggerVar, currentAcond string, config Config) (interface{}, bool) {

	vlist := config.acond[currentAcond].Vars

	max := -100000.0

	for _, v := range vlist {

		if val, ok := config.vars[v]; ok {

			if val.(float64) > max {

				max = val.(float64)
			}
		}
	}

	result := round(max)

	return result, true
}

func min(triggerVar, currentAcond string, config Config) (interface{}, bool) {

	vlist := config.acond[currentAcond].Vars

	min := 100000.0

	for _, v := range vlist {

		if val, ok := config.vars[v]; ok {

			if val.(float64) < min {

				min = val.(float64)
			}
		}
	}

	result := round(min)

	return result, true
}

func sum(triggerVar, currentAcond string, config Config) (interface{}, bool) {

	vlist := config.acond[currentAcond].Vars

	sum := 0.0

	for _, v := range vlist {

		if val, ok := config.vars[v]; ok {

			switch val.(type) {

			case float64:
				sum = sum + val.(float64)
			case bool:
				if val.(bool) == true {
					sum = sum + 1.0
				}
			}
		}
	}

	result := round(sum)

	return result, true
}

func and(triggerVar, currentAcond string, config Config) (interface{}, bool) {

	vlist := config.acond[currentAcond].Vars

	for _, v := range vlist {

		if val, ok := config.vars[v]; ok {

			if !val.(bool) {

				return false, true
			}
		}
	}

	return true, true
}

func or(triggerVar, currentAcond string, config Config) (interface{}, bool) {

	vlist := config.acond[currentAcond].Vars

	for _, v := range vlist {

		if val, ok := config.vars[v]; ok {

			if val.(bool) {

				return true, true
			}
		}
	}

	return false, true
}

func lastactive(triggerVar, currentAcond string, config Config) (interface{}, bool) {

	val := config.vars[triggerVar]

	switch val.(type) {

	case float64:
		if val.(float64) == 1 {
			return triggerVar, true
		}
	case string:
		if val.(string) == "1" {
			return triggerVar, true
		}
	case bool:
		if val.(bool) == true {
			return triggerVar, true
		}
	}

	return nil, false
}

func diff(triggerVar, currentAcond string, config Config) (interface{}, bool) {

	vlist := config.acond[currentAcond].Vars

	v0 := 0.0
	v1 := 0.0

	if val, ok := config.vars[vlist[0]]; ok {

		v0 = val.(float64)
	}

	if val, ok := config.vars[vlist[1]]; ok {

		v1 = val.(float64)
	}

	result := math.Abs(v0 - v1)

	return result, true
}

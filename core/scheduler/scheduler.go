package main

import (
	"encoding/json"
	"flag"
	"net/http"

	"github.com/bamzi/jobrunner"
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

/* Job */

// Job Specific Functions
type Schedule struct {
	Name        string
	active      bool
	domain      string
	variable    string
	value       interface{}
	sendChannel chan []byte
}

func (job Schedule) Run() {

	if job.active {

		p := make(map[string]interface{})
		p[job.variable] = job.value
		jsontools.GenerateRpcMessage(&job.sendChannel, "variables", "set", p, job.domain, "axihome")
	}
}

/* Main app */

func main() {

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
	started := false

	// Starting axihome communication

	go client.Run()

	http.HandleFunc("/", statehandler)
	go http.ListenAndServe(":3341", nil)

	for {

		select {

		case state := <-stateChannel:

			domain = state.Domain

			log.Println("New domain : " + domain)

			if state.Tld {

				jsontools.GenerateRpcMessage(&sendChannel, "bucket", "getAll", "Scheduler", instance+".core.axihome", "axihome")
			}

		case message := <-rpcMessageChannel:

			modfct := message.Body.Module + "." + message.Body.Fct

			if modfct == "bucket.setAll" {

				params := message.Body.Params.(map[string]interface{})

				if params["bucket"] == "Scheduler" {

					content := params["content"].(map[string]interface{})

					if started {
						jobrunner.Stop()
					}

					jobrunner.Start()
					started = true

					for k, v := range content {

						time := v.(map[string]interface{})["time"].(string)
						active := v.(map[string]interface{})["active"].(bool)
						jobs := v.(map[string]interface{})["write"].([]interface{})

						log.Println(k)
						log.Println(time)

						for _, jobdefif := range jobs {

							jobdef := jobdefif.(map[string]interface{})

							job := Schedule{
								Name:        k,
								domain:      domain,
								variable:    jobdef["variable"].(string),
								value:       jobdef["value"],
								sendChannel: sendChannel,
								active:      active,
							}

							jobrunner.Schedule(time, job)
						}
					}
				}
			}
		}
	}
}

func statehandler(w http.ResponseWriter, r *http.Request) {

	js, _ := json.Marshal(jobrunner.StatusJson())
	w.Write(js)
}

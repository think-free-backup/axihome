package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
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
	POOLINTERVAL = 3
)

type Answer struct {
	Contents struct {
		Quotes []struct {
			Author     string   `json:"author"`
			Background string   `json:"background"`
			Category   string   `json:"category"`
			Date       string   `json:"date"`
			ID         string   `json:"id"`
			Length     string   `json:"length"`
			Quote      string   `json:"quote"`
			Tags       []string `json:"tags"`
			Title      string   `json:"title"`
		} `json:"quotes"`
	} `json:"contents"`
	Success struct {
		Total int `json:"total"`
	} `json:"success"`
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

	for {

		select {

		case state := <-stateChannel:

			domain = state.Domain

			log.Println("New domain received : " + domain)

			url := "http://quotes.rest/qod.json"

			go func() {

				for {

					resp, err := http.Get(url)
					if err != nil {

						log.Println("Error in http request")
						time.Sleep(time.Hour)
						continue
					}
					body, err := ioutil.ReadAll(resp.Body)
					resp.Body.Close()
					if err != nil {

						log.Println("Error reading body")
						time.Sleep(time.Hour)
						continue
					}

					var ans Answer

					err = json.Unmarshal(body, &ans)
					if err != nil {

						log.Println("Error unmarshalling json")
						time.Sleep(time.Hour)
						continue
					}

					p := make(map[string]interface{})
					p["qod.quote"] = ans.Contents.Quotes[0].Quote
					p["qod.author"] = ans.Contents.Quotes[0].Author

					jsontools.GenerateRpcMessage(&sendChannel, "variables", "set", p, instance+".backend.axihome", "axihome")

					time.Sleep(POOLINTERVAL * time.Hour)
				}
			}()
		}
	}
}

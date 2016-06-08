package main

import (
	"encoding/json"
	"flag"
	"strings"
	"time"

	"github.com/lenkaiser/go.fitbit"
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

					config := &fitbit.Config{
						false, //Debug
						false, //Disable SSL
					}

					token := p["value"].(map[string]interface{})["params"].(map[string]interface{})["token"].(string)
					secret := p["value"].(map[string]interface{})["params"].(map[string]interface{})["secret"].(string)

					log.Println("Initializing fitbit with token : " + token + " and secret : " + secret)

					fapi, err := fitbit.NewAPI(token, secret, config)
					if err != nil {
						log.Println(err)
					}
					log.Println("API initialised")

					//Add client
					client, err := fapi.NewClient()
					if err != nil {
						log.Println(err)
					}
					log.Println("New client initialised")

					for {

						//Today
						date := time.Now()

						// Value map
						p := make(map[string]interface{})

						// Get value from api

						// -> Devices info
						dev, err1 := client.GetDevices()
						if err1 != nil {

							log.Println(err1.Error())
							time.Sleep(time.Hour)
							continue
						}
						p[instance+".battery"] = dev[0].Battery
						p[instance+".lastsync"] = formatTime(dev[0].LastSyncTime)

						log.Println(p[instance+".lastsync"])

						// -> Activity

						act, err2 := client.GetActivities(date)
						if err2 != nil {

							log.Println(err2.Error())
							time.Sleep(time.Hour)
							continue
						}
						p[instance+".floors"] = act.Summary.Floors
						p[instance+".steps"] = act.Summary.Steps
						p[instance+".distance"] = act.Summary.Distances[0].Distance
						p[instance+".sedentary"] = getDurationFromUnit(act.Summary.SedentaryMinutes, time.Minute)

						// -> Sleep

						sleep, err3 := client.GetSleep(date)
						if err3 != nil {

							log.Println(err3.Error())
							time.Sleep(time.Hour)
							continue
						}
						p[instance+".total.sleeptime"] = getDurationFromUnit(sleep.Summary.TotalMinutesAsleep, time.Minute)
						p[instance+".total.sleeprecords"] = sleep.Summary.TotalSleepRecords

						for i := 0; i < int(sleep.Summary.TotalSleepRecords); i++ {

							if sleep.Sleep[i].IsMainSleep {

								startTime := formatTime(sleep.Sleep[i].StartTime)
								p[instance+".mainsleep.timetobed"] = strings.Split(startTime, " ")[1]
								p[instance+".mainsleep.sleeptime"] = getDurationFromUnit(sleep.Sleep[i].MinutesAsleep, time.Minute)
								p[instance+".mainsleep.bedtime"] = getDurationFromUnit(sleep.Sleep[i].TimeInBed, time.Minute)
								p[instance+".mainsleep.efficiency"] = sleep.Sleep[i].Efficiency
								p[instance+".mainsleep.awakecount"] = sleep.Sleep[i].AwakeCount
								p[instance+".mainsleep.restlesscount"] = sleep.Sleep[i].RestlessCount
							}
						}

						// Send value to server
						jsontools.GenerateRpcMessage(&sendChannel, "variables", "set", p, instance+".backend.axihome", "axihome")

						// Wait next iteration
						time.Sleep(POOLINTERVAL * time.Minute)
					}
				}

			} else {

				js, _ := json.Marshal(message.Body.Params)
				log.Println("Unknow request received : " + message.Body.Module + "." + message.Body.Fct + " " + string(js))
			}
		}
	}
}

func formatTime(time string) string {

	return strings.TrimSuffix(strings.Replace(time, "T", " ", 1), ":00.000")
}

func getDurationFromUnit(duration uint64, unit time.Duration) string {

	var retdur time.Duration = time.Duration(duration) * unit
	return strings.TrimSuffix(retdur.String(), "m0s")
}

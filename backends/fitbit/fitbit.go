package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/fitbit"

	"github.com/think-free/axihome/backends/fitbit/types"

	"github.com/think-free/jsonrpc/client"
	"github.com/think-free/jsonrpc/common/messages"
	"github.com/think-free/jsonrpc/common/tools"
	"github.com/think-free/log"

	"github.com/think-free/axihome/server/variablemanager"
)

const (
	SERVERIP     = "localhost"
	SERVERPORT   = "3332"
	SERVERNAME   = "axihome"
	CACHEPATH    = "/etc/axihome/config/fitbit.json"
	POOLINTERVAL = 2
)

/* Main */
/* ***************************** */

func main() {

	flag.Parse()
	instance := flag.Arg(0)

	log.SetProcess(instance)

	sendChannel := make(chan []byte)
	stateChannel := make(chan *jsonrpcmessage.StateBody)
	rpcMessageChannel := make(chan *jsonrpcmessage.RpcMessage)

	client := jsonrpclient.New(instance, SERVERIP, SERVERPORT, sendChannel, stateChannel, rpcMessageChannel, nil, nil)
	go client.Run()

	history := variablemanager.NewHistory("http://" + SERVERIP + ":7777/")

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

				log.Println(message)

				p := message.Body.Params.(map[string]interface{})

				if p["bucket"] == "Instances" {

					clientId := p["value"].(map[string]interface{})["params"].(map[string]interface{})["clientId"].(string)
					clientSecret := p["value"].(map[string]interface{})["params"].(map[string]interface{})["clientSecret"].(string)

					// Generating client

					conf := &oauth2.Config{
						ClientID:     clientId,
						ClientSecret: clientSecret,
						Scopes:       []string{"heartrate", "activity", "profile", "sleep", "settings"},
						Endpoint:     fitbit.Endpoint,
						RedirectURL:  "http://localhost:7319/",
					}

					token, err := tokenFromFile(CACHEPATH)
					if err != nil && os.IsNotExist(err) {
						token, err = authorize(conf)
						js, _ := json.Marshal(token)
						log.Println(string(js))
					}
					if err != nil {
						log.Println(err)
					}

					c := fbclient(conf, token)

					for {

						// Value map
						p := make(map[string]interface{})

						// NEW

						// Requesting fitbit

						var hr types.HeartRate
						bt, _ := get(c, "https://api.fitbit.com/1/user/-/activities/heart/date/today/1d/1sec.json")
						json.Unmarshal(bt, &hr)

						var sleep types.Sleep
						bt, _ = get(c, "https://api.fitbit.com/1/user/-/sleep/date/today.json")
						json.Unmarshal(bt, &sleep)

						var act types.Activity
						bt, _ = get(c, "https://api.fitbit.com/1/user/-/activities/date/today.json")
						json.Unmarshal(bt, &act)

						var dev types.Devices
						bt, _ = get(c, "https://api.fitbit.com/1/user/-/devices.json")
						json.Unmarshal(bt, &dev)

						// Saving variables

						p[instance+".battery"] = dev[0].Battery
						p[instance+".lastsync"] = formatTime(dev[0].LastSyncTime)

						p[instance+".floors"] = act.Summary.Floors
						p[instance+".steps"] = act.Summary.Steps
						p[instance+".distance"] = act.Summary.Distances[0].Distance
						p[instance+".sedentary"] = getDurationFromUnit(act.Summary.SedentaryMinutes, time.Minute)

						p[instance+".total.sleeptime"] = getDurationFromUnit(sleep.Summary.TotalMinutesAsleep, time.Minute)
						p[instance+".total.sleeprecords"] = sleep.Summary.TotalSleepRecords

						// Main sleep reset

						p[instance+".mainsleep.timetobed"] = ""
						p[instance+".mainsleep.sleeptime"] = ""
						p[instance+".mainsleep.bedtime"] = ""
						p[instance+".mainsleep.efficiency"] = 0
						p[instance+".mainsleep.awakecount"] = 0
						p[instance+".mainsleep.restlesscount"] = 0

						// Get main sleep

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

						p[instance+".heartrate.resting"] = hr.Activities_heart[0].Value.RestingHeartRate

						for _, v := range hr.Activities_heart_intraday.Dataset {

							now := time.Now()
							fbtime := strings.Split(v.Time, ":")
							t, _ := strconv.ParseInt(fbtime[0], 10, 64)
							m, _ := strconv.ParseInt(fbtime[1], 10, 64)
							s, _ := strconv.ParseInt(fbtime[2], 10, 64)

							tm := time.Date(now.Year(), now.Month(), now.Day(), int(t), int(m), int(s), 0, time.UTC)

							//log.Println(tm.Unix(), v.Value)

							history.SaveWithTS("HeartRate", float64(v.Value), tm.Unix())
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

/* Helpers */
/* ***************************** */

func formatTime(time string) string {

	return strings.TrimSuffix(strings.Replace(time, "T", " ", 1), ":00.000")
}

func getDurationFromUnit(duration int, unit time.Duration) string {

	var retdur time.Duration = time.Duration(duration) * unit
	return strings.TrimSuffix(retdur.String(), "m0s")
}

/* Token managment */
/* ***************************** */

type cacherTransport struct {
	Base *oauth2.Transport
	Path string
}

func (c *cacherTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	cachedToken, err := tokenFromFile(c.Path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if _, err := c.Base.Source.Token(); err != nil {
		return nil, errors.New("expired token")
	}
	resp, err = c.Base.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	newTok, err := c.Base.Source.Token()
	if err != nil {
		// While we’re unable to obtain a new token, the request was still
		// successful, so let’s gracefully handle this error by not caching a
		// new token. In either case, the user will need to re-authenticate.
		return resp, nil
	}
	if cachedToken == nil ||
		cachedToken.AccessToken != newTok.AccessToken ||
		cachedToken.RefreshToken != newTok.RefreshToken {
		bytes, err := json.Marshal(&newTok)
		if err != nil {
			return nil, err
		}
		if err := ioutil.WriteFile(c.Path, bytes, 0600); err != nil {
			return nil, err
		}
	}
	return resp, nil
}

func tokenFromFile(path string) (*oauth2.Token, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var t oauth2.Token
	if err := json.Unmarshal(content, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func authorize(conf *oauth2.Config) (*oauth2.Token, error) {
	tokens := make(chan *oauth2.Token)
	errors := make(chan error)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := r.FormValue("code")
		if code == "" {
			http.Error(w, "Missing 'code' parameter", http.StatusBadRequest)
			return
		}
		tok, err := conf.Exchange(oauth2.NoContext, code)
		if err != nil {
			errors <- fmt.Errorf("could not exchange auth code for a token: %v", err)
			return
		}
		tokens <- tok
	})
	go func() {
		// Unfortunately, we need to hard-code this port — when registering
		// with fitbit, full RedirectURLs need to be whitelisted (incl. port).
		errors <- http.ListenAndServe(":7319", nil)
	}()

	authUrl := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	fmt.Println("Please visit the following URL to authorize:")
	fmt.Println(authUrl)
	select {
	case err := <-errors:
		return nil, err
	case token := <-tokens:
		return token, nil
	}
}

/* Oauth client */
/* ***************************** */

func fbclient(config *oauth2.Config, token *oauth2.Token) *http.Client {
	return &http.Client{
		Transport: &cacherTransport{
			Path: CACHEPATH,
			Base: &oauth2.Transport{
				Source: config.TokenSource(oauth2.NoContext, token),
			},
		},
	}
}

/* Get from fitbit api */
/* ***************************** */

func get(c *http.Client, url string) ([]byte, error) {

	response, err := c.Get(url)
	if err != nil {
		log.Println(err)
	} else {

		defer response.Body.Close()
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Println(err)
		}
		return body, err
	}

	return nil, err
}

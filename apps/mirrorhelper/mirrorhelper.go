package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/howeyc/fsnotify"
	"github.com/stianeikeland/go-rpio"

	"github.com/think-free/jsonrpc/client"
	"github.com/think-free/jsonrpc/common/messages"
	"github.com/think-free/jsonrpc/common/tools"
)

type Config struct {
	ServerIp         string
	ServerPort       string
	ServerName       string
	ClientName       string
	ButtonGpio       int
	ButtonVariable   string
	MotionFileWatch  string
	MotionVariable   string
	ScreenOnVariable string
}

func main() {

	/* Load config */

	content, e := ioutil.ReadFile("config.json")
	if e != nil {
		log.Printf("File error: %v\n", e)
		os.Exit(1)
	}

	var config Config
	json.Unmarshal(content, &config)

	/* Communication variables */

	var domain string
	var connected bool
	connected = false
	sendChannel := make(chan []byte)
	stateChannel := make(chan *jsonrpcmessage.StateBody)
	rpcMessageChannel := make(chan *jsonrpcmessage.RpcMessage)

	/* Watch file */

	log.Println("Watch file :", config.MotionFileWatch)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				if connected && ev.Name == config.MotionFileWatch+"/motiondetected" {

					log.Println(ev)

					if ev.IsCreate() {

						para := make(map[string]interface{})
						para[config.MotionVariable] = true
						jsontools.GenerateRpcMessage(&sendChannel, "variables", "set", para, domain, config.ServerName)

					} else if ev.IsDelete() {

						para := make(map[string]interface{})
						para[config.MotionVariable] = false
						jsontools.GenerateRpcMessage(&sendChannel, "variables", "set", para, domain, config.ServerName)
					}
				}

			case err := <-watcher.Error:
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Watch(config.MotionFileWatch)
	if err != nil {
		log.Fatal(err)
	}

	defer watcher.Close()

	log.Println("Starting :", config.ClientName, "using gpio :", config.ButtonGpio)

	client := jsonrpclient.New(config.ClientName, config.ServerIp, config.ServerPort, sendChannel, stateChannel, rpcMessageChannel, nil, nil)
	go client.Run()

	go func() {
		err := rpio.Open()

		if err != nil {

			os.Exit(0)
		}

		pin := rpio.Pin(config.ButtonGpio)

		pin.Input()
		pin.PullUp()

		latest := false

		for {

			if connected {

				detected := false
				res := pin.Read()

				if res == 0 {

					detected = true

					log.Println("Button pushed")
				}

				if latest != detected {

					log.Println(config.ButtonVariable, ":", detected)

					latest = detected

					para := make(map[string]interface{})
					para[config.ButtonVariable] = detected
					jsontools.GenerateRpcMessage(&sendChannel, "variables", "set", para, domain, config.ServerName)
				}
			}

			time.Sleep(100 * time.Millisecond)
		}
	}()

	for {

		select {

		case state := <-stateChannel:

			domain = state.Domain

			connected = state.Tld

			log.Println("New domain : " + state.Domain)

		case message := <-rpcMessageChannel:

			modfct := message.Body.Module + "." + message.Body.Fct

			if modfct == "variables.set" {

				vars := message.Body.Params.(map[string]interface{})

				for k, v := range vars {

					if k == config.ScreenOnVariable {

						var cmd string

						if v.(bool) == true {

							log.Println("Turning screen on")
							cmd = "tvservice -p; fbset -depth 8; fbset -depth 16;"

						} else {

							log.Println("Turning screen off")
							cmd = "tvservice -o"
						}

						var command *exec.Cmd
						command = exec.Command("bash", "-c", cmd)
						out, err := command.CombinedOutput()
						if err != nil {

							log.Println(out)
						}
					}
				}
			}
		}
	}
}

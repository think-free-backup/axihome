package instancesmanager

import (
	"archive/zip"
	"encoding/json"

	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/remyoudompheng/go-misc/zipfs"

	"github.com/think-free/jsonrpc/client"
	"github.com/think-free/jsonrpc/common/messages"
	"github.com/think-free/log"
)

type InstanceManager struct {

	// TODO : add mutex !

	instances InstanceMap

	develMode         bool
	sendChannel       chan []byte
	stateChannel      chan *jsonrpcmessage.StateBody
	rpcMessageChannel chan *jsonrpcmessage.RpcMessage
	client            *jsonrpclient.Client
}

/* Instance manager */
/* ******************************************** */

func New(dev bool) *InstanceManager {

	var im *InstanceManager = &InstanceManager{}

	im.develMode = dev

	im.sendChannel = make(chan []byte)
	im.stateChannel = make(chan *jsonrpcmessage.StateBody)
	im.rpcMessageChannel = make(chan *jsonrpcmessage.RpcMessage)
	im.client = jsonrpclient.New("instancemanager", "localhost", "3330", im.sendChannel, im.stateChannel, im.rpcMessageChannel, nil, nil)

	return im
}

func (im *InstanceManager) Run() {

	go im.client.Run()
	go im.channelHandler()

	log.Println("Axihome instance manager listening on port 3340")

	http.HandleFunc("/", im.ziphandler)
	http.HandleFunc("/getstate", im.statehandler)
	http.HandleFunc("/start", im.starthandler)
	http.HandleFunc("/stop", im.stophandler)
	http.HandleFunc("/restart", im.restarthandler)
	http.HandleFunc("/reload", im.reloadhandler)
	http.HandleFunc("/stats", im.statsHandler)

	http.ListenAndServe(":3340", nil)
}

/* Communication */
/* ******************************************** */

func (im *InstanceManager) channelHandler() {

	for {

		select {

		// Rpc
		case message := <-im.rpcMessageChannel:

			if message.Body.Module+"."+message.Body.Fct == "bucket.setAll" {

				params := message.Body.Params.(map[string]interface{})

				if params["bucket"] == "Instances" {

					if im.instances == nil {

						im.instances = make(InstanceMap)

					} else {
						return
					}

					instances := params["content"].(map[string]interface{})

					// Saving current instances

					current := make(map[string]bool)

					for k := range im.instances {
						log.Println("Instance manager : Found running instance : " + k)
						current[k] = false
					}

					// Creating instance if not exists currently

					for _, v := range instances {

						val := v.(map[string]interface{})

						instance := Instance{
							Name:           val["name"].(string),
							Backend:        val["backend"].(string),
							Params:         val["params"],
							Run:            val["run"].(bool),
							Process:        nil,
							ProcessRunning: false,
						}

						if im.instances[instance.Backend+"-"+instance.Name] == nil {

							log.Println("Instance manager : Creating instance " + instance.Name + "-" + instance.Backend)

							im.instances[instance.Backend+"-"+instance.Name] = &instance

							if instance.Run && !im.develMode {

								im.startProcess(instance.Backend, instance.Name)
							}
						} else {

							current[instance.Name+"-"+instance.Backend] = true
						}
					}

					// Stopping and deleting removed instances

					for k, v := range current {

						if !v {

							log.Println("Instance manager : Stopping : " + k)
							im.stopProcess(im.instances[k].Backend, im.instances[k].Name)
							delete(im.instances, k) // TODO : FIXME : CRASH
						}
					}
				}
			}

		// State
		case state := <-im.stateChannel:

			if state.Tld {
				im.getInstances()
			}
		}
	}
}

func (im *InstanceManager) getInstances() {

	body := &jsonrpcmessage.RpcBody{Module: "bucket", Fct: "getAll", Params: "Instances"}

	mes := jsonrpcmessage.NewRoutedMessage("rpc", body, "instancemanager.core.axihome", "axihome")
	js, _ := json.Marshal(mes)
	im.sendChannel <- js
}

/* Process managment base */
/* ******************************************** */

func (im *InstanceManager) startProcess(process, instance string) {

	pname := process + "-" + instance

	if _, ok := im.instances[pname]; !ok {

		log.Println("Can't find instance : " + pname)
	}

	if im.instances[pname].ProcessRunning {

		log.Println("Can't start process, already started : " + pname)
		return
	}

	comchan := make(chan string)

	im.instances[pname].Process = &comchan
	im.instances[pname].ProcessRunning = true

	go func() {

		for {
			log.Println("Starting process : " + process)

			cmd := exec.Command("bin/"+process, instance)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Start()

			done := make(chan error, 1)
			go func() {
				done <- cmd.Wait()
			}()

			select {

			case mes := <-comchan:

				if mes == "restart" {

					log.Println("Restarting process : " + process)
					cmd.Process.Kill()
					cmd.Wait()

				} else if mes == "stop" {

					log.Println("Stopping process : " + process)
					cmd.Process.Kill()
					cmd.Wait()

					log.Println("Removing channel for ", process, "instance", instance)
					im.instances[pname].Process = nil
					im.instances[pname].ProcessRunning = false
					return
				}

			case err := <-done:

				log.Println("Process : " + process + " stopped")
				if err != nil {
					log.Println("    |-> with error :", err)

					log.Println("    |-> desactivating process for 10 minutes")
					time.Sleep(10 * time.Minute)
				}
			}
		}

		log.Println("Removing channel for ", process, "instance", instance)
		im.instances[pname].Process = nil
		im.instances[pname].ProcessRunning = false
	}()
}

func (im *InstanceManager) stopProcess(process, instance string) {

	ist := im.instances[process+"-"+instance]
	if ist != nil {
		c := ist.Process

		if c == nil {

			log.Println("Channel not found for ", process, "instance", instance)
		} else {

			*c <- "stop"
		}
	}
}

func (im *InstanceManager) restartProcess(process, instance string) {

	ist := im.instances[process+"-"+instance]

	if ist != nil {
		c := ist.Process

		if c == nil {

			log.Println("Channel not found for ", process, "instance", instance)
		} else {

			*c <- "restart"
		}
	}
}

/* Web server */
/* ******************************************** */

func (im *InstanceManager) statehandler(w http.ResponseWriter, r *http.Request) {

	js, err := json.Marshal(im.instances)

	if err != nil {

		log.Println(err)
	}

	w.Write(js)
}

func (im *InstanceManager) statsHandler(w http.ResponseWriter, r *http.Request) {

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	str := fmt.Sprintf("Version : %s \nRoutines : %d\nMem Alloc : %d\nMem TotalAlloc : %d\nMem HeapAlloc : %d\nMem HeapSys %d",
		runtime.Version(),
		runtime.NumGoroutine(),
		mem.Alloc,
		mem.TotalAlloc,
		mem.HeapAlloc,
		mem.HeapSys)

	w.Write([]byte(str))
}

func (im *InstanceManager) ziphandler(w http.ResponseWriter, r *http.Request) {

	zippath := "assets/axihome.assets"

	z, err := zip.OpenReader(zippath)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	defer z.Close()

	r.URL.Path = "/assets/" + r.URL.Path

	http.FileServer(zipfs.NewZipFS(&z.Reader)).ServeHTTP(w, r)
}

func (im *InstanceManager) starthandler(w http.ResponseWriter, r *http.Request) {

	p := r.URL.Query()
	im.startProcess(p.Get("process"), p.Get("instance"))
}

func (im *InstanceManager) stophandler(w http.ResponseWriter, r *http.Request) {

	p := r.URL.Query()
	im.stopProcess(p.Get("process"), p.Get("instance"))
}

func (im *InstanceManager) restarthandler(w http.ResponseWriter, r *http.Request) {

	p := r.URL.Query()
	im.restartProcess(p.Get("process"), p.Get("instance"))
}

func (im *InstanceManager) reloadhandler(w http.ResponseWriter, r *http.Request) {

	im.getInstances()
}

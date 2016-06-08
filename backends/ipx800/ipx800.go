package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/think-free/jsonrpc/client"
	"github.com/think-free/jsonrpc/common/messages"
	"github.com/think-free/jsonrpc/common/tools"
	"github.com/think-free/log"
)

const (
	SERVERIP   = "localhost"
	SERVERPORT = "3332"
	SERVERNAME = "axihome"
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
	var ipxconn net.Conn

	for {

		select {

		case state := <-stateChannel:

			domain = state.Domain

			log.Println("New domain : " + state.Domain)

			if state.Tld {

				// Backend

				params := make(map[string]string)
				params["bucket"] = "Instances"
				params["variable"] = flag.Arg(0)
				jsontools.GenerateRpcMessage(&sendChannel, "bucket", "getVar", params, domain, SERVERNAME)
			}

		case message := <-rpcMessageChannel:

			if (message.Body.Module + "." + message.Body.Fct) == "bucket.setVar" {

				p := message.Body.Params.(map[string]interface{})

				if p["bucket"] == "Instances" {

					ipx := p["value"].(map[string]interface{})["params"].(map[string]interface{})["ip"].(string)

					for {

						log.Println(instance + " connecting to " + ipx + "...")

						conn, err := net.Dial("tcp", ipx+":9870")
						if err != nil {
							log.Println(err)
							time.Sleep(5 * time.Second)
							continue
						}

						ipxconn = conn

						log.Println(instance + " connected ...")

						connbuf := bufio.NewReader(ipxconn)

						cache := make(map[string]interface{})

						for {

							str, err := connbuf.ReadString('\n')
							if len(str) > 0 {

								strarr := strings.Split(str, "&")

								params := make(map[string]interface{})

								addToParams(instance+".ts", float64(time.Now().Unix()), params, cache)

								for _, val := range strarr {

									kv := strings.Split(val, "=")

									if kv[0] == "I" || kv[0] == "O" {

										io := "input"

										if kv[0] == "O" {

											io = "output"
										}

										addToParams(instance+"."+io+".all", kv[1], params, cache)

										for i := 0; i < 32; i++ {

											b := false

											if string(kv[1][i]) == "1" {
												b = true
											}

											addToParams(instance+"."+io+"."+strconv.Itoa(i), b, params, cache)
										}

									} else {

										ac := "analog"

										if string(kv[0][0]) == "C" {

											ac = "counter"
										}

										flt, _ := strconv.ParseFloat(kv[1], 64)
										addToParams(instance+"."+ac+"."+string(kv[0][1]), flt, params, cache)
									}
								}
								jsontools.GenerateRpcMessage(&sendChannel, "variables", "set", params, domain, SERVERNAME)
							}
							if err != nil {
								break
							}
						}

						log.Println(instance + " disconnected, trying to connect ...")
						time.Sleep(5 * time.Second)
					}
				}

			} else if (message.Body.Module + "." + message.Body.Fct) == "variables.write" {

				params := message.Body.Params.(map[string]interface{})

				for k, v := range params {

					karr := strings.Split(k, ".")

					pin := karr[2]
					if len(pin) == 1 {
						pin = "0" + pin
					}

					if karr[1] == "impulse" {

						ipxconn.Write([]byte("Set" + pin + "1p"))

					} else if karr[1] == "output" {

						if pin == "all" {

							// Write mask or double mask

							log.Println("WRITE MASK OR DOUBLE MASK NOT IMPLEMENTED")

						} else {

							if v.(bool) {

								ipxconn.Write([]byte("Set" + pin + "1"))

							} else {

								ipxconn.Write([]byte("Set" + pin + "0"))
							}
						}
					}
				}

			} else {

				js, _ := json.Marshal(message.Body.Params)
				log.Println("Unknow request received : " + message.Body.Module + "." + message.Body.Fct + " " + string(js))
			}
		}
	}
}

func addToParams(k string, v interface{}, params, cache map[string]interface{}) {

	if cvalue, ok := cache[k]; ok {

		if cvalue != v {

			params[k] = v
		}
	}

	cache[k] = v
}

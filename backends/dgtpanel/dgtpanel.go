package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/think-free/jsonrpc/client"
	"github.com/think-free/jsonrpc/common/messages"
	"github.com/think-free/jsonrpc/common/tools"
	"github.com/think-free/log"
)

type PanelInfo struct {
	NumAlternancias           string `json:"NumAlternancias"`
	DrawDer1                  string `json:"drawDer1"`
	DrawDer2                  string `json:"drawDer2"`
	DrawIz1                   string `json:"drawIz1"`
	DrawIz2                   string `json:"drawIz2"`
	ImgTxtDer1                string `json:"imgTxtDer1"`
	ImgTxtDer2                string `json:"imgTxtDer2"`
	ImgTxtIzq1                string `json:"imgTxtIzq1"`
	ImgTxtIzq2                string `json:"imgTxtIzq2"`
	IndiceMapa                string `json:"indiceMapa"`
	Mensaje1                  string `json:"mensaje1"`
	Mensaje2                  string `json:"mensaje2"`
	TextoAdvertenciaPrecision string `json:"textoAdvertenciaPrecision"`
	Tipo                      string `json:"tipo"`
}

const (
	SERVERIP     = "localhost"
	SERVERPORT   = "3332" // -> backend
	SERVERNAME   = "axihome"
	POOLINTERVAL = 5

	URLBASE = "http://infocar.dgt.es/etraffic/BuscarElementos?accion=getDetalles&codEle="
	URLEND  = "&tipo=Panel_CMS&indiceMapa=0"
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
	initialized := false

	reg := regexp.MustCompile(`<.*?>`)

	for {

		select {

		case state := <-stateChannel:

			domain = state.Domain

			log.Println("New domain : " + state.Domain)

			if state.Tld {

				params := make(map[string]string)
				params["bucket"] = "Instances"
				params["variable"] = flag.Arg(0)

				jsontools.GenerateRpcMessage(&sendChannel, "bucket", "getVar", params, domain, SERVERNAME)
			}

		case message := <-rpcMessageChannel:

			if (message.Body.Module + "." + message.Body.Fct) == "bucket.setVar" { // Backend

				p := message.Body.Params.(map[string]interface{})

				if p["bucket"] == "Instances" {

					if !initialized {

						initialized = true

						panels := p["value"].(map[string]interface{})["params"].([]interface{})

						// Backend

						go func() {

							for {

								params := make(map[string]interface{})

								for _, pan := range panels {

									resp, err := http.Get(URLBASE + pan.(string) + URLEND)
									if err != nil {

										log.Println("Error in http request")
										time.Sleep(time.Hour)
										continue
									}
									body, err := ioutil.ReadAll(resp.Body)
									resp.Body.Close()

									var panel PanelInfo

									err = json.Unmarshal(body, &panel)
									if err != nil {

										log.Println("Error unmarshalling json")
										time.Sleep(time.Hour)
										continue
									}

									params[instance+"."+pan.(string)+".message.1"] = getString(reg, panel.Mensaje1)
									params[instance+"."+pan.(string)+".message.2"] = getString(reg, panel.Mensaje2)

									log.Println(instance + "." + pan.(string) + ".message.1 : " + pan.(string) + " : " + params[instance+"."+pan.(string)+".message.1"].(string))
									log.Println(instance + "." + pan.(string) + ".message.2 : " + pan.(string) + " : " + params[instance+"."+pan.(string)+".message.2"].(string))
								}

								jsontools.GenerateRpcMessage(&sendChannel, "variables", "set", params, instance+".backend.axihome", "axihome")

								time.Sleep(POOLINTERVAL * time.Minute)
							}
						}()
					}
				}

			} else {

				js, _ := json.Marshal(message.Body.Params)
				log.Println("Unknow request received : " + message.Body.Module + "." + message.Body.Fct + " " + string(js))
			}
		}
	}
}

func getString(reg *regexp.Regexp, src string) string {

	return removeDoubleSpaces(string(reg.ReplaceAllLiteral([]byte(src), []byte(" "))))
}

func removeDoubleSpaces(str string) string {

	retstr := str
	for {

		if !strings.Contains(retstr, "  ") {

			return retstr
		}

		retstr = strings.Replace(retstr, "  ", " ", -1)
	}
}

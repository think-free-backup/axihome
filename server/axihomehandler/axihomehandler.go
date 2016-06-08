package axihomehandler

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/think-free/axihome/server/variablemanager"
	"github.com/think-free/jsonrpc/common/messages"
	"github.com/think-free/log"
)

type Handler struct {
	generaldb *variablemanager.VariableManager
	rtdb      *variablemanager.VariableManager
	historydb *variablemanager.VariableManager

	bucketList map[string]*variablemanager.VariableManager
	sendChan   *chan jsonrpcmessage.RoutedMessage
	analogic   map[string]interface{}

	mapping map[string]string
}

func New(sendChan *chan jsonrpcmessage.RoutedMessage, configPath string) *Handler {

	appHandler := &Handler{}

	// Bucket registration

	appHandler.bucketList = make(map[string]*variablemanager.VariableManager)

	// Create  db and buckets

	vmgeneraldb := variablemanager.New(configPath + "/axihome.db")
	appHandler.generaldb = vmgeneraldb
	appHandler.RegisterBucket("Variables")
	appHandler.RegisterBucket("VariablesAddrMapping")

	appHandler.RegisterBucket("Instances")
	appHandler.RegisterBucket("Config")

	// TODO : Following bucket should come from config and shouldn't be hardcoded here

	bucketcontent := appHandler.bucketList["Config"].Get("Config", "buckets")
	if bucketcontent != nil {

		for _, v := range bucketcontent.([]interface{}) {

			log.Println("Registering application bucket : " + v.(string))
			appHandler.RegisterBucket(v.(string))
		}

	} else {

		log.Println("Can't read bucket")
	}

	// Realtime database

	vmrtdb := variablemanager.New(configPath + "/axihome-rtdb.db")

	appHandler.rtdb = vmrtdb

	vmrtdb.CreateBucket("RealtimeDB")
	appHandler.bucketList["RealtimeDB"] = vmrtdb

	// Historic database

	vmhistorydb := variablemanager.New(configPath + "/axihome-history.db")

	appHandler.historydb = vmhistorydb

	vmhistorydb.CreateBucket("History")
	appHandler.bucketList["History"] = vmhistorydb

	vmhistorydb.CreateBucket("HistoryRtdbDump")
	appHandler.bucketList["HistoryRtdbDump"] = vmhistorydb

	// Save analog and full rtdb dump in history

	analogTicker := time.NewTicker(time.Minute * 5).C
	fullDbTicker := time.NewTicker(time.Hour).C
	go func() {
		for {
			select {
			case <-analogTicker:

				log.Println("Saving analog variables in historic")
				now := strconv.FormatInt(time.Now().Unix(), 10)

				for k, v := range appHandler.analogic {

					historyvar := make(map[string]interface{})
					historyvar["Key"] = k
					historyvar["Value"] = v

					err := appHandler.historydb.Set("History", now, historyvar)
					if err != nil {

						log.Println("Can't save history :", now, "->", historyvar)
					}
				}

			case <-fullDbTicker:

				rtdbcontent := appHandler.rtdb.GetAll("RealtimeDB")
				if rtdbcontent != nil {

					js, _ := json.Marshal(rtdbcontent)

					log.Println("Saving rtdb dump in historic")
					now := strconv.FormatInt(time.Now().Unix(), 10)

					err := appHandler.historydb.Set("HistoryRtdbDump", now, js)
					if err != nil {

						log.Println("Can't save history :", now, "->", js)
					} else {

						log.Println(string(js))
					}

				} else {

					log.Println("Can't read rtdb")
				}
			}
		}
	}()

	// Set app handler variables

	appHandler.sendChan = sendChan
	appHandler.analogic = make(map[string]interface{})

	// Generated RTDB

	appHandler.generateRtdbMissingValues()

	appHandler.generateKeyAddrMapping()

	// Return the Handler

	return appHandler
}

/* Db Creation */
/* ************************************************************** */

func (handler *Handler) RegisterBucket(name string) {

	handler.generaldb.CreateBucket(name)
	handler.bucketList[name] = handler.generaldb
}

func (handler *Handler) generateRtdbMissingValues() {

	log.Println("Generating default value for rtdb")

	variablesbucketcontent := handler.generaldb.GetAll("Variables")
	rtdbcontent := handler.rtdb.GetAll("RealtimeDB")

	if variablesbucketcontent != nil {

		for k, v := range variablesbucketcontent {

			if _, ok := rtdbcontent[k]; !ok {

				vmap := v.(map[string]interface{})
				value := vmap["default"]
				if value != nil {

					handler.rtdb.Set("RealtimeDB", k, vmap["default"])
					log.Println("Setting default value for " + k)
				}
			}
		}
	}
}

func (handler *Handler) generateKeyAddrMapping() {

	log.Println("Generating key mapping for variables")

	handler.mapping = make(map[string]string)

	bucket := handler.bucketList["Variables"]
	if bucket == nil {

		log.Println("Can't find Variables bucket")
		return
	}

	bucketcontent := bucket.GetAll("Variables")

	for k, v := range bucketcontent {

		addr := v.(map[string]interface{})["addr"].(string)
		if addr != "" {

			handler.mapping[addr] = k
		} else {

			handler.mapping[k] = k
		}
	}
}

/* Type handlers */
/* ************************************************************** */

func (handler *Handler) Rpc(mes jsonrpcmessage.RoutedMessage) error {

	body, err := jsonrpcmessage.GerRpcBodyFromMesBody(mes.Body)
	if err != nil {
		return err
	}

	if body.Module == "variables" {

		switch body.Fct {

		case "set": // set variable (variable changed in backend -> bucket)

			go func() {

				log.Debug("Variable set request")

				now := strconv.FormatInt(time.Now().Unix(), 10)

				if body.Params == nil {
					return
				}

				for k, v := range body.Params.(map[string]interface{}) {

					vname := handler.mapping[k]
					vconf := handler.generaldb.Get("Variables", vname)

					if vconf == nil {

						log.Println("Variable not found : addr '" + k + "' variable name '" + k + "' not found")
						continue
					}

					if handler.rtdb.Get("RealtimeDB", vname) == v {

						log.Debug("Variable set " + vname + " is the same not updating")
						continue
					}

					log.Debug("Setting variable :", vname, "to", v)

					params := make(map[string]interface{})
					params[vname] = v
					body.Params = params

					mesCore := jsonrpcmessage.NewRoutedMessage("rpc", body, "axihome", "*.core.axihome")
					mesFrontend := jsonrpcmessage.NewRoutedMessage("rpc", body, "axihome", "*.frontend.axihome")
					*handler.sendChan <- *mesCore
					*handler.sendChan <- *mesFrontend

					handler.rtdb.Set("RealtimeDB", vname, v)

					if !vconf.(map[string]interface{})["analog"].(bool) {

						historyvar := make(map[string]interface{})
						historyvar["Key"] = vname
						historyvar["Value"] = v

						err := handler.historydb.Set("History", now, historyvar)
						if err != nil {

							log.Println("Can't save history :", now, "->", historyvar)
						}
					} else {

						handler.analogic[vname] = v
					}

					log.Debug("Variable set " + vname + " saved")
				}

			}()

		case "write": // write variable (from app to backend)

			go func() {

				for k, v := range body.Params.(map[string]interface{}) {

					vname := k
					vval := v
					vconf := handler.generaldb.Get("Variables", k)

					if vconf == nil {

						variables := handler.generaldb.GetAll("Variables")

						for k, v := range variables {

							if v.(map[string]interface{})["addr"].(string) == k {

								vname = k
								vconf = v
								break
							}
						}
					}

					if vconf != nil {

						params := make(map[string]interface{})
						params[vname] = vval
						body.Params = params

						log.Debug("Will write variable to backend : " + vname)
						log.Debug(vconf)

						mes := jsonrpcmessage.NewRoutedMessage("rpc", body, "axihome", vconf.(map[string]interface{})["backend"].(string)+".backend.axihome")
						*handler.sendChan <- *mes

					} else {

						log.Println("\x1b[31mCan't find variable definition for " + k + "\x1b[0m")
					}
				}
			}()

		case "getAll": // get RTDB content

			go func() {

				rtdbcontent := handler.rtdb.GetAll("RealtimeDB")
				if rtdbcontent != nil {

					body := &jsonrpcmessage.RpcBody{Module: "variables", Fct: "set", Params: rtdbcontent}

					mes := jsonrpcmessage.NewRoutedMessage("rpc", body, "axihome", mes.Src)
					*handler.sendChan <- *mes

				} else {

					log.Println("Can't read rtdb")
				}

			}()

		case "generateRtdbMissingValues":

			handler.generateRtdbMissingValues()
		}

	} else if body.Module == "bucket" {

		switch body.Fct {

		// Get variable from any bucket
		case "getVar":

			params := body.Params.(map[string]interface{})
			bname := params["bucket"].(string)
			vname := params["variable"].(string)

			log.Println("Bucket variable requested : " + bname + " -> " + vname)

			bucketcontent := handler.bucketList[bname].Get(bname, vname)
			if bucketcontent != nil {

				p := make(map[string]interface{})
				p["bucket"] = bname
				p["variable"] = vname
				p["value"] = bucketcontent

				body := &jsonrpcmessage.RpcBody{Module: "bucket", Fct: "setVar", Params: p}

				mes := jsonrpcmessage.NewRoutedMessage("rpc", body, "axihome", mes.Src)
				*handler.sendChan <- *mes

			} else {

				log.Println("Can't read bucket")
			}

		// Get all variable from any bucket
		case "getAll":

			bname := body.Params.(string)

			log.Println("Full bucket requested : " + bname + " by " + mes.Src)

			bucket := handler.bucketList[bname]
			if bucket == nil {

				log.Println("Can't find bucket " + bname)
				return nil
			}

			bucketcontent := bucket.GetAll(bname)

			if bucketcontent != nil {

				p := make(map[string]interface{})
				p["bucket"] = bname
				p["content"] = bucketcontent

				body := &jsonrpcmessage.RpcBody{Module: "bucket", Fct: "setAll", Params: p}

				mes := jsonrpcmessage.NewRoutedMessage("rpc", body, "axihome", mes.Src)
				*handler.sendChan <- *mes

			} else {

				log.Println("Can't read bucket")
			}

		// Set variable to any bucket
		case "setVar":

			params := body.Params.(map[string]interface{})
			bname := params["bucket"].(string)
			vname := params["variable"].(string)
			vval := params["value"].(string)

			log.Println("Bucket set variable requested : " + bname + " -> " + vname)

			bucketcontent := handler.bucketList[bname].SetJson(bname, vname, vval)
			if bucketcontent != nil {

				log.Println("Variable set error")
			}

		// Set bucket content
		case "setAll":

			params := body.Params.(map[string]interface{})
			bname := params["bucket"].(string)
			content := []byte(params["content"].(string))

			log.Println("Bucket setAll variable requested : " + bname)

			handler.RegisterBucket(bname)

			err := handler.bucketList[bname].SetAll(bname, content)
			if err != nil {

				log.Println("Variable setAll error")
			} else {

				p := make(map[string]interface{})
				p["bucket"] = bname

				body := &jsonrpcmessage.RpcBody{Module: "bucket", Fct: "setAllAck", Params: p}

				mes := jsonrpcmessage.NewRoutedMessage("rpc", body, "axihome", mes.Src)
				*handler.sendChan <- *mes
			}
		}

	} else if body.Module == "history" {

		switch body.Fct {

		// Get variable from any bucket
		case "get":

			params := body.Params.(map[string]interface{})
			start := params["start"].(float64)
			end := params["end"].(float64)

			startstr := strconv.FormatInt(int64(start), 10)
			endstr := strconv.FormatInt(int64(end), 10)

			//handler.historydb.
			bucketcontent := handler.historydb.GetRange("History", startstr, endstr)
			if bucketcontent != nil {

				p := make(map[string]interface{})
				p["start"] = start
				p["end"] = end
				p["value"] = bucketcontent

				body := &jsonrpcmessage.RpcBody{Module: "history", Fct: "set", Params: p}

				mes := jsonrpcmessage.NewRoutedMessage("rpc", body, "axihome", mes.Src)
				*handler.sendChan <- *mes

			} else {

				log.Println("Can't read bucket")
			}

		case "getDump":
		}
	}

	return nil
}

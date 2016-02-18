package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/boltdb/bolt"

	"github.com/think-free/jsonrpc/client"
	"github.com/think-free/jsonrpc/common/messages"
)

func main() {

	var file = flag.String("db", "axihome.db", "file : filename")
	var action = flag.String("a", "none", "action : get, set")
	var bucket = flag.String("b", "bucket", "bucket name")
	var importfile = flag.String("f", "file.json", "file to import")
	var server = flag.String("s", "none", "server ip for online import")
	var port = flag.String("p", "", "server port for online import")

	flag.Parse()

	if *server == "none" {

		db, err := bolt.Open(*file, 0666, nil)
		if err != nil {
			fmt.Println(err)
		}
		defer db.Close()

		switch *action {

		case "get":
			getfile(db, *bucket)

		case "set":
			setfile(db, *bucket, *importfile)

		default:
			fmt.Println("Unknow action")
		}
	} else {

		sendChannel := make(chan []byte)
		stateChannel := make(chan *jsonrpcmessage.StateBody)
		rpcMessageChannel := make(chan *jsonrpcmessage.RpcMessage)

		client := jsonrpclient.New("axdbManager"+*bucket, *server, *port, sendChannel, stateChannel, rpcMessageChannel, nil, nil)
		go client.Run()

		for {
			select {

			case state := <-stateChannel:

				fmt.Println("Importing", *bucket, "to server", *server)
				content, e := ioutil.ReadFile(*importfile)
				if e != nil {
					fmt.Printf("File error: %v\n", e)
					os.Exit(1)
				}

				body := make(map[string]interface{})
				body["module"] = "bucket"
				body["fct"] = "setAll"
				params := make(map[string]interface{})
				params["bucket"] = *bucket
				params["content"] = string(content)
				body["params"] = params

				mes := jsonrpcmessage.NewRoutedMessage("rpc", body, state.Domain, "axihome")
				json, _ := json.Marshal(mes)
				sendChannel <- json

			case message := <-rpcMessageChannel:

				if message.Body.Fct == "setAllAck" {
					os.Exit(0)
				}
			}
		}
	}
}

func getfile(db *bolt.DB, bucket string) {

	fmt.Println(bucket + " content :")

	db.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(bucket))

		b.ForEach(func(k, v []byte) error {

			fmt.Printf("- %s -> %s\n", k, v)
			return nil
		})
		return nil
	})
}

func setfile(db *bolt.DB, bucket, file string) {

	fmt.Println("")
	fmt.Println("\x1b[31mImporting variable from : " + file + " to " + bucket + "\x1b[0m")
	fmt.Println("")

	db.Batch(func(tx *bolt.Tx) error {

		tx.DeleteBucket([]byte(bucket))
		b, err := tx.CreateBucket([]byte(bucket))

		content, e := ioutil.ReadFile(file)
		if e != nil {
			fmt.Printf("File error: %v\n", e)
			os.Exit(1)
		}

		var js interface{}
		json.Unmarshal(content, &js)

		conf := js.(map[string]interface{})

		for k, v := range conf {

			js, _ := json.Marshal(v)
			fmt.Println("\x1b[32m" + k + "\x1b[0m : " + string(js))
			fmt.Println("")
			b.Put([]byte(k), js)
		}

		return err
	})
}

func getnet(bucket string) {

}

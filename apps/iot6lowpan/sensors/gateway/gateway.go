package gateway

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	POOLINTERVAL = 5
)

func Run(c *chan map[string]interface{}, paramsIface interface{}) {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	params := paramsIface.([]interface{})

	for {

		m := make(map[string]interface{})

		for _, v := range params {

			e := false

			fmt.Println("Getting sensor value for " + v.(string))
			ip := v.(string)

			client := &http.Client{Transport: tr}
			resp, err := client.Get("http://[" + ip + "]:6103")
			if err != nil {

				log.Println("Error in http request")
				m[ip+".error"] = true
				e = true
			}

			if !e {

				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)

				bodyjson := make(map[string]interface{})

				err = json.Unmarshal(body, &bodyjson)

				if err != nil {

					log.Println(err)
					m[ip+".error"] = true
					e = true
				}

				if !e {

					for k, value := range bodyjson {

						m[ip+"."+k] = value
					}

					m[ip+".error"] = false
					m[ip+".ts"] = float64(time.Now().Unix())
				}
			}
		}

		*c <- m

		time.Sleep(POOLINTERVAL * time.Second)
	}
}

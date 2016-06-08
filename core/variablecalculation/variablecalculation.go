/*
	This core app apply a formula to a variable to calculate other variable

	Sample config :

		{
		    "calc-ipxshutters-an-1-temperature" :  {"formula" : "PercentMinMax", "params" : {"min" : "heating.deposit.min", "max" : "heating.deposit.max"} ,"write" : "heating.deposit.percent"}
		}

	Available formulas :

		- PercentMinMax :  	calc = (value - min) /  ((max-min) / 100)
							if calc <0 -> 0
							if calc > 100 -> 100
							else -> calc
		- IpxTC100
		- IpxTC4012
		- IpxTC5050
		- IpxLights
		- IpxRH100
*/

package main

import (
	"errors"
	"flag"
	"math"
	"reflect"

	"github.com/think-free/jsonrpc/client"
	"github.com/think-free/jsonrpc/common/messages"
	"github.com/think-free/jsonrpc/common/tools"
	"github.com/think-free/log"
)

const (
	SERVERIP   = "localhost"
	SERVERPORT = "3330"
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

	// Starting axihome communication

	go client.Run()

	initialized := false
	calcmap := make(map[string]interface{})
	fct := NewCalc()

	for {

		select {

		case state := <-stateChannel:

			log.Println("New domain : " + state.Domain)

			if state.Tld {

				jsontools.GenerateRpcMessage(&sendChannel, "variables", "getAll", "", "variablecalculation.core.axihome", "axihome")
				jsontools.GenerateRpcMessage(&sendChannel, "bucket", "getAll", "VariablesCalculation", "variablecalculation.core.axihome", "axihome")
			}

		case message := <-rpcMessageChannel:

			modfct := message.Body.Module + "." + message.Body.Fct

			if modfct == "bucket.setAll" {

				params := message.Body.Params.(map[string]interface{})

				if params["bucket"] == "VariablesCalculation" {

					log.Println("Received binding list")

					calcmap = params["content"].(map[string]interface{})

					initialized = true
				}

			} else if modfct == "variables.set" {

				params := message.Body.Params.(map[string]interface{})

				for k, v := range params {

					fct.Keyvaluemap[k] = v

					if initialized {

						if calc, ok := calcmap[k]; ok {

							calcm := calc.(map[string]interface{})

							variableToWrite := calcm["write"].(string)

							res, err := fct.Call(calcm["formula"].(string), variableToWrite, v, calcm["params"])

							if err == nil {

								if fct.Keyvaluemap[variableToWrite] != res {

									p := make(map[string]interface{})
									p[variableToWrite] = res

									jsontools.GenerateRpcMessage(&sendChannel, "variables", "set", p, "variablecalculation.core.axihome", "axihome")
								}
							} else {

								log.Println(err.Error())
							}
						}
					}
				}
			}
		}
	}
}

type Calcs struct {
	fctmap      map[string]func(string, interface{}, interface{}) (interface{}, error)
	Keyvaluemap map[string]interface{}
}

func NewCalc() *Calcs {

	c := Calcs{}

	c.Keyvaluemap = make(map[string]interface{})

	c.fctmap = map[string]func(string, interface{}, interface{}) (interface{}, error){

		"PercentMinMax": c.PercentMinMax,
		"IpxTC100":      c.IpxTC100,
		"IpxTC4012":     c.IpxTC4012,
		"IpxTC5050":     c.IpxTC5050,
		"IpxLights":     c.IpxLights,
		"IpxRH100":      c.IpxRH100,
	}

	return &c
}

func (c *Calcs) Call(fct, variableToWrite string, variable interface{}, params interface{}) (interface{}, error) {

	return c.fctmap[fct](variableToWrite, variable, params)
}

func (c *Calcs) PercentMinMax(variableToWrite string, variable interface{}, params interface{}) (interface{}, error) {

	if reflect.TypeOf(variable).Kind() != reflect.Float64 {

		return 0, errors.New("Bad type for variable : " + variableToWrite)
	}

	pmap := params.(map[string]interface{})

	if value, ok := variable.(float64); ok {

		minif := c.Keyvaluemap[pmap["min"].(string)]
		maxif := c.Keyvaluemap[pmap["max"].(string)]

		if minif != nil && maxif != nil {

			min := minif.(float64)
			max := maxif.(float64)

			calc := (value - min) / ((max - min) / 100)

			ret := calc

			if calc < 0 {

				ret = 0

			} else if calc > 100 {

				ret = 100
			}

			return math.Floor(ret), nil
		}
	}

	if ret, ok := c.Keyvaluemap[variableToWrite]; ok {

		return ret, nil
	}

	return 0, nil
}

func (c *Calcs) IpxTC100(variableToWrite string, variable interface{}, params interface{}) (interface{}, error) {

	if reflect.TypeOf(variable).Kind() != reflect.Float64 {

		return 0, errors.New("Bad type for variable : " + variableToWrite)
	}

	value := variable.(float64)
	ret := ((value * 0.00323) - 0.25) / 0.028

	return math.Floor(ret), nil
}

func (c *Calcs) IpxTC4012(variableToWrite string, variable interface{}, params interface{}) (interface{}, error) {

	if reflect.TypeOf(variable).Kind() != reflect.Float64 {

		return 0, errors.New("Bad type for variable : " + variableToWrite)
	}

	value := variable.(float64)
	ret := (value * 0.323) - 50

	return math.Floor(ret), nil
}

func (c *Calcs) IpxTC5050(variableToWrite string, variable interface{}, params interface{}) (interface{}, error) {

	if reflect.TypeOf(variable).Kind() != reflect.Float64 {

		return 0, errors.New("Bad type for variable : " + variableToWrite)
	}

	value := variable.(float64)
	ret := ((value * 0.00323) - 1.63) / 0.0326

	return math.Floor(ret*10) / 10, nil
}

func (c *Calcs) IpxLights(variableToWrite string, variable interface{}, params interface{}) (interface{}, error) {

	if reflect.TypeOf(variable).Kind() != reflect.Float64 {

		return 0, errors.New("Bad type for variable : " + variableToWrite)
	}

	value := variable.(float64)
	ret := value * 0.09775

	return math.Floor(ret), nil
}

func (c *Calcs) IpxRH100(variableToWrite string, variable interface{}, params interface{}) (interface{}, error) {

	if reflect.TypeOf(variable).Kind() != reflect.Float64 {

		return 0, errors.New("Bad type for variable : " + variableToWrite)
	}

	value := variable.(float64)
	ret := (((value * 0.00323) / 3.3) - 0.1515) / 0.00636

	return math.Floor(ret*10) / 10, nil
}

package ds18b20

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

const basepath = "/sys/bus/w1/devices/"

func Run(c *chan map[string]interface{}, paramsIface interface{}) {

	for {

		m := make(map[string]interface{})

		files, _ := ioutil.ReadDir(basepath)
		for _, f := range files {

			sensor := f.Name()

			if strings.HasPrefix(sensor, "28-") {

				file, _ := ioutil.ReadFile(basepath + sensor + "/w1_slave")
				content := string(file)

				value, _ := strconv.ParseFloat(strings.Split(strings.Split(content, "t=")[1], "\n")[0], 64)
				fmt.Println(sensor+" :", value/1000)

				m[sensor] = value / 1000
			}
		}

		*c <- m

		time.Sleep(5 * time.Second)
	}
}

package si1145

import (
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const basepath = "/sys/bus/w1/devices/"

func Run(c *chan map[string]interface{}, paramsIface interface{}) {

	for {

		m := make(map[string]interface{})

		command := exec.Command("bash", "-c", "python /usr/local/bin/uvindex.py")
		out, _ := command.Output()

		lines := strings.Split(string(out), "\n")

		for _, line := range lines {

			if line != "" {
				val := strings.Split(line, ":")
				m[strings.TrimSpace(val[0])], _ = strconv.ParseFloat(strings.TrimSpace(val[1]), 64)
			}
		}

		*c <- m

		time.Sleep(5 * time.Second)
	}
}

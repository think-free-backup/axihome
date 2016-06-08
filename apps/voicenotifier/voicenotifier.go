package main

import (
	"bytes"
	"flag"
	"fmt"
	"net/http"
	"os/exec"
)

func main() {

	var lang = flag.String("lang", "fr-FR", "Local to be used")
	flag.Parse()

	http.HandleFunc("/say/", func(w http.ResponseWriter, r *http.Request) {

		sentence := r.URL.Path[len("/say/"):]

		fmt.Println("Running : /opt/svox-pico/say", "\""+sentence+"\"", "\""+*lang+"\"")

		cmd := exec.Command("/opt/svox-pico/say", "\""+sentence+"\"", *lang)
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr

		cmd.Start()
		err := cmd.Wait()

		if err != nil {

			w.Write([]byte(err.Error() + " " + stderr.String()))
		} else {

			w.Write([]byte("Ok"))
		}
	})

	http.ListenAndServe(":3333", nil)
}

package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/boltdb/bolt"
)

type server struct {
	db *bolt.DB
}

func main() {

	path := flag.String("path", "/etc/axihome/history/", "Path to database")
	flag.Parse()

	// Create server

	s := &server{}

	// Open db

	var err error
	s.db, err = bolt.Open(*path+"axihome-history.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer s.db.Close()

	// Server http

	http.HandleFunc("/favicon.ico", s.icon)
	http.HandleFunc("/set", s.set)
	http.HandleFunc("/get", s.get)
	err = http.ListenAndServe(":7777", nil)

	if err != nil {
		log.Println("ListenAndServe: ", err)
	}
}

func (s *server) icon(w http.ResponseWriter, r *http.Request) {
}

func (s *server) set(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()

	err := s.db.Update(func(tx *bolt.Tx) error {

		// Create bucket for key
		b, err := tx.CreateBucketIfNotExists([]byte(r.Form["K"][0]))
		if err != nil {
			return fmt.Errorf("Error creating bucket: %s", err)
		}

		// Save value to bucket
		err = b.Put([]byte(r.Form["TS"][0]), []byte(r.Form["V"][0]))
		if err != nil {

			log.Println("Can't save :", r.Form["K"][0])
		}
		return err
	})

	if err != nil {

		fmt.Fprintf(w, err.Error())
	} else {

		fmt.Fprintf(w, "Done")
	}

}

func (s *server) get(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()

	var buffer bytes.Buffer

	err := s.db.View(func(tx *bolt.Tx) error {

		c := tx.Bucket([]byte(r.Form["K"][0])).Cursor()

		// Our time range spans the 90's decade.
		min := []byte(r.Form["StartTS"][0])
		max := []byte(r.Form["StopTS"][0])

		buffer.WriteString("{")

		// Iterate over the 90's.
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {

			ok := false

			buffer.WriteString("\"")
			buffer.Write(k)
			buffer.WriteString("\" : ")

			vs := string(v)

			_, errb := strconv.ParseBool(vs)
			if errb == nil && !ok {

				buffer.Write(v)
				ok = true
			}

			_, errf := strconv.ParseFloat(vs, 64)
			if errf == nil && !ok {

				buffer.Write(v)
				ok = true
			}

			_, erri := strconv.ParseInt(vs, 10, 64)
			if erri == nil && !ok {

				buffer.Write(v)
				ok = true
			}

			if !ok {

				buffer.WriteString("\"")
				buffer.Write(v)
				buffer.WriteString("\"")
			}

			buffer.WriteString(",")
		}

		buffer.Truncate(buffer.Len() - 1)
		buffer.WriteString("}")

		return nil
	})

	if err != nil {

		fmt.Fprintf(w, err.Error())
	}

	fmt.Fprintf(w, buffer.String()) // send data to client side
}

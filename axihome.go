package main

import (
	"flag"

	"os"
	"os/signal"
	"syscall"

	"github.com/think-free/axihome/instancesmanager"
	"github.com/think-free/axihome/server"

	"github.com/think-free/log"
)

const version = "0.3"

func main() {

	debug := flag.Bool("d", false, "-d for debug info")

	dev := flag.Bool("dev", false, "-dev for devel mode")
	configPath := flag.String("c", ".", "-c to specify config path")
	flag.Parse()
	log.SetDebug(*debug)
	log.SetProcess("axihome")

	log.Println("")
	log.Println("Starting axihome server v" + version)
	log.Println("-----------------------------")
	log.Println("")

	if *debug {

		log.Println("Debug mode activated")
	}

	// Starting server
	server := axihomeserver.New(*configPath)
	go server.Run()

	// Starting instance manager
	im := instancesmanager.New(*dev)

	go im.Run()

	// Handle ctrl+c and exit signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)

	for {
		select {
		case _ = <-c:

			log.Println("\nClosing application")

			os.Exit(1)
		}
	}
}

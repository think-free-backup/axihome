package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/think-free/axihome/instancesmanager"
	"github.com/think-free/axihome/server"

	"github.com/think-free/log"
)

const version = "0.2"

func main() {

	debug := flag.Bool("d", false, "-d for debug info")
	flag.Parse()
	log.SetDebug(*debug)

	fmt.Println("")
	fmt.Println("Starting axihome server v" + version)
	fmt.Println("-----------------------------")
	fmt.Println("")

	if *debug {

		fmt.Println("Debug mode activated")
	}

	// Starting server
	server := axihomeserver.New()
	go server.Run()

	// Starting instance manager
	im := instancesmanager.New()
	go im.Run()

	// Handle ctrl+c and exit signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)

	for {
		select {
		case _ = <-c:
			fmt.Println("\nClosing application")
			os.Exit(1)
		}
	}
}
